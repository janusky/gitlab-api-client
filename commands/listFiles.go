package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	gitlab "github.com/xanzy/go-gitlab"

	gitlabapi "github.com/janusky/gitlab-api-client/gitlab"
	"github.com/janusky/gitlab-api-client/utils"
)

// listFilesCmd represents the listFiles command
var listFilesCmd = &cobra.Command{
	Use:   "list-files",
	Short: "List projects files",
	RunE:  doListFiles,
	Example: `  List files in project

  gitlab-api-client gitlab list-files \
    --file 'main' \
    --project 'gitlab' \
    --api-url https://gitlab.localhost/api/v4/ \
    --private-token token \
    --trusted-certificates @certificates.pem`,
}

var (
	listFilesGroup      string
	listFilesProject    string
	listFilesBranch     string
	listFilesTag        string
	listFilesFile       string
	listFilesListFormat string
	listFilesCountLines bool
)

func init() {
	rootCmd.AddCommand(listFilesCmd)
	listFilesCmd.Flags().StringVarP(&listFilesGroup, "group", "g", "", "The pattern to match groups to be searched for files")
	listFilesCmd.Flags().StringVarP(&listFilesProject, "project", "p", "", "The pattern to match projects to be searched for files")
	listFilesCmd.Flags().StringVar(&listFilesBranch, "branch", "", "The pattern to match branches to be searched for files")
	listFilesCmd.Flags().StringVar(&listFilesTag, "tag", "", "The pattern to match tags to be searched for files")
	listFilesCmd.Flags().StringVarP(&listFilesFile, "file", "f", "", "The pattern to match file paths")
	listFilesCmd.Flags().StringVarP(&listFilesListFormat, "format", "F", "csv", "The listing format (json, csv or plain) - if is Debug does not print")
	listFilesCmd.Flags().BoolVarP(&listFilesCountLines, "count-lines", "l", false, "Calculate and report the line count of each listed file")
}

type file struct {
	Group   string `json:"group,omitempty"`
	Project string `json:"project,omitempty"`
	Branch  string `json:"branch,omitempty"`
	Tag     string `json:"tag,omitempty"`
	Path    string `json:"path,omitempty"`
	Lines   int    `json:"lines,omitempty"`
}

func (f *file) CSV() []string {
	ref := f.Branch
	if ref == "" {
		ref = f.Tag
	}
	return []string{f.Group, f.Project, ref, f.Path, strconv.Itoa(f.Lines)}
}

func (f *file) JSON() string {
	s, err := json.Marshal(f)
	if err != nil {
		log.Errorf("marshalling file %+v: %v", f, err)
	}
	return string(s)
}

func (f *file) Plain() string {
	ref := f.Branch
	if ref == "" {
		ref = f.Tag
	}
	return fmt.Sprintf("%s:%s:%s:%s:%d", f.Group, f.Project, ref, f.Path, f.Lines)
}

func countBytesLines(bs []byte) int {
	count := 0
	if len(bs) > 0 {
		for _, b := range bs {
			if b == '\n' {
				count++
			}
		}
		if bs[len(bs)-1] != '\n' {
			count++
		}
	}
	return count
}

func listFiles(
	helper *gitlabapi.GitlabApi,
	gitlabFileRegexp *regexp.Regexp,
	group *gitlab.Group,
	project *gitlab.Project,
	branch *gitlab.Branch,
	tag *gitlab.Tag) error {

	var ref string

	switch {
	case branch != nil:
		ref = branch.Name
	case tag != nil:
		ref = tag.Name
	default:
		return errors.New("missing branch or tag")
	}

	nodes, err := helper.ListTree(project, ref, "")
	if err != nil {
		return errors.Wrap(err, "listing ref tree")
	}

	for _, node := range nodes {

		if gitlabFileRegexp != nil && !gitlabFileRegexp.MatchString(node.Path) {
			log.Infof("skipped gitlab project '%s' file '%s' not matching '%s'", project.Name, node.Path, listFilesFile)
			continue
		}

		f := &file{
			Group:   group.Name,
			Project: project.Name,
			Path:    node.Path,
		}

		switch {
		case branch != nil:
			f.Branch = branch.Name
		case tag != nil:
			f.Tag = tag.Name
		}

		if listFilesCountLines {
			bs, _, err := helper.Client.Repositories.RawBlobContent(project.ID, node.TreeNode.ID)
			if err != nil {
				return errors.Wrapf(err, "getting file '%s' content", node.Path)
			}
			f.Lines = countBytesLines(bs)
		}

		switch listFilesListFormat {
		case "csv":
			utils.PrintCSV(f.CSV())
		case "plain":
			utils.PrintPlain(f.Plain())
		case "json":
			utils.PrintJSON(f.JSON())
		default:
			return errors.Errorf("unknown list format: %s", listFilesListFormat)
		}

	}

	return nil
}

func doListFiles(cmd *cobra.Command, args []string) error {
	var gitlabGroupRegexp, gitlabProjectRegexp, gitlabTagRegexp, gitlabBranchRegexp, gitlabFileRegexp *regexp.Regexp
	if listFilesGroup != "" {
		gitlabGroupRegexp = regexp.MustCompile(listFilesGroup)
	}
	if listFilesProject != "" {
		gitlabProjectRegexp = regexp.MustCompile(listFilesProject)
	}
	if listFilesTag != "" {
		gitlabTagRegexp = regexp.MustCompile(listFilesTag)
	}
	if listFilesBranch != "" {
		gitlabBranchRegexp = regexp.MustCompile(listFilesBranch)
	}
	if listFilesFile != "" {
		gitlabFileRegexp = regexp.MustCompile(listFilesFile)
	}
	helper, err := gitlabAPI()
	if err != nil {
		return errors.Wrap(err, "creating gitlab helper")
	}
	groups, err := helper.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return err
	}
	for _, group := range groups {
		if gitlabGroupRegexp != nil && !gitlabGroupRegexp.MatchString(group.Name) {
			log.Infof("skipped gitlab group '%s' not matching '%s'", group.Name, listFilesGroup)
			continue
		}
		projects, err := helper.ListGroupProjects(group)
		if err != nil {
			return errors.Wrap(err, "listing group projects")
		}
		for _, project := range projects {
			if gitlabProjectRegexp != nil && !gitlabProjectRegexp.MatchString(project.Name) {
				log.Infof("skipped gitlab project '%s' not matching '%s'", project.Name, listFilesProject)
				continue
			}

			if gitlabBranchRegexp != nil || gitlabTagRegexp == nil {
				branches, err := helper.ListProjectBranches(project)
				if err != nil {
					return errors.Wrap(err, "listing projects branches")
				}
				for _, branch := range branches {
					if gitlabBranchRegexp != nil && !gitlabBranchRegexp.MatchString(branch.Name) {
						log.Infof("skipped gitlab project '%s' branch '%s' not matching '%s'", project.Name, branch.Name, listFilesBranch)
						continue
					}
					err := listFiles(helper, gitlabFileRegexp, group, project, branch, nil)
					if err != nil {
						return errors.Wrap(err, "listing branch files")
					}
				}
			}

			if gitlabTagRegexp != nil || gitlabBranchRegexp == nil {
				tags, err := helper.ListProjectTags(project)
				if err != nil {
					return err
				}
				for _, tag := range tags {
					if gitlabTagRegexp != nil && !gitlabTagRegexp.MatchString(tag.Name) {
						log.Infof("skipped gitlab project '%s' tag '%s' not matching '%s'", project.Name, tag.Name, listFilesTag)
						continue
					}
					err := listFiles(helper, gitlabFileRegexp, group, project, nil, tag)
					if err != nil {
						return errors.Wrap(err, "listing tag files")
					}
				}
			}
		}
	}

	return nil
}
