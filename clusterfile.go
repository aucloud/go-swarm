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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

const (
	// RoleTag is the tag (Custom Attribute in vSphere)
	// for tagging VM(s) with either "manager" or "worker"
	// This is used to assign Docker Swarm roles to VM(s).
	RoleTag = "role"

	// ManagerRole denotates a Docker Swarm role of "manager"
	ManagerRole = "manager"

	// WorkerRole denotates a Docker Swarm role of "worker"
	WorkerRole = "worker"

	// LabelsTag is the tag (Custom Attribute in vSphere)
	// for freeform labels applied to VM(s) in the form
	// `key1=value1&key2=value2&key3&key4`
	// (This uses the URL Query String format).
	LabelsTag = "labels"
)

// VMNode represents a single VM Node and at a bare minimum contains the
// node's hostname, private and public ip addresses as well as a list of tags
// used to label the nodes for different purposes such as Manager ndoes.
type VMNode struct {
	Hostname       string            `json:"hostname"`
	PublicAddress  string            `json:"public_address"`
	PrivateAddress string            `json:"private_address"`
	Tags           map[string]string `json:"tags"`
}

func (vm VMNode) Stirng() string {
	return fmt.Sprintf(
		"VMNode{Hostname: %q, PublicAddress: %q}",
		vm.Hostname, vm.PublicAddress,
	)
}

func (vm VMNode) GetTag(name string) string {
	return vm.Tags[name]
}

func (vm VMNode) HasTag(name, value string) bool {
	actual, ok := vm.Tags[name]
	return ok && actual == value
}

type VMNodes []VMNode

func (vms VMNodes) FilterByTag(name, value string) VMNodes {
	var res VMNodes

	for _, vm := range vms {
		if vm.HasTag(name, value) {
			res = append(res, vm)
		}
	}

	return res
}

func (vms VMNodes) FilterByPrivateAddress(address string) VMNodes {
	var res VMNodes

	for _, vm := range vms {
		if vm.PrivateAddress == address {
			res = append(res, vm)
		}
	}

	return res
}

func (vms VMNodes) FilterByPublicAddress(address string) VMNodes {
	var res VMNodes

	for _, vm := range vms {
		if vm.PublicAddress == address {
			res = append(res, vm)
		}
	}

	return res
}

// Clusterfile represents a set of VMNode(s) as a collection of VM(s)
// along with the region, enviornment, cluster and domain those nodes
// belong to.
type Clusterfile struct {
	Region      string `json:"region"`
	Environment string `json:"environment"`
	Cluster     string `json:"cluster"`
	Domain      string `json:"domain"`

	Nodes VMNodes `json:"nodes"`
}

func (cf *Clusterfile) Validate() error {
	var managers int

	for _, node := range cf.Nodes {
		if node.HasTag(RoleTag, ManagerRole) {
			managers++
		}
	}

	// hard-coded based on knowledge of Raft consensus algorithms
	// where you would typically have 3 or 5 manager nodes to form
	// a quorum.
	if managers == 3 || managers == 5 {
		return nil
	}

	return fmt.Errorf("number of managers should be 3 or 5 not %d", managers)
}

// ReadClusterfile reads a `Clusterfile` or `Clusterfile.json` from an
// `io.Reader` such as an open file or stadnard input and parses it into
// a `ClusterInfo` struct.
func ReadClusterfile(r io.Reader) (Clusterfile, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return Clusterfile{}, fmt.Errorf("error reading from reader: %w", err)
	}

	var clusterFile Clusterfile

	if err := json.Unmarshal(data, &clusterFile); err != nil {
		return Clusterfile{}, fmt.Errorf("error parsing json: %s", err)
	}

	return clusterFile, nil
}
