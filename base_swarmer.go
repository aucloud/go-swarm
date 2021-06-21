package swarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"strings"

	"github.com/prologic/jsonlines"
	log "github.com/sirupsen/logrus"

	"gitlab.mgt.aom.australiacloud.com.au/aom/golib/runcmd"
)

type baseSwarmer struct {
	runner runcmd.Runner
}

func (s *baseSwarmer) runCmd(cmd string, args ...string) (io.Reader, error) {
	if s.runner == nil {
		return nil, fmt.Errorf("error no runner configured")
	}

	log.WithField("args", args).Debugf("running cmd on %s: %s", s.String(), cmd)

	worker, err := s.runner.Command(cmd)
	if err != nil {
		return nil, fmt.Errorf("error creating worker: %w", err)
	}

	buf := &bytes.Buffer{}
	worker.SetStdout(buf)
	worker.SetStderr(buf)

	if err := worker.Start(); err != nil {
		return nil, fmt.Errorf("error starting worker: %w", err)
	}

	if err := worker.Wait(); err != nil {
		log.WithError(err).WithField("out", buf.String()).Error("error running worker")
		return nil, fmt.Errorf("error running worker: %s", err)
	}

	return buf, nil
}

func (s *baseSwarmer) ensureManager() error {
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

func (s *baseSwarmer) joinSwarm(newNode VMNode, managerNode VMNode, token string) error {
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

func (s *baseSwarmer) labelNode(node VMNode) error {
	if err := s.SwitchNode(node.PublicAddress); err != nil {
		return fmt.Errorf("error switching nodes to %s: %w", node.PublicAddress, err)
	}

	info, err := s.GetInfo()
	if err != nil {
		return fmt.Errorf("error getting node info from: %w", err)
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

	cmd := fmt.Sprintf(
		updateCommand,
		strings.Join(labelOptions, " "),
		info.Swarm.NodeID,
	)
	_, err = s.runCmd(cmd)
	if err != nil {
		return fmt.Errorf("error running update command: %w", err)
	}

	return nil
}

func (s *baseSwarmer) String() string {
	return ""
}

func (s *baseSwarmer) SwitchNode(host string) error {
	return fmt.Errorf("error SwitchNode() not implemented on %v", s)
}

func (s *baseSwarmer) GetInfo() (NodeInfo, error) {
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

func (s *baseSwarmer) GetManagers() ([]NodeInfo, error) {
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

func (s *baseSwarmer) GetNodes() ([]NodeStatus, error) {
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

func (s *baseSwarmer) CreateSwarm(vms VMNodes) error {
	managers := vms.FilterByTag(RoleTag, ManagerRole)
	if !(len(managers) == 3 || len(managers) == 5) {
		return fmt.Errorf("error expected 3 or 5 managers but got %d", len(managers))
	}

	workers := vms.FilterByTag(RoleTag, WorkerRole)

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
	if err := s.labelNode(manager); err != nil {
		return fmt.Errorf("error labelling worker: %w", err)
	}

	// Refresh node and get new Swarm Clsuter ID
	node, err = s.GetInfo()
	if err != nil {
		return fmt.Errorf("error refreshing node info: %w", err)
	}
	clusterID = node.Swarm.Cluster.ID

	managerToken, err := s.JoinToken(managerToken)
	if err != nil {
		return fmt.Errorf("error getting manager join token: %w", err)
	}

	workerToken, err := s.JoinToken(workerToken)
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
		if err := s.labelNode(newManager); err != nil {
			return fmt.Errorf("error labelling manager: %w", err)
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
		if err := s.labelNode(worker); err != nil {
			return fmt.Errorf("error labelling worker: %w", err)
		}
	}

	if err := s.SwitchNode(manager.PublicAddress); err != nil {
		return fmt.Errorf("error switching to manager node: %w", err)
	}

	return nil
}

func (s *baseSwarmer) JoinToken(tokenType string) (string, error) {
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
