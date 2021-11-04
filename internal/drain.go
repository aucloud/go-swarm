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
	"strings"

	"github.com/aucloud/go-swarm"
)

func Drain(m *swarm.Manager, args []string) int {
	if err := m.DrainNodes(args); err != nil {
		fmt.Fprintf(os.Stderr, "error draining nodes: %s\n", err)
		return StatusError
	}

	fmt.Fprintf(os.Stdout, "Nodes %s successfully drained\n", strings.Join(args, ","))

	return Status(m, nil)
}
