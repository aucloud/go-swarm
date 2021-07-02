package internal

import (
	"fmt"
	"os"
	"strings"

	"gitlab.mgt.aom.australiacloud.com.au/aom/swarm"
)

func Drain(m *swarm.Manager, args []string) int {
	if err := m.DrainNodes(args); err != nil {
		fmt.Fprintf(os.Stderr, "error draining nodes: %s\n", err)
		return StatusError
	}

	fmt.Fprintf(os.Stdout, "Nodes %s successfully drained\n", strings.Join(args, ","))

	return Status(m, nil)
}
