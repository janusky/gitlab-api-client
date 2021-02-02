package main

import (
	"fmt"

	"github.com/janusky/gitlab-api-client/commands"
)

var (
	version string
	commit  string
	date    string
)

func main() {
	commands.Execute(fmt.Sprintf("%s-%s (%s)", version, commit, date))
}
