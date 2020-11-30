package swarm

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strings"

	"github.com/prologic/jsonlines"
	log "github.com/sirupsen/logrus"

	"gitlab.mgt.aom.australiacloud.com.au/aom/golib/runcmd"
)

const (
	infoCommand  = "docker info --format='{{ json . }}'"
	nodesCommand = "docker node ls --format='{{ json . }}'"
	initCommand  = "docker swarm init --advertise-addr=%s --listen-addr=%s"
	joinCommand  = "docker swarm join --advertise-addr=%s --listen-addr=%s --token=%s %s:2377"
	tokenCommand = "docker swarm join-token -q %s"
)

type sshSwarmer struct {
	user string
	addr string
	key  string

	runner runcmd.Runner
}

func NewSSHSwarmer(user, addr, key string) (Swarmer, error) {
	key = os.ExpandEnv(key)

	s := &sshSwarmer{
		user: user,
		addr: addr,
		key:  key,
	}

	if addr != "" {
		if err := s.SwitchNode(addr); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *sshSwarmer) runCmd(cmd string, args ...string) (io.Reader, error) {
	if s.runner == nil {
		return nil, fmt.Errorf("error no runner configured")
	}

	log.Debugf("running cmd: %s", cmd)

	worker, err := s.runner.Command(cmd)
	if err != nil {
		return nil, fmt.Errorf("error creating ssh worker: %w", err)
	}

	stdout, err := worker.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating stdout pipe: %w", err)
	}
	stderr, err := worker.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating stdout pipe: %w", err)
	}

	if err := worker.Start(); err != nil {
		return nil, fmt.Errorf("error starting worker: %w", err)
	}

	if err := worker.Wait(); err != nil {
		r := io.MultiReader(stderr, stdout)
		out, _ := ioutil.ReadAll(r)
		log.Debugf("\n%s\n", string(out))
		return nil, fmt.Errorf("error running worker: %s", err)
	}

	return stdout, nil
}

func (s *sshSwarmer) ensureManager() error {
	node, err := s.GetInfo()
	if err != nil {
		return fmt.Errorf("error getting node info: %w", err)
	}
	if !node.IsManager() {
		for _, remoteManager := range node.Swarm.RemoteManagers {
			host, _, err := net.SplitHostPort(remoteManager.Addr)
			if err != nil {
				log.WithError(err).Warn("error parsing remote manager address (trying next manager): %w", err)
				continue
			}
			if err := s.SwitchNode(host); err != nil {
				log.WithError(err).Warn("error switch to remote manager (trying next manager): %w", err)
				continue
			}
			return nil
		}
		return fmt.Errorf("unable to connect to suitable manager")
	}

	return nil
}

func (s *sshSwarmer) joinSwarm(newNode VMNode, managerNode VMNode, token string) error {
	if err := s.SwitchNode(newNode.PublicAddress); err != nil {
		return fmt.Errorf("error switching nodes to %s: %w", newNode.PublicAddress, err)
	}

	cmd := fmt.Sprintf(
		joinCommand,
		newNode.PrivateAddress,
		newNode.PrivateAddress,
		token,
		managerNode.PrivateAddress,
	)
	_, err := s.runCmd(cmd)
	if err != nil {
		return fmt.Errorf("error running join command: %w", err)
	}

	return nil
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

	log.Debugf("Using SSH Addr: %s", addr)
	log.Debugf("Using SSH Key:  %s", s.key)
	log.Debugf("Using SSH User: %s", s.user)

	runner, err := runcmd.NewRemoteKeyAuthRunner(s.user, addr, s.key)
	if err != nil {
		return fmt.Errorf("error creating runner: %w", err)
	}

	s.addr = addr
	s.runner = runner

	return nil
}

func (s *sshSwarmer) GetInfo() (NodeInfo, error) {
	var node NodeInfo

	cmd := infoCommand
	stdout, err := s.runCmd(cmd)
	if err != nil {
		return NodeInfo{}, fmt.Errorf("error running info command: %w", err)
	}

	data, err := ioutil.ReadAll(stdout)
	if err != nil {
		return NodeInfo{}, fmt.Errorf("error reading info command output: %w", err)
	}

	if err := json.Unmarshal(data, &node); err != nil {
		return NodeInfo{}, fmt.Errorf("error parsing json data: %s", err)
	}

	return node, nil
}

func (s *sshSwarmer) GetManagers() ([]NodeInfo, error) {
	node, err := s.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("error getting node info: %w", err)
	}

	var managers []NodeInfo
	for _, remoteManager := range node.Swarm.RemoteManagers {
		host, _, err := net.SplitHostPort(remoteManager.Addr)
		if err != nil {
			return nil, fmt.Errorf("error parsing remote manager address: %w", err)
		}
		if err := s.SwitchNode(host); err != nil {
			return nil, fmt.Errorf("error switching nodes to %s: %w", host, err)
		}
		node, err := s.GetInfo()
		if err != nil {
			return nil, fmt.Errorf("error getting manager node info: %w", err)
		}
		managers = append(managers, node)
	}

	return managers, nil
}

func (s *sshSwarmer) GetNodes() ([]NodeStatus, error) {
	if err := s.ensureManager(); err != nil {
		return nil, fmt.Errorf("error connecting to manager node: %w", err)
	}

	cmd := nodesCommand
	stdout, err := s.runCmd(cmd)
	if err != nil {
		return nil, fmt.Errorf("error running nodes command: %w", err)
	}

	var nodes []NodeStatus

	if err := jsonlines.Decode(stdout, &nodes); err != nil {
		return nil, fmt.Errorf("error parsing json data: %s", err)
	}

	return nodes, nil
}

func (s *sshSwarmer) CreateSwarm(vms VMNodes) error {
	managers := vms.FilterByTag("role", "manager")
	if !(len(managers) == 3 || len(managers) == 5) {
		return fmt.Errorf("error expected 3 or 5 managers but got %d", len(managers))
	}

	workers := vms.FilterByTag("role", "worker")

	// Pick a random manager out of the candidates
	randomIndex := rand.Intn(len(managers))
	manager := managers[randomIndex]

	if err := s.SwitchNode(manager.PublicAddress); err != nil {
		return fmt.Errorf("error switching to a manager node: %w", err)
	}

	node, err := s.GetInfo()
	if err != nil {
		return fmt.Errorf("error getting node info: %w", err)
	}

	clusterID := node.Swarm.Cluster.ID

	if clusterID != "" {
		return fmt.Errorf("error swarm cluster with id %s already exists", clusterID)
	}

	cmd := fmt.Sprintf(initCommand, manager.PrivateAddress, manager.PrivateAddress)
	if _, err := s.runCmd(cmd); err != nil {
		return fmt.Errorf("error running init command: %w", err)
	}

	// Refresh node and get new Swarm Clsuter ID
	node, err = s.GetInfo()
	if err != nil {
		return fmt.Errorf("error refreshing node info: %w", err)
	}
	clusterID = node.Swarm.Cluster.ID

	managerToken, err := s.JoinToken("manager")
	if err != nil {
		return fmt.Errorf("error getting manager join token: %w", err)
	}

	workerToken, err := s.JoinToken("worker")
	if err != nil {
		return fmt.Errorf("error getting worker join token: %w", err)
	}

	// Join remaining managers
	for _, newManager := range managers {
		// Skip the leader we just created the swarm with
		if newManager.PublicAddress == manager.PublicAddress {
			continue
		}

		if err := s.joinSwarm(newManager, manager, managerToken); err != nil {
			return fmt.Errorf(
				"error joining manager %s to %s on swarm clsuter %s: %w",
				newManager.PublicAddress, manager.PublicAddress,
				clusterID, err,
			)
		}
	}

	// Join workers
	for _, worker := range workers {
		if err := s.joinSwarm(worker, manager, workerToken); err != nil {
			return fmt.Errorf(
				"error joining worker %s to %s on swarm clsuter %s: %w",
				worker.PublicAddress, manager.PublicAddress,
				clusterID, err,
			)
		}
	}

	if err := s.SwitchNode(manager.PublicAddress); err != nil {
		return fmt.Errorf("error switching to manager node: %w", err)
	}

	return nil
}

func (s *sshSwarmer) JoinToken(tokenType string) (string, error) {
	cmd := fmt.Sprintf(tokenCommand, tokenType)
	stdout, err := s.runCmd(cmd)
	if err != nil {
		return "", fmt.Errorf("error running token command: %w", err)
	}

	data, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", fmt.Errorf("error reading stdout: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}
