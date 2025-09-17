package gowhoson

import (
	"context"
	"errors"

	"github.com/tai-ga/gowhoson/pkg/whoson"
	"github.com/urfave/cli/v3"
)

func cmdLogout(ctx context.Context, c *cli.Command) error {
	config := c.Root().Metadata["config"].(*whoson.ClientConfig)
	optOverwite(c, config)

	if !c.Args().Present() || c.Args().Len() != 1 {
		err := errors.New("arguments error, required 1 options")
		displayError(c.Root().ErrWriter, err)
		return err
	}

	ip := c.Args().Slice()[0]
	client, err := whoson.Dial(config.Mode, config.Server)
	if err != nil {
		displayError(c.Root().ErrWriter, err)
		return err
	}
	defer client.Quit()

	res, err := client.Logout(ip)
	if err != nil {
		displayError(c.Root().ErrWriter, err)
		return err
	}
	display(c.Root().Writer, res.String())

	return nil
}
