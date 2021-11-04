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

	"github.com/aucloud/go-swarm/internal"
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
		internal.Info(manager, args)
	},
}
