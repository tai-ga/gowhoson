// +build tools

package tools

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/fzipp/gocyclo"
	_ "github.com/gordonklaus/ineffassign"
	_ "github.com/trawler/goviz"
	_ "golang.org/x/lint/golint"
)
