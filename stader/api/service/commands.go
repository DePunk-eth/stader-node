package service

import (
	"github.com/urfave/cli"

	"github.com/stader-labs/stader-node/shared/utils/api"
	cliutils "github.com/stader-labs/stader-node/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool deposit queue",
		Subcommands: []cli.Command{
			{
				Name:      "terminate-data-folder",
				Aliases:   []string{"t"},
				Usage:     "Deletes the data folder including the wallet file, password file, and all validator keys - don't use this unless you have a very good reason to do it (such as switching from Prater to Mainnet)",
				UsageText: "stader-cli api service terminate-data-folder",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(terminateDataFolder(c))
					return nil

				},
			},

			{
				Name:      "get-client-status",
				Aliases:   []string{"g"},
				Usage:     "Gets the status of the configured Execution and Beacon clients",
				UsageText: "stader-cli api service get-client-status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getClientStatus(c))
					return nil

				},
			},
		},
	})
}
