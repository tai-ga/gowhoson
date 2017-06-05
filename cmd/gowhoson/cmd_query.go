package gowhoson

import (
	"errors"

	"github.com/tai-ga/gowhoson/whoson"
	"github.com/urfave/cli"
)

func cmdQuery(c *cli.Context) error {
	config := c.App.Metadata["config"].(*ClientConfig)
	optOverwite(c, config)

	if !c.Args().Present() || len(c.Args()) != 1 {
		err := errors.New("arguments error, required 1 options")
		displayError(c.App.ErrWriter, err)
		return err
	}

	ip := c.Args()[0]
	client, err := whoson.Dial(config.Mode, config.Server)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return err
	}
	defer client.Quit()

	res, err := client.Query(ip)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return err
	}
	display(c.App.Writer, res.String())

	return nil
}
