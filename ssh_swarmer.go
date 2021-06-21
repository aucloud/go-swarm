package swarm

import (
	"fmt"
	"net"
	"os"

	"gitlab.mgt.aom.australiacloud.com.au/aom/golib/runcmd"
)

type sshSwarmer struct {
	*baseSwarmer

	user string
	addr string
	key  string
}

// NewSSHSwarmer constructs a new Swarmer that connect to remote Docker nodes'
// UNIX Sockets over SSH
func NewSSHSwarmer(user, addr, key string) (Swarmer, error) {
	key = os.ExpandEnv(key)

	s := &sshSwarmer{
		baseSwarmer: &baseSwarmer{},
		user:        user,
		addr:        addr,
		key:         key,
	}

	if addr != "" {
		if err := s.SwitchNode(addr); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *sshSwarmer) String() string {
	return fmt.Sprintf("ssh://%s@%s", s.user, s.addr)
}

func (s *sshSwarmer) SwitchNode(host string) error {
	_, port, err := net.SplitHostPort(s.addr)
	if err != nil {
		if addrError, ok := err.(*net.AddrError); ok && addrError.Err == "missing port in address" {
			port = "22"
		} else {
			return fmt.Errorf("error parsing addr: %w", err)
		}
	}
	if port == "" {
		port = "22"
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	runner, err := runcmd.NewRemoteKeyAuthRunner(s.user, addr, s.key)
	if err != nil {
		return fmt.Errorf("error creating runner: %w", err)
	}

	s.addr = addr
	s.runner = runner

	return nil
}
