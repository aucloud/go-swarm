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

const (
	// DefaultSSHUser is the default SSH Username when executing remote commands
	DefaultSSHUser = "rancher"

	// DefaultSSHKey is the default SSH Kry when executing remote commands
	DefaultSSHKey = "$HOME/.ssh/id_rsa"

	// DefaultSockPath is the default path to the Docker API's UNIX Socket
	DefaultSockPath = "/var/run/docker.sock"

	// MinSwarmClusterNodes is the minimum number of  nodes to form a swam cluster
	MinSwarmClusterNodes = 1
)
