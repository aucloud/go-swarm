package main

import (
	"github.com/spf13/cobra"

	"gitlab.mgt.aom.australiacloud.com.au/aom/swarm/internal"
)

func init() {
	RootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{},
	Short:   "Updates an existing Swarm Cluster",
	Long: `This command uses a Clusterfile that describes the number of
and types of nodes that should exist in the Swarm Cluster. If there are
nodes that are missing from the cluster that should be new managers or
workers, they are added. Any that should be removed are drained and
removed from the cluster gracefully.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		internal.Update(manager, args)
	},
}
