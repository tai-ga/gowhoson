package main

import (
	"os"

	"github.com/tai-ga/gowhoson/internal/gowhoson"
)

var (
	gVersion   string
	gGitcommit string
)

func main() {
	gowhoson.AppVersions = gowhoson.NewVersions(gVersion, gGitcommit)
	os.Exit(gowhoson.Run())
}
