package commands

import (
	"encoding/csv"
	"io"
	"os"
	"sort"
	"unicode"

	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/janusky/gitlab-api-client/utils"
)

var createProjectsCmd = &cobra.Command{
	Use:     "create-projects",
	Short:   "Create one or more gitlab projects",
	Aliases: []string{"create-project"},
	RunE:    doCreateProjects,
	Example: `  Creation of projects in *grupo* and with users *owner*

  gitlab-api-client create-project test1-app test1-db  \
    --group test1 \
    --owners user1,userN \
    --api-url https://gitlab.localhost/api/v4/ \	
    --private-token token \
    --trusted-certificates @certificates.pem`,
}

var (
	createProjectsGroup  string
	createProjectsOwners []string
	createProjectsFrom   string
)

func init() {
	rootCmd.AddCommand(createProjectsCmd)
	createProjectsCmd.Flags().StringVarP(&createProjectsGroup, "group", "g", "", "The group where projects will be created")
	createProjectsCmd.Flags().StringSliceVarP(&createProjectsOwners, "owners", "o", []string{}, "The owners of the new projects")
	createProjectsCmd.Flags().StringVarP(&createProjectsFrom, "from", "f", "", "CSV input file (group,project) to read groups and projects from")
}

func validateName(arg string) error {
	for _, char := range arg {
		if !unicode.In(char, unicode.Digit, unicode.Lower) && char != '-' {
			return errors.Errorf("character %q not allowed in name %q (must be all lower or '-')", char, arg)
		}
	}
	return nil
}

func doCreateProjects(cmd *cobra.Command, args []string) error {
	helper, err := gitlabAPI()
	if err != nil {
		return errors.Wrap(err, "creating gitlab helper")
	}
	if createProjectsFrom != "" {
		f, err := os.Open(createProjectsFrom)
		if err != nil {
			return errors.Wrapf(err, "opening input file %q", createProjectsFrom)
		}
		defer f.Close()
		groups := make(map[string][]string)
		r := csv.NewReader(f)
		for {
			rec, err := r.Read()
			if err == io.EOF {
				log.Info("csv end")
				break
			}
			log.Infof("read %+v", rec)
			if err != nil {
				return errors.Wrapf(err, "parsing csv data from file %q", createProjectsFrom)
			}
			g := rec[0]
			err = validateName(g)
			if err != nil {
				return errors.Wrap(err, "invalid group name")
			}
			p := rec[1]
			err = validateName(p)
			if err != nil {
				return errors.Wrap(err, "invalid project name")
			}
			ps := groups[g]
			i := sort.SearchStrings(ps, p)
			log.Infof("g=%s i=%d ps=%+v", g, i, ps)
			switch {
			case i >= len(ps):
				groups[g] = append(ps, p)
			case i < len(ps) && ps[i] != p:
				groups[g] = append(ps[:i], append([]string{p}, ps[i:]...)...)
			}
			log.Infof("groups=%+v", groups)
		}
		log.Infof("groups are %+v", groups)
		for group, projects := range groups {
			log.Infof("creating projects %v in group %v with owners %v", projects, group, createProjectsOwners)
			err := helper.CreateProjects(createProjectsOwners, group, projects)
			if err != nil {
				return errors.Wrapf(err, "creating projects %v in group %q", projects, group)
			}
		}
	}
	if len(args) > 0 {
		if createProjectsGroup == "" {
			return utils.NotImplementedError("user projects creation")
		}
		err := helper.CreateProjects(createProjectsOwners, createProjectsGroup, args)
		if err != nil {
			return errors.Wrap(err, "creating projects")
		}
	}
	return nil
}
