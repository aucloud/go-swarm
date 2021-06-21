package swarm

import (
	"fmt"

	"gitlab.mgt.aom.australiacloud.com.au/aom/golib/runcmd"
)

type localSwarmer struct {
	*baseSwarmer
}

// NewLocalSwarmer constructs a new Swarmer that talks directly to a local
// Docker UNIX Socket on a single-node
func NewLocalSwarmer() (Swarmer, error) {
	s := &localSwarmer{&baseSwarmer{}}

	if err := s.SwitchNode(""); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *localSwarmer) String() string {
	return "local"
}

func (s *localSwarmer) SwitchNode(host string) error {
	runner, err := runcmd.NewLocalRunner()
	if err != nil {
		return fmt.Errorf("error creating runner: %w", err)
	}

	s.runner = runner

	return nil
}
