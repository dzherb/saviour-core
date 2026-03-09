package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"saviour/internal/command"
)

func main() {
	cmd := &cli.Command{
		Name:  "saviour",
		Usage: "manage your backups with ease",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:      "config",
				Aliases:   []string{"c"},
				Required:  true,
				TakesFile: true,
				Usage:     "path to a config file (one or more)",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "run the saviour backend",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return command.Run(
						ctx,
						cmd.StringSlice("config"),
					)
				},
			},
			{
				Name:  "create-admin",
				Usage: "create an admin user",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "username",
						Aliases: []string{"u"},
						Value:   "admin",
					},
					&cli.StringFlag{
						Name:     "password",
						Required: true,
						Aliases:  []string{"p"},
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return command.CreateAdmin(
						ctx,
						cmd.StringSlice("config"),
						command.CreateAdminParams{
							Username: cmd.String("username"),
							Password: cmd.String("password"),
						},
					)
				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
