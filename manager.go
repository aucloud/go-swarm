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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mills.io/jsonlines"

	"github.com/aucloud/go-runcmd"
)

const (
	infoCommand        = `docker info --format "{{ json . }}"`
	nodesCommand       = `docker node ls --format "{{ json . }}"`
	tasksCommand       = `docker node ps --format "{{ json .}}" %s`
	initCommand        = `docker swarm init --advertise-addr %s --listen-addr %s`
	joinCommand        = `docker swarm join --advertise-addr %s --listen-addr %s --token %s %s:2377`
	tokenCommand       = `docker swarm join-token -q %s`
	updateCommand      = `docker node update %s %s`
	setAvailability    = `--availability %s`
	labelAdd           = `--label-add %s`
	availabilityDrain  = `drain`
	availabilityActive = `active`

	managerToken = "manager"
	workerToken  = "worker"

	drainTimeout = time.Minute * 10 // 10 minutes
)

const (
	DefaultTimeout = time.Minute * 5
)

type Config struct {
	Timeout time.Duration
}

func NewDefaultConfig() *Config {
	return &Config{
		Timeout: DefaultTimeout,
	}
}

// Manager manages all operations of a Docker Swarm cluster with flexible
// Switcher implementations that permit talking to Docker Nodes over different
// types of transport (e.g: local or remote).
type Manager struct {
	config   *Config
	switcher Switcher
}

type Option func(*Config) error

func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		cfg.Timeout = timeout
		return nil
	}
}

// NewManager constructs a new Manager type with the provider Switcher
func NewManager(switcher Switcher, options ...Option) (*Manager, error) {
	m := &Manager{switcher: switcher, config: NewDefaultConfig()}

	for _, opt := range options {
		if err := opt(m.config); err != nil {
			log.WithError(err).Error("error configuring swarm manager")
			return nil, err
		}
	}

	return m, nil
}

// Switcher returns the current Switcher for the manager being used
func (m *Manager) Switcher() Switcher {
	return m.switcher
}

// Runner returns the current Runner for the current Switcher being used
func (m *Manager) Runner() runcmd.Runner {
	return m.Switcher().Runner()
}

// SwitchNode switches to a new node given by nodeAddr to perform operations on
func (m *Manager) SwitchNode(nodeAddr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
	defer cancel()
	if err := m.Switcher().Switch(ctx, nodeAddr); err != nil {
		log.WithError(err).Errorf("error switching to node %s", nodeAddr)
		return fmt.Errorf("error switching to node %s: %s", nodeAddr, err)
	}

	return nil
}

// SwitchNodeVia switches to a new node given by nodeAddr by jumping through
// the current node as a "bastion" host to perform operations on the node.
func (m *Manager) SwitchNodeVia(nodeAddr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
	defer cancel()
	if err := m.Switcher().SwitchVia(ctx, nodeAddr); err != nil {
		log.WithError(err).Errorf("error switching to node %s via %s", nodeAddr, m.Switcher())
		return fmt.Errorf("error switching to node %s via %s: %s", nodeAddr, m.Switcher(), err)
	}

	return nil
}

func (m *Manager) runCmd(cmd string, args ...string) (io.Reader, error) {
	if m.Runner() == nil {
		return nil, fmt.Errorf("error no runner configured")
	}

	log.WithField("args", args).Debugf("running cmd on %s: %s", m.switcher.String(), cmd)

	worker, err := m.Runner().Command(cmd)
	if err != nil {
		return nil, fmt.Errorf("error creating worker: %w", err)
	}

	stdout := &bytes.Buffer{}
	worker.SetStdout(stdout)

	stderr := &bytes.Buffer{}
	worker.SetStderr(stderr)

	if err := worker.Start(); err != nil {
		return nil, fmt.Errorf("error starting worker: %w", err)
	}

	if err := worker.Wait(); err != nil {
		log.WithError(err).
			WithField("stdout", string(stdout.String())).
			WithField("stderr", string(stderr.String())).
			Error("error running worker")
		return nil, fmt.Errorf(
			"error running worker: %w (stderr=%q stdout=%q)",
			err, stderr.String(), stdout.String(),
		)
	}

	return stdout, nil
}

func (m *Manager) ensureManager() error {
	node, err := m.GetInfo()
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
			if err := m.SwitchNodeVia(host); err != nil {
				log.WithError(err).Warn("error switching to remote manager (trying next manager): %w", err)
				continue
			}
			return nil
		}
		return fmt.Errorf("unable to connect to suitable manager")
	}

	return nil
}

func (m *Manager) joinSwarm(newNode VMNode, managerNode VMNode, token string) error {
	if err := m.SwitchNode(newNode.PublicAddress); err != nil {
		return fmt.Errorf("error switching nodes to %s: %w", newNode.PublicAddress, err)
	}

	cmd := fmt.Sprintf(
		joinCommand,
		newNode.PrivateAddress,
		newNode.PrivateAddress,
		token,
		managerNode.PrivateAddress,
	)
	_, err := m.runCmd(cmd)
	if err != nil {
		return fmt.Errorf("error running join command: %w", err)
	}

	return nil
}

func (m *Manager) LabelNode(node VMNode) error {
	if err := m.SwitchNode(node.PublicAddress); err != nil {
		return fmt.Errorf("error switching nodes to %s: %w", node, err)
	}

	info, err := m.GetInfo()
	if err != nil {
		return fmt.Errorf("error getting node info: %w", err)
	}

	labelOptions := []string{}

	labels, err := ParseLabels(node.GetTag(LabelsTag))
	if err != nil {
		log.WithError(err).Error("error parsing labels")
		return fmt.Errorf("error parsing labels: %w", err)
	}

	if labels == nil || len(labels) == 0 {
		// No labels, nothing to do.
		return nil
	}

	for key, values := range labels {
		label := key
		if values != nil || len(values) > 0 {
			label += fmt.Sprintf("=%s", strings.Join(values, ","))
		}
		labelOptions = append(labelOptions, fmt.Sprintf(labelAdd, label))
	}

	if err := m.ensureManager(); err != nil {
		return fmt.Errorf("error connecting to manager node: %w", err)
	}

	cmd := fmt.Sprintf(
		updateCommand,
		strings.Join(labelOptions, " "),
		info.Swarm.NodeID,
	)
	_, err = m.runCmd(cmd)
	if err != nil {
		return fmt.Errorf("error running update command: %w", err)
	}

	return nil
}

// GetInfo returns information about the current node
func (m *Manager) GetInfo() (NodeInfo, error) {
	var node NodeInfo

	cmd := infoCommand
	out, err := m.runCmd(cmd)
	if err != nil {
		return NodeInfo{}, fmt.Errorf("error running info command: %w", err)
	}

	data, err := ioutil.ReadAll(out)
	if err != nil {
		return NodeInfo{}, fmt.Errorf("error reading info command output: %w", err)
	}

	if err := json.Unmarshal(data, &node); err != nil {
		return NodeInfo{}, fmt.Errorf("error parsing json data: %s", err)
	}

	return node, nil
}

// GetManagers returns a list of manager nodes and their information
func (m *Manager) GetManagers() ([]NodeInfo, error) {
	node, err := m.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("error getting node info: %w", err)
	}

	var managers []NodeInfo
	for _, remoteManager := range node.Swarm.RemoteManagers {
		host, _, err := net.SplitHostPort(remoteManager.Addr)
		if err != nil {
			return nil, fmt.Errorf("error parsing remote manager address: %w", err)
		}
		if err := m.SwitchNode(host); err != nil {
			return nil, fmt.Errorf("error switching nodes to %s: %w", host, err)
		}
		node, err := m.GetInfo()
		if err != nil {
			return nil, fmt.Errorf("error getting manager node info: %w", err)
		}
		managers = append(managers, node)
	}

	return managers, nil
}

// GetNodes returns all nodes in the cluster
func (m *Manager) GetNodes() ([]NodeStatus, error) {
	if err := m.ensureManager(); err != nil {
		return nil, fmt.Errorf("error connecting to manager node: %w", err)
	}

	cmd := nodesCommand
	stdout, err := m.runCmd(cmd)
	if err != nil {
		return nil, fmt.Errorf("error running nodes command: %w", err)
	}

	var nodes []NodeStatus

	if err := jsonlines.Decode(stdout, &nodes); err != nil {
		return nil, fmt.Errorf("error parsing json data: %s", err)
	}

	return nodes, nil
}

// CreateSwarm creates a new Docker Swarm cluster given a set of nodes
func (m *Manager) CreateSwarm(vms VMNodes, force bool) error {
	managers := vms.FilterByTag(RoleTag, ManagerRole)

	if force {
		log.Warnf("skipping manager validation and forcing creation of cluster with %d managers", len(managers))
	} else {
		if !(len(managers) == 3 || len(managers) == 5) {
			return fmt.Errorf("error expected 3 or 5 managers but got %d", len(managers))
		}
	}

	workers := vms.FilterByTag(RoleTag, WorkerRole)

	// Pick a random manager out of the candidates
	randomIndex := rand.Intn(len(managers))
	manager := managers[randomIndex]

	if err := m.SwitchNode(manager.PublicAddress); err != nil {
		return fmt.Errorf("error switching to a manager node: %w", err)
	}

	node, err := m.GetInfo()
	if err != nil {
		return fmt.Errorf("error getting node info: %w", err)
	}

	clusterID := node.Swarm.Cluster.ID

	if clusterID != "" {
		return fmt.Errorf("error swarm cluster with id %s already exists", clusterID)
	}

	cmd := fmt.Sprintf(initCommand, manager.PrivateAddress, manager.PrivateAddress)
	if _, err := m.runCmd(cmd); err != nil {
		return fmt.Errorf("error running init command: %w", err)
	}

	// Refresh node and get new Swarm Clsuter ID
	node, err = m.GetInfo()
	if err != nil {
		return fmt.Errorf("error refreshing node info: %w", err)
	}
	clusterID = node.Swarm.Cluster.ID

	managerToken, err := m.JoinToken(managerToken)
	if err != nil {
		return fmt.Errorf("error getting manager join token: %w", err)
	}

	workerToken, err := m.JoinToken(workerToken)
	if err != nil {
		return fmt.Errorf("error getting worker join token: %w", err)
	}

	// Join remaining managers
	for _, newManager := range managers {
		// Skip the leader we just created the swarm with
		if newManager.PublicAddress == manager.PublicAddress {
			continue
		}

		if err := m.joinSwarm(newManager, manager, managerToken); err != nil {
			return fmt.Errorf(
				"error joining manager %s to %s on swarm clsuter %s: %w",
				newManager.PublicAddress, manager.PublicAddress,
				clusterID, err,
			)
		}
	}

	// Join workers
	for _, worker := range workers {
		if err := m.joinSwarm(worker, manager, workerToken); err != nil {
			return fmt.Errorf(
				"error joining worker %s to %s on swarm clsuter %s: %w",
				worker.PublicAddress, manager.PublicAddress,
				clusterID, err,
			)
		}
	}

	if err := m.SwitchNode(manager.PublicAddress); err != nil {
		return fmt.Errorf("error switching to manager node: %w", err)
	}

	// Label nodes
	for _, vm := range vms {
		log.Debugf("labellig node %s with tags=%v", vm.Tags)
		if err := m.LabelNode(vm); err != nil {
			return fmt.Errorf("error labelling node %s: %w", vm, err)
		}
	}

	if err := m.SwitchNode(manager.PublicAddress); err != nil {
		return fmt.Errorf("error switching to manager node: %w", err)
	}

	return nil
}

// UpdateSwarm updates an existing Docker Swarm cluster by adding any
// missing manager or worker nodes that aren't already part of the cluster
func (m *Manager) UpdateSwarm(vms VMNodes) error {
	currentNodes := make(map[string]bool)
	desiredNodes := make(map[string]bool)

	nodes, err := m.GetNodes()
	if err != nil {
		return fmt.Errorf("error getting current nodes: %w", err)
	}
	for _, node := range nodes {
		currentNodes[node.Hostname] = true
	}
	for _, vm := range vms {
		desiredNodes[vm.Hostname] = true
	}

	var newNodes VMNodes

	for _, vm := range vms {
		if _, ok := currentNodes[vm.Hostname]; !ok {
			newNodes = append(newNodes, vm)
		}
	}

	var nodesToDrain []string

	for node := range currentNodes {
		if _, ok := desiredNodes[node]; !ok {
			nodesToDrain = append(nodesToDrain, node)
		}
	}

	managers := vms.FilterByTag(RoleTag, ManagerRole)
	if !(len(managers) == 3 || len(managers) == 5) {
		return fmt.Errorf("error expected 3 or 5 managers but got %d", len(managers))
	}

	// Pick a random manager out of the candidates
	randomIndex := rand.Intn(len(managers))
	manager := managers[randomIndex]

	newWorkers := newNodes.FilterByTag(RoleTag, WorkerRole)
	newManagers := newNodes.FilterByTag(RoleTag, ManagerRole)

	if err := m.ensureManager(); err != nil {
		return fmt.Errorf("error connecting to manager node: %w", err)
	}

	node, err := m.GetInfo()
	if err != nil {
		return fmt.Errorf("error getting node info: %w", err)
	}

	clusterID := node.Swarm.Cluster.ID

	if clusterID == "" {
		return fmt.Errorf("error no swarm cluster found")
	}

	managerToken, err := m.JoinToken(managerToken)
	if err != nil {
		return fmt.Errorf("error getting manager join token: %w", err)
	}

	workerToken, err := m.JoinToken(workerToken)
	if err != nil {
		return fmt.Errorf("error getting worker join token: %w", err)
	}

	// Join new managers
	for _, newManager := range newManagers {
		if err := m.joinSwarm(newManager, manager, managerToken); err != nil {
			return fmt.Errorf(
				"error joining manager %s to %s on swarm clsuter %s: %w",
				newManager.PublicAddress, manager.PublicAddress,
				clusterID, err,
			)
		}
		if err := m.LabelNode(newManager); err != nil {
			return fmt.Errorf("error labelling manager: %w", err)
		}
	}

	// Join new workers
	for _, newWorker := range newWorkers {
		if err := m.joinSwarm(newWorker, manager, workerToken); err != nil {
			return fmt.Errorf(
				"error joining worker %s to %s on swarm clsuter %s: %w",
				newWorker.PublicAddress, manager.PublicAddress,
				clusterID, err,
			)
		}
		if err := m.LabelNode(newWorker); err != nil {
			return fmt.Errorf("error labelling worker: %w", err)
		}
	}

	// Remove old nodes
	if err := m.DrainNodes(nodesToDrain); err != nil {
		log.WithError(err).Error("error ddraining old nodes")
		return fmt.Errorf("error draining old nodes: %w", err)
	}

	if err := m.SwitchNode(manager.PublicAddress); err != nil {
		return fmt.Errorf("error switching to manager node: %w", err)
	}

	return nil
}

func (m *Manager) getTasks(node string) (Tasks, error) {
	cmd := fmt.Sprintf(tasksCommand, node)
	stdout, err := m.runCmd(cmd)
	if err != nil {
		return nil, fmt.Errorf("error running tasks command: %w", err)
	}

	var tasks Tasks

	if err := jsonlines.Decode(stdout, &tasks); err != nil {
		return nil, fmt.Errorf("error parsing json data: %s", err)
	}

	return tasks, nil
}

func (m *Manager) drainNode(node string) error {
	startedAt := time.Now()

	cmd := fmt.Sprintf(updateCommand, fmt.Sprintf(setAvailability, availabilityDrain), node)
	_, err := m.runCmd(cmd)
	if err != nil {
		return fmt.Errorf("error running update command: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), drainTimeout)
	defer cancel()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startedAt)

			tasks, err := m.getTasks(node)
			if err != nil {
				log.WithError(err).Warnf("error getting tasks from node %s (retrying)", node)
				continue
			}

			if tasks.AllShutdown() {
				log.Infof("Successfully drained %s after %s", node, elapsed)
				return nil
			}

			log.Infof("Still waiting for %s to drain after %s ...", node, elapsed)
		case <-ctx.Done():
			elapsed := time.Since(startedAt)
			log.Errorf("timed out waiting for %s to drain after %s", node, elapsed)
			return fmt.Errorf("error timed out waiting for %s to drain after %s", node, elapsed)
		}
	}

	// Unreachable
}

// DrainNodes drains one or more nodes from an existing Docker Swarm cluster
// and blocks until there are no more tasks running on thoese nodes.
func (m *Manager) DrainNodes(nodes []string) error {
	if err := m.ensureManager(); err != nil {
		return fmt.Errorf("error connecting to manager node: %w", err)
	}

	for _, node := range nodes {
		if err := m.drainNode(node); err != nil {
			log.WithError(err).Errorf("error draining node: %s", node)
			return fmt.Errorf("error draining node %s: %w", node, err)
		}
	}

	return nil
}

// JoinToken retrieves the current join token for the given type
// "manager" or "worker" from any of the managers in the cluster
func (m *Manager) JoinToken(tokenType string) (string, error) {
	cmd := fmt.Sprintf(tokenCommand, tokenType)
	stdout, err := m.runCmd(cmd)
	if err != nil {
		return "", fmt.Errorf("error running token command: %w", err)
	}

	data, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", fmt.Errorf("error reading stdout: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}
