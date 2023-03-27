package cmd

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func HomeCommand() *cobra.Command {
	return &cobra.Command{
		Use: "home",
		// TODO: Check usage output - currently not describing case for setting new value
		Short: "Read/Change the home directory. No home directory is set when using the tool standalone.",
		Long:  "Read/Change the home directory. Given no argument, it outputs the string being used as the home path. Given one argument, the home directory is written to the config file. No home directory is set when using the tool standalone.",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			if clientCtx.HomeDir == "" {
				cmd.Println("No home directory set.")
			} else {
				switch len(args) {
				case 0: // if no arguments are given, print the home directory
					cmd.Println(clientCtx.HomeDir)
				case 1: // if one argument is given, set the home directory
					newHome := args[0]
					newConfigPath := filepath.Join(newHome, "config.toml")
					if _, err := os.Stat(newConfigPath); os.IsNotExist(err) {
						if err := config.CreateNewConfigAtPath(newHome, clientCtx.ChainID); err != nil {
							return err
						}
					}
					if err := config.WriteHomeDirToFile(clientCtx.HomeFilePath, newHome); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
}
