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
