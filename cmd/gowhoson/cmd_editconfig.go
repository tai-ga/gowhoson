package gowhoson

import (
	"path/filepath"

	"github.com/urfave/cli"
)

func CmdEditConfig(c *cli.Context) error {
	file := filepath.Join(GetClientConfigDir(), CLIENT_CONFIG)
	e := NewFileEdit(file)
	e.Edit()

	return nil
}
