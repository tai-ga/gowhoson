package main

import (
	"os"

	"github.com/tai-ga/gowhoson/cmd/gowhoson"
)

var (
	gVersion   string
	gGitcommit string
	gGoversion string
)

func main() {
	gowhoson.AppVersions = gowhoson.NewVersions(gVersion, gGitcommit, gGoversion)
	os.Exit(gowhoson.Run())
}
