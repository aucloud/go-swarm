package swarm

import (
	"fmt"
	"net"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"gitlab.mgt.aom.australiacloud.com.au/aom/golib/runcmd"
)

// Switcher is the interface that describes how to switch between Docker Nodes
// Implementations must implement the `Swithc()` method that should return an
// appropriate `runcmd.Runner` interface type for operating on Docker Nodes.
type Switcher interface {
	fmt.Stringer
	Switch(nodeAddr string) error
	Runner() runcmd.Runner
}

type nullSwitcher struct{}

func NewNullSwitcher() (Switcher, error)         { return &nullSwitcher{}, nil }
func (s *nullSwitcher) String() string           { return "" }
func (s *nullSwitcher) Switch(addr string) error { return nil }
func (s *nullSwitcher) Runner() runcmd.Runner    { return nil }

type localSwitcher struct {
	sync.RWMutex
	runner runcmd.Runner
}

// NewLocalSwitcher constructs a new Switcher that talks directly to a local
// Docker UNIX Socket on a single-node
func NewLocalSwitcher() (Switcher, error) {
	return &localSwitcher{}, nil
}

func (s *localSwitcher) String() string {
	return "local://"
}

func (s *localSwitcher) Runner() runcmd.Runner {
	s.RLock()
	defer s.RUnlock()
	return s.runner
}

func (s *localSwitcher) Switch(host string) error {
	runner, err := runcmd.NewLocalRunner()
	if err != nil {
		log.WithError(err).Error("error creating local runner")
		return fmt.Errorf("error creating local runner: %w", err)
	}

	s.Lock()
	s.runner = runner
	s.Unlock()

	return nil
}

type sshSwitcher struct {
	sync.RWMutex
	runner runcmd.Runner

	user string
	addr string
	key  string
}

// NewSSHSwitcher constructs a new Switcher that connect to remote Docker nodes'
// UNIX Sockets over SSH
func NewSSHSwitcher(user, addr, key string) (Switcher, error) {
	key = os.ExpandEnv(key)

	s := &sshSwitcher{
		user: user,
		addr: addr,
		key:  key,
	}

	if addr != "" {
		if err := s.Switch(addr); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *sshSwitcher) String() string {
	return fmt.Sprintf("ssh://%s@%s", s.user, s.addr)
}

func (s *sshSwitcher) Runner() runcmd.Runner {
	s.RLock()
	defer s.RUnlock()
	return s.runner
}

func (s *sshSwitcher) Switch(nodeAddr string) error {
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

	addr := fmt.Sprintf("%s:%s", nodeAddr, port)

	runner, err := runcmd.NewRemoteKeyAuthRunner(s.user, addr, s.key)
	if err != nil {
		return fmt.Errorf("error creating remote runner: %w", err)
	}

	s.Lock()
	s.addr = addr
	s.runner = runner
	s.Unlock()

	return nil
}
