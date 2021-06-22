package main

import (
	"github.com/spf13/cobra"

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
		internal.Status(manager, args)
	},
}
