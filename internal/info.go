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

package internal

import (
	"fmt"
	"os"

	"github.com/aucloud/go-swarm"
)

func Info(m *swarm.Manager, args []string) int {
	node, err := m.GetInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting node info: %s\n", err)
		return StatusError
	}

	managers, err := m.GetManagers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting managers: %s\n", err)
		return StatusError
	}

	fmt.Fprintf(os.Stdout, "Cluster ID: %s\n", node.Swarm.Cluster.ID)
	fmt.Fprintf(os.Stdout, "Nodes: %d\n", node.Swarm.Nodes)
	fmt.Fprintf(os.Stdout, "Managers: %d\n", node.Swarm.Managers)
	fmt.Fprintf(os.Stdout, "Workers: %d\n", (node.Swarm.Nodes - node.Swarm.Managers))

	fmt.Fprintf(os.Stdout, "Managers:\n")
	for _, manager := range managers {
		fmt.Fprintf(os.Stdout, "  %s\n", manager.Name)
	}

	return StatusOK
}
