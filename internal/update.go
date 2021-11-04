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
	"io"
	"os"

	"github.com/aucloud/go-swarm"
)

func Update(m *swarm.Manager, args []string) int {
	var (
		f   io.ReadCloser
		err error
	)

	clusterFile := args[0]

	if clusterFile == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(clusterFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading Clusterfile: %s\n", err)
			return StatusError
		}
	}

	cf, err := swarm.ReadClusterfile(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing Clusterfile: %s\n", err)
		return StatusError
	}

	// TODO: Validate no existing cluster exists in this cf.Nodes (VMNodes)
	// TODO: Modify Validate to take VMNodes as input.
	if err := cf.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "error validating Clusterfile: %s\n", err)
		return StatusError
	}

	if err := m.UpdateSwarm(cf.Nodes); err != nil {
		fmt.Fprintf(os.Stderr, "error updating swarm cluster: %s\n", err)
		return StatusError
	}

	node, err := m.GetInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting node info: %s\n", err)
		return StatusError
	}

	fmt.Fprintf(os.Stdout, "Swarm Cluster successfully updated with id: %s\n", node.Swarm.Cluster.ID)

	return Status(m, nil)
}
