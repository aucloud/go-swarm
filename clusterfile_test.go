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
	"testing"

	"github.com/stretchr/testify/assert"
)

const testClusterfile = `{
  "region": "local",
  "environment": "test",
  "cluster": "c1",
  "domain": "localdomain",
  "nodes": [{
    "hostname": "dm1",
    "public_address": "10.0.0.1",
    "private_address": "172.16.0.1",
    "tags": {
	  "role": "manager"
	}
  }, {
    "hostname": "dw1",
    "public_address": "10.0.0.2",
    "private_address": "172.16.0.2",
    "tags": {
	  "role": "worker"
	}
  }]
}
`

// TestReadClusterfile tests the `ReadClusterfiel` function that reads and
// parses a valid `Clusterfile` or `Clusterfile.json`.
func TestReadClusterfile(t *testing.T) {
	assert := assert.New(t)

	actual, err := ReadClusterfile(bytes.NewBufferString(testClusterfile))
	assert.Nil(err)
	expected := Clusterfile{
		Region:      "local",
		Environment: "test",
		Cluster:     "c1",
		Domain:      "localdomain",
		Nodes: VMNodes{
			{
				Hostname:       "dm1",
				PublicAddress:  "10.0.0.1",
				PrivateAddress: "172.16.0.1",
				Tags: map[string]string{
					"role": "manager",
				},
			},
			{
				Hostname:       "dw1",
				PublicAddress:  "10.0.0.2",
				PrivateAddress: "172.16.0.2",
				Tags: map[string]string{
					"role": "worker",
				},
			},
		},
	}
	assert.Equal(expected, actual)
}

// TestFilterByTags tests the `VMNodes.FilterByTags()` functionality to ensure
// we can filter a list of nodes by tag/value pairs.
func TestFilterByTags(t *testing.T) {
	assert := assert.New(t)

	cf, err := ReadClusterfile(bytes.NewBufferString(testClusterfile))
	assert.Nil(err)

	vms := cf.Nodes.FilterByTag("role", "manager")
	assert.Len(vms, 1)
	assert.Equal(vms[0].Hostname, "dm1")
}

// TestFilterByPrivateAddress tests the `VMNodes.FilterByPrivateAddress()`
// functionality to ensure we can filter a list of nodes by private address.
func TestFilterByPrivateAddress(t *testing.T) {
	assert := assert.New(t)

	cf, err := ReadClusterfile(bytes.NewBufferString(testClusterfile))
	assert.Nil(err)

	vms := cf.Nodes.FilterByPrivateAddress("172.16.0.2")
	assert.Len(vms, 1)
	assert.Equal(vms[0].Hostname, "dw1")
}

// TestFilterByPublicAddress tests the `VMNodes.FilterByPublicAddress()`
// functionality to ensure we can filter a list of nodes by public address.
func TestFilterByPublicAddress(t *testing.T) {
	assert := assert.New(t)

	cf, err := ReadClusterfile(bytes.NewBufferString(testClusterfile))
	assert.Nil(err)

	vms := cf.Nodes.FilterByPublicAddress("10.0.0.1")
	assert.Len(vms, 1)
	assert.Equal(vms[0].Hostname, "dm1")
}
