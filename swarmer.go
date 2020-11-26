package swarm

type Swarmer interface {
	SwitchNode(addr string) error
	GetInfo() (NodeInfo, error)
	GetManagers() ([]NodeInfo, error)
	GetNodes() ([]NodeStatus, error)
	CreateSwarm(vms VMNodes) error
	JoinToken(tokenType string) (string, error)
}
