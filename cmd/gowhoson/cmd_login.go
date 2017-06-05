package gowhoson

import (
	"errors"

	"github.com/tai-ga/gowhoson/whoson"
	"github.com/urfave/cli"
)

func cmdLogin(c *cli.Context) error {
	config := c.App.Metadata["config"].(*ClientConfig)
	optOverwite(c, config)

	if !c.Args().Present() || len(c.Args()) != 2 {
		err := errors.New("arguments error, required 2 options")
		displayError(c.App.ErrWriter, err)
		return err
	}

	ip := c.Args()[0]
	data := c.Args()[1]
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
