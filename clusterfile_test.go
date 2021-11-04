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

// TestReadClusterfile tests the `ReadClusterfiel` function that reads and
// parses a valid `Clusterfile` or `Clusterfile.json`.
func TestReadClusterfile(t *testing.T) {
	assert := assert.New(t)

	clusterfile := `{
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
  }]
}
`
	actual, err := ReadClusterfile(bytes.NewBufferString(clusterfile))
	assert.Nil(err)
	expected := Clusterfile{
		Region:      "local",
		Environment: "test",
		Cluster:     "c1",
		Domain:      "localdomain",
		Nodes: []VMNode{
			VMNode{
				Hostname:       "dm1",
				PublicAddress:  "10.0.0.1",
				PrivateAddress: "172.16.0.1",
				Tags: map[string]string{
					"role": "manager",
				},
			},
		},
	}
	assert.Equal(expected, actual)
}
