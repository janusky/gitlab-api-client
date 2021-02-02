package commands

import (
	"fmt"
	"strconv"

	"github.com/janusky/gitlab-api-client/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	deployKeyID        int
	deployKeyProjectID int
	deployKeyDisabled  bool
)

var deployKeyCmd = &cobra.Command{
	Use:     "deploy-key",
	Short:   "Gitlab projects deploy key",
	Aliases: []string{"deploy-key-projects"},
	RunE:    doDeployKeyProjects,
	Example: `  Operating gitlab projects deploy key (disabled|enabled)
	
  # Enabled deploy-key
  gitlab-api-client deploy-key \
    --api-url https://gitlab.localhost/api/v4/ \
    --private-token token \
    --trusted-certificates @certificates.pem

  # Disabled deploy-key
  gitlab-api-client deploy-key --disabled \
    --api-url https://gitlab.localhost/api/v4/ \
    --private-token token \
    --trusted-certificates @certificates.pem`,
}

func init() {
	rootCmd.AddCommand(deployKeyCmd)
	deployKeyCmd.Flags().BoolVarP(&deployKeyDisabled, "disabled", "d", false, "The project to disabled or enabled deploy key")
	deployKeyCmd.Flags().IntVarP(&deployKeyProjectID, "project-id", "q", 0, "The project to be searched for keys")
	deployKeyCmd.Flags().IntVarP(&deployKeyID, "key-id", "k", 0, "Id deploy key")
}

func doDeployKeyProjects(cmd *cobra.Command, args []string) error {
	gitlabAPI, err := gitlabAPI()
	if err != nil {
		return errors.Wrap(err, "creating gitlab api")
	}
	projects, err := gitlabAPI.ListAllGroupsProjects()
	if err != nil {
		return errors.Wrap(err, "listing projects")
	}
	deployKey, err := gitlabAPI.GetDeployKey(deployKeyProjectID, deployKeyID)
	if err != nil {
		return err
	}
	countEdit := 0
	countNotEdit := 0
	for _, project := range projects {
		if project.Public {
			if deployKeyDisabled {
				err = gitlabAPI.DisabledDeployKey(deployKey, project.ID)
			} else {
				err = gitlabAPI.EnableDeployKey(deployKey, project.ID)
			}
			if err != nil {
				utils.PrintCSV([]string{project.Name, fmt.Sprintf("Fail %v", err)})
				countNotEdit++
			} else {
				utils.PrintCSV([]string{project.Name, "ok"})
				countEdit++
			}
		} else {
			countNotEdit++
		}
	}
	utils.PrintCSV([]string{"total", strconv.Itoa(len(projects))})
	utils.PrintCSV([]string{"edit", strconv.Itoa(countEdit)})
	utils.PrintCSV([]string{"notEdit", strconv.Itoa(countNotEdit)})
	return nil
}
