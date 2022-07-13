/*
	go-swarm is a Go library and ccommand-line tool for managing the creation
	and maintenance of Docker Swarm cluster.

    Copyright (C) 2021 Sovereign Cloud Australia Pty Ltd

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package swarm

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/aucloud/go-runcmd"
	log "github.com/sirupsen/logrus"
)

// Switcher is the interface that describes how to switch between Docker Nodes
// Implementations must implement the `Swithc()` method that should return an
// appropriate `runcmd.Runner` interface type for operating on Docker Nodes.
type Switcher interface {
	fmt.Stringer
	Switch(ctx context.Context, nodeAddr string) error
	SwitchVia(ctx context.Context, nodeAddr string) error
	Runner() runcmd.Runner
}

type nullSwitcher struct{}

func NewNullSwitcher() (Switcher, error)                                 { return &nullSwitcher{}, nil }
func (s *nullSwitcher) String() string                                   { return "" }
func (s *nullSwitcher) Switch(ctx context.Context, addr string) error    { return nil }
func (s *nullSwitcher) SwitchVia(ctx context.Context, addr string) error { return nil }
func (s *nullSwitcher) Runner() runcmd.Runner                            { return nil }

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

func (s *localSwitcher) Switch(ctx context.Context, host string) error {
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

func (s *localSwitcher) SwitchVia(ctx context.Context, host string) error {
	return s.Switch(ctx, host)
}

type SSHRunnerAuth interface {
	createRunner(context.Context, string, string) (*runcmd.Remote, error)
	createRunnerVia(context.Context, string, string, string) (*runcmd.Remote, error)
}

type keyRunnerAuth struct {
	key string
}

func (s keyRunnerAuth) createRunner(ctx context.Context, user, addr string) (*runcmd.Remote, error) {
	key := os.ExpandEnv(s.key)

	return runcmd.NewRemoteKeyAuthRunner(ctx, user, addr, key)
}

func (s keyRunnerAuth) createRunnerVia(ctx context.Context, user, addr, jumphost string) (*runcmd.Remote, error) {
	key := os.ExpandEnv(s.key)

	return runcmd.NewRemoteKeyAuthRunnerViaJumphost(ctx, user, addr, jumphost, key)
}

type agentRunnerAuth struct {
	agentSock string
}

func (s *agentRunnerAuth) createRunner(ctx context.Context, user, addr string) (*runcmd.Remote, error) {
	return runcmd.NewRemoteAgentAuthRunner(ctx, user, addr, s.agentSock)
}

func (s *agentRunnerAuth) createRunnerVia(ctx context.Context, user, addr, jumphost string) (*runcmd.Remote, error) {
	return nil, fmt.Errorf("agent authentication is not supported for jump hosts")
}

type sshSwitcher struct {
	sync.RWMutex
	runner runcmd.Runner

	user string
	addr string
	auth SSHRunnerAuth
	jump string
}

func NewKeySSHRunnerAuth(key string) SSHRunnerAuth {
	return &keyRunnerAuth{key}
}

func NewAgentSSHRunnerAuth(agentSock string) SSHRunnerAuth {
	return &agentRunnerAuth{agentSock}
}

// NewSSHSwitcher constructs a new Switcher that connect to remote Docker nodes'
// UNIX Sockets over SSH
func NewSSHSwitcher(user, addr string, auth SSHRunnerAuth, timeout time.Duration) (Switcher, error) {
	s := &sshSwitcher{
		user: user,
		addr: addr,
		auth: auth,
	}

	if addr != "" {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := s.Switch(ctx, addr); err != nil {
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

func (s *sshSwitcher) Switch(ctx context.Context, nodeAddr string) error {
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

	runner, err := s.auth.createRunner(ctx, s.user, addr)
	if err != nil {
		return fmt.Errorf("error creating remote runner: %w", err)
	}

	s.Lock()
	s.addr = addr
	s.runner = runner
	s.Unlock()

	return nil
}

func (s *sshSwitcher) SwitchVia(ctx context.Context, nodeAddr string) error {
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

	runner, err := s.auth.createRunnerVia(ctx, s.user, addr, s.addr)
	if err != nil {
		return fmt.Errorf("error creating remote runner: %w", err)
	}

	s.Lock()
	s.jump = s.addr
	s.addr = addr
	s.runner = runner
	s.Unlock()

	return nil
}
