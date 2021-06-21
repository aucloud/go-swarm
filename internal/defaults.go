package internal

const (
	// DefaultSSHUser is the default SSH Username when executing remote commands
	DefaultSSHUser = "rancher"

	// DefaultSSHKey is the default SSH Kry when executing remote commands
	DefaultSSHKey = "$HOME/.ssh/id_rsa"

	// DefaultSockPath is the default path to the Docker API's UNIX Socket
	DefaultSockPath = "/var/run/docker.sock"

	// MinSwarmClusterNodes is the minimum number of  nodes to form a swam cluster
	MinSwarmClusterNodes = 1
)
