package commands

import (
	"github.com/apex/log"
	"github.com/janusky/gitlab-api-client/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var removeProjectsCmd = &cobra.Command{
	Use:     "remove-projects",
	Aliases: []string{"remove-project"},
	Short:   "Remove one or more gitlab projects",
	RunE:    doRemoveProjects,
	Example: `  Remove one or more gitlab projects

  gitlab-api-client remove-project test1-db \
	--group test1 \
	--api-url https://gitlab.localhost/api/v4/ \
    --private-token token \
    --trusted-certificates @certificates.pem`,
}

var (
	removeProjectsGroup string
)

func init() {
	rootCmd.AddCommand(removeProjectsCmd)
	removeProjectsCmd.Flags().StringVarP(&removeProjectsGroup, "group", "g", "", "The group where projects will be removed")
}

func doRemoveProjects(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.Errorf("no project names specified")
	}
	gitlabAPI, err := gitlabAPI()
	if err != nil {
		return errors.Wrap(err, "creating gitlab api")
	}
	if removeProjectsGroup != "" {
		g, _, err := gitlabAPI.Client.Groups.GetGroup(removeProjectsGroup)
		if err != nil {
			return errors.Wrapf(err, "getting group %q", removeProjectsGroup)
		}
		groupProjects, err := gitlabAPI.ListGroupProjects(g)
		if err != nil {
			return errors.Wrapf(err, "listing group %q projects", removeProjectsGroup)
		}
		for _, projectName := range args {
			found := false
			for _, project := range groupProjects {
				if project.Name == projectName {
					_, err := gitlabAPI.Client.Projects.DeleteProject(project.ID)
					if err != nil {
						return errors.Wrapf(err, "removing project %q (%d) from group %q", projectName, project.ID, removeProjectsGroup)
					}
					log.Infof("removed project %q (%d) from group %q", projectName, project.ID, removeProjectsGroup)
					found = true
					break
				}
			}
			if !found {
				return errors.Errorf("project %q not found in group %q", projectName, removeProjectsGroup)
			}
		}
		return nil
	}
	return utils.NotImplementedError("user projects removal")
}
