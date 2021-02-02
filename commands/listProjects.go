package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/janusky/gitlab-api-client/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	gitlab "github.com/xanzy/go-gitlab"
)

var (
	listProject       string
	listProjectGroup  string
	listProjectFormat string
)

var listProjectsCmd = &cobra.Command{
	Use:   "list-projects",
	Short: "List gitlab projects",
	RunE:  doListProjects,
	Example: `  List user projects

  gitlab-api-client list-projects \
    --api-url https://gitlab.localhost/api/v4/ \
    --private-token token \
    --trusted-certificates @certificates.pem`,
}

func init() {
	rootCmd.AddCommand(listProjectsCmd)
	listProjectsCmd.Flags().StringVarP(&listProjectGroup, "group", "g", "", "The pattern to match groups")
	listProjectsCmd.Flags().StringVarP(&listProject, "project", "p", "", "The pattern to match projects")
	listProjectsCmd.Flags().StringVarP(&listProjectFormat, "format", "F", "csv", "The listing format (json, csv or plain) - if is Debug does not print")
}

func doListProjects(cmd *cobra.Command, args []string) error {
	gitlabAPI, err := gitlabAPI()
	if err != nil {
		return errors.Wrap(err, "creating gitlab api")
	}
	groups, err := gitlabAPI.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return errors.Wrap(err, "listing groups")
	}
	var gitlabGroupRegexp, gitlabProjectRegexp *regexp.Regexp
	if listProjectGroup != "" && strings.Trim(listProjectGroup, " ") != "" {
		gitlabGroupRegexp = regexp.MustCompile(listProjectGroup)
	}
	if listProject != "" && strings.Trim(listProject, " ") != "" {
		gitlabProjectRegexp = regexp.MustCompile(listProject)
	}
	for _, group := range groups {
		if gitlabGroupRegexp != nil && !gitlabGroupRegexp.MatchString(group.Name) {
			log.Debugf("skipped gitlab group '%s' not matching '%s'", group.Name, listProjectGroup)
			continue
		}
		projects, err := gitlabAPI.ListGroupProjects(group)
		if err != nil {
			return errors.Wrap(err, "listing group projects")
		}
		for _, project := range projects {
			if gitlabProjectRegexp != nil && !gitlabProjectRegexp.MatchString(project.Name) {
				log.Debugf("skipped gitlab project '%s' not matching '%s'", project.Name, listProject)
				continue
			}
			visibility := fmt.Sprintf("%v", project.Visibility)
			createdAtFmt := fmt.Sprintf("%d/%02d/%02d", project.CreatedAt.Year(), project.CreatedAt.Month(), project.CreatedAt.Day())

			switch listProjectFormat {
			case "csv":
				utils.PrintCSV([]string{visibility, project.PathWithNamespace, strconv.Itoa(project.ID), createdAtFmt})
			case "plain":
				utils.PrintPlain(fmt.Sprintf("%s:%s:%d:%s", visibility, project.PathWithNamespace, project.ID, createdAtFmt))
			case "json":
				utils.PrintJSON(&map[string]string{"visibility": visibility, "name": project.PathWithNamespace, "id": strconv.Itoa(project.ID), "created": createdAtFmt})
			default:
				return errors.Errorf("unknown list format: %s", listProjectFormat)
			}
		}
	}
	return nil
}
