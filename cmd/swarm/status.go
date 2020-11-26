package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gitlab.mgt.aom.australiacloud.com.au/aom/swarm"
	"gitlab.mgt.aom.australiacloud.com.au/aom/swarm/internal"
)

func init() {
	RootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{},
	Short:   "Retrieve and display Swarm Cluster Status",
	Long: `This command retrives and display information about the Swarm Clsuter
status of all nodes participating int he warm including which ndoes are mangers,
workers and who the current leader is.`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		user := viper.GetString("ssh-user")
		addr := viper.GetString("ssh-addr")
		key := viper.GetString("ssh-key")
		swarmer, err := swarm.NewSSHSwarmer(user, addr, key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating swarmer: %s\n", err)
			os.Exit(-1)
		}

		internal.Status(swarmer, args)
	},
}
