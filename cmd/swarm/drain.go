package main

import (
	"github.com/spf13/cobra"

	"github.com/aucloud/go-swarm/internal"
)

func init() {
	RootCmd.AddCommand(drainCmd)
}

var drainCmd = &cobra.Command{
	Use:     "drain",
	Aliases: []string{},
	Short:   "Drains one or more nodes in an existing Swarm Cluster",
	Long: `This command drains one or more nodes from an existing Swarm Cluster
and waits for tasks to be shutdown on those nodes before returning.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		internal.Drain(manager, args)
	},
}
