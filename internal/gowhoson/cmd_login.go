package gowhoson

import (
	"errors"

	"github.com/tai-ga/gowhoson/pkg/whoson"
	"github.com/urfave/cli/v2"
)

func cmdLogin(c *cli.Context) error {
	config := c.App.Metadata["config"].(*whoson.ClientConfig)
	optOverwite(c, config)

	if !c.Args().Present() || c.Args().Len() != 2 {
		err := errors.New("arguments error, required 2 options")
		displayError(c.App.ErrWriter, err)
		return err
	}

	ip := c.Args().Slice()[0]
	data := c.Args().Slice()[1]
	client, err := whoson.Dial(config.Mode, config.Server)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return err
	}
	defer client.Quit()

	res, err := client.Login(ip, data)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return err
	}
	display(c.App.Writer, res.String())

	return nil
}
