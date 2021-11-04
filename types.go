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

package swarm

import (
	"strings"
)

type ClusterInfo struct {
	ID        string
	CreatedAt string
}

type RemoteManager struct {
	NodeID string
	Addr   string
}

type SwarmInfo struct {
	NodeID           string
	NodeAddr         string
	LocalNodeState   string
	ControlAvailable bool

	Nodes          int
	Managers       int
	RemoteManagers []RemoteManager

	Cluster ClusterInfo
}

type NodeInfo struct {
	ID     string
	Name   string
	Labels []string

	OSType          string
	OSVersion       string
	KernelVersion   string
	OperatingSystem string

	NCPU     int
	MemTotal int64

	ServerVersion string

	Swarm SwarmInfo
}

func (node NodeInfo) IsManager() bool {
	return node.Swarm.ControlAvailable
}

type NodeStatus struct {
	ID            string
	Hostname      string
	EngineVersion string
	Availability  string
	ManagerStatus string
	Status        string
}

type Nodes []NodeStatus

type TaskStatus struct {
	ID           string
	Name         string
	Image        string
	Error        string
	Node         string
	Ports        string
	CurrentState string
	DesiredState string
}

func (t TaskStatus) Shutdown() bool {
	return strings.HasPrefix(strings.ToLower(t.CurrentState), "shutdown")
}

type Tasks []TaskStatus

func (ts Tasks) AllShutdown() bool {
	for _, t := range ts {
		if !t.Shutdown() {
			return false
		}
	}
	return true
}
