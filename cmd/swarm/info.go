package main

import (
	"github.com/spf13/cobra"

	"gitlab.mgt.aom.australiacloud.com.au/aom/swarm/internal"
)

func init() {
	RootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{},
	Short:   "Retrieve and display Swarm Cluster Information",
	Long: `This command retrives and display information about the Swarm Clsuter
such as the number of worker nodes, manager nodes and cluster size.`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		internal.Info(swarmer, args)
	},
}
