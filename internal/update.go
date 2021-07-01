package internal

import (
	"fmt"
	"io"
	"os"

	"gitlab.mgt.aom.australiacloud.com.au/aom/swarm"
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
