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
	listGroup       string
	listGroupFormat string
)

var listGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "List gitlab groups",
	RunE:  doListGroups,
	Example: `  List user grupos

  gitlab-api-client list-groups \
    --api-url https://gitlab.localhost/api/v4/ \
    --private-token token \
    --trusted-certificates @certificates.pem`,
}

func init() {
	rootCmd.AddCommand(listGroupsCmd)
	listGroupsCmd.Flags().StringVarP(&listGroup, "group", "g", "", "The pattern to match groups")
	listGroupsCmd.Flags().StringVarP(&listGroupFormat, "format", "F", "csv", "The listing format (json, csv or plain) - if is Debug does not print")
}

func doListGroups(cmd *cobra.Command, args []string) error {
	gitlabAPI, err := gitlabAPI()
	if err != nil {
		return errors.Wrap(err, "creating gitlab api")
	}
	groups, err := gitlabAPI.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return errors.Wrap(err, "listing groups")
	}
	var gitlabGroupRegexp *regexp.Regexp
	if listGroup != "" && strings.Trim(listGroup, " ") != "" {
		gitlabGroupRegexp = regexp.MustCompile(listGroup)
	}
	for _, group := range groups {
		if gitlabGroupRegexp != nil && !gitlabGroupRegexp.MatchString(group.Name) {
			log.Debugf("skipped gitlab group '%s' not matching '%s'", group.Name, listGroup)
			continue
		}
		switch listGroupFormat {
		case "csv":
			utils.PrintCSV([]string{strconv.Itoa(group.ID), group.Name})
		case "plain":
			utils.PrintPlain(fmt.Sprintf("%d:%s", group.ID, group.Name))
		case "json":
			utils.PrintJSON(&map[string]string{"id": strconv.Itoa(group.ID), "name": group.Name})
		default:
			return errors.Errorf("unknown list format: %s", listGroupFormat)
		}
	}
	return nil
}
