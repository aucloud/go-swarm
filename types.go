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
