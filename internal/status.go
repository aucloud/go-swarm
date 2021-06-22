package internal

import (
	"fmt"
	"os"

	"gitlab.mgt.aom.australiacloud.com.au/aom/swarm"
)

func Status(m *swarm.Manager, args []string) int {
	nodes, err := m.GetNodes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting nodes: %s\n", err)
		return StatusError
	}

	for _, node := range nodes {
		fmt.Fprintf(
			os.Stdout, "%s %s %s %s %s %s\n",
			node.ID,
			node.Hostname,
			node.Status,
			node.Availability,
			node.ManagerStatus,
			node.EngineVersion,
		)
	}

	return StatusOK
}
