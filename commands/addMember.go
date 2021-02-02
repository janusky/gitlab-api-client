package commands

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/apex/log"
	"github.com/janusky/gitlab-api-client/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	gitlab "github.com/xanzy/go-gitlab"
)

var (
	username    string
	accessLevel int

	accessLevelOptions map[int]gitlab.AccessLevelValue

	addMemberProject string
)

var addMemberCmd = &cobra.Command{
	Use:   "add-member",
	Short: "Operating Gitlab Users add to project",
	Long: `TODO
	
  https://docs.gitlab.com/ee/api/members.html
  The access levels are defined in the Gitlab::Access module. Currently, these levels are recognized:

    No access (0)
    Minimal access (5) (Introduced in GitLab 13.5.)
    Guest (10)
    Reporter (20)
    Developer (30)
    Maintainer (40)
    Owner (50) - Only valid to set for groups`,
	Aliases: []string{"member-add"},
	RunE:    doAddMember,
}

func init() {
	rootCmd.AddCommand(addMemberCmd)
	addMemberCmd.Flags().StringVarP(&addMemberProject, "project", "p", "", "The pattern to match projects")
	addMemberCmd.Flags().StringVarP(&username, "username", "U", "", "The username to associate")
	addMemberCmd.Flags().IntVarP(&accessLevel, "access", "L", int(gitlab.ReporterPermissions), "The access level in project (default ReporterPermissions)")

	accessLevelOptions = make(map[int]gitlab.AccessLevelValue)
	accessLevelOptions[int(gitlab.ReporterPermissions)] = gitlab.ReporterPermissions
	accessLevelOptions[int(gitlab.MasterPermissions)] = gitlab.MasterPermissions
	accessLevelOptions[int(gitlab.OwnerPermissions)] = gitlab.OwnerPermissions
}

func doAddMember(cmd *cobra.Command, args []string) error {
	gitlabAPI, err := gitlabAPI()
	if err != nil {
		return errors.Wrap(err, "creating gitlab api")
	}
	accessLevelValue, ok := accessLevelOptions[accessLevel]
	if !ok {
		return errors.Wrap(err, "access level value is not allowed")
	}
	userAdd, err := gitlabAPI.GetUser(username)
	if err != nil {
		return errors.Wrap(err, "getting user")
	}
	var gitlabProjectRegexp *regexp.Regexp
	if addMemberProject != "" {
		gitlabProjectRegexp = regexp.MustCompile(addMemberProject)
	}
	projects, err := gitlabAPI.ListAllGroupsProjects()
	if err != nil {
		return errors.Wrap(err, "listing projects")
	}
	countEdit := 0
	countNotEdit := 0
	for _, project := range projects {
		if gitlabProjectRegexp != nil && !gitlabProjectRegexp.MatchString(project.Name) {
			log.Infof("skipped gitlab project '%s' not matching '%s'", project.Name, addMemberProject)
			continue
		}
		if project.Visibility == "public" {
			err := gitlabAPI.AddMembers(project, &accessLevelValue, userAdd)
			if err != nil {
				utils.PrintCSV([]string{project.Name, fmt.Sprintf("Fail %v", err)})
				countNotEdit++
			} else {
				utils.PrintCSV([]string{project.Name, "ok"})
				countEdit++
			}
		} else {
			log.Infof("skipped gitlab project '%s' is not public (%v)", project.Name, project.Visibility)
			countNotEdit++
		}
	}
	utils.PrintCSV([]string{"total", strconv.Itoa(len(projects))})
	utils.PrintCSV([]string{"edit", strconv.Itoa(countEdit)})
	utils.PrintCSV([]string{"notEdit", strconv.Itoa(countNotEdit)})
	return nil
}
