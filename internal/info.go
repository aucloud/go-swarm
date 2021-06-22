package internal

import (
	"fmt"
	"os"

	"gitlab.mgt.aom.australiacloud.com.au/aom/swarm"
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
