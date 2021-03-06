/*
	go-swarm is a Go library and ccommand-line tool for managing the creation
	and maintenance of Docker Swarm cluster.

    Copyright (C) 2021 Sovereign Cloud Australia Pty Ltd

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aucloud/go-swarm/internal"
)

func init() {
	createCmd.Flags().BoolP(
		"force-single-manager-cluster", "f", false,
		"Force creation of single-mnager-node clusters",
	)
	viper.BindPFlag("force-single-manager-cluster", createCmd.Flags().Lookup("force"))
	viper.SetDefault("force-single-manager-cluster", false)

	RootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{},
	Short:   "Creates a new Swarm Cluster",
	Long: `This command uses a Clusterfile that describes a new VM Cluster
of nodes to create a new Docker Swarm Cluster. The Clusterfile is expected to
have information about the region, enviornment, cluaster and a list of nodes
along with their public and private ip address. Each node must also have a set
of labels that are used to assign nodes as managers and others as workers.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force := viper.GetBool("force-single-manager-cluster")
		internal.Create(manager, args, force)
	},
}
