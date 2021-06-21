package swarm

import "fmt"

const (
	infoCommand   = `docker info --format "{{ json . }}"`
	nodesCommand  = `docker node ls --format "{{ json . }}"`
	initCommand   = `docker swarm init --advertise-addr "%s" --listen-addr "%s"`
	joinCommand   = `docker swarm join --advertise-addr "%s" --listen-addr "%s" --token "%s" "%s:2377"`
	tokenCommand  = `docker swarm join-token -q "%s"`
	updateCommand = `docker node update "%s" "%s"`
	labelAdd      = `--label-add %s`

	managerToken = "manager"
	workerToken  = "worker"
)

// Swarmer is the interfac that encapsualtes how to run operations against
// Docker Nodes whether they are local or remote and wraps common functionality
// to deal with the creation and management of a Docker Swarm Cluster.
//
// Implementations must at least implement SwitchNode() and can inherit
// baseSwarmer to have default functionality for all implemtnations.
type Swarmer interface {
	fmt.Stringer

	SwitchNode(addr string) error
	GetInfo() (NodeInfo, error)
	GetManagers() ([]NodeInfo, error)
	GetNodes() ([]NodeStatus, error)
	CreateSwarm(vms VMNodes) error
	JoinToken(tokenType string) (string, error)
}
