# go-swarm

[![Go](https://github.com/aucloud/go-swarm/actions/workflows/go.yml/badge.svg)](https://github.com/aucloud/go-swarm/actions/workflows/go.yml)

`go-swarm` is a Go library and ccommand-line tool for managing the creation
and maintenance of Docker Swarm cluster.

Features:

- Creates new Swarm Cluters given a `Clusterfile` as input.
- Retrives information about Swarm Clsuters.
- Join new workers or managers to an existing Swarm Cluster.
- Assigning Swarm Labels based on underlying VM Node labels.
- Add-hoc adding new worker or manager nodes
- Draining Swarm nodes
- Removing Swarm ndoes

## Install

Currently tehre is a command-line tool called `swarm` that can be installed with:

```#!console
go install github.com/aucloud/cmd/swarm@latest
```

Using as a library is to be documented at a later date.

## usage

Using the `swarm` CLI tool is easy:

```#!console
$ ./swarm
This is a command-line Docker Swarm Manager

This tool is an implementation of the swarm management library used to help
facilitate and automate the creation and management of Docker Swarm Clusters.

Supported functions include:

- Creating a Swarm Clsuter
- Adding new worker or manager nodes
- Draining nodes
- Removing nodes
- Displaying cluster information

Usage:
  swarm [command]

Available Commands:
  create      Creates a new Swarm Cluster
  help        Help about any command
  info        Retrieve and display Swarm Cluster Information
  status      Retrieve and display Swarm Cluster Status

Flags:
      --config string     config file (default is $HOME/.swarm.yaml)
  -D, --debug             Enable debug logging
  -h, --help              help for swarm
  -A, --ssh-addr string   SSH Address to connect to
  -K, --ssh-key string    SSH Key to use for remote execution (default "$HOME/.ssh/id_rsa")
  -U, --ssh-user string   SSH User to use for remote execution (default "rancher")
  -v, --version           version for swarm

Use "swarm [command] --help" for more information about a command.
```

For example to create a new Swarm clsuter from a Terraform run:

```#!console
terraform output -json Clusterfile | swarm -D create -
```

This will take the `Clusterfile` (_a JSON representing the VM Nodes created via Terraform_)
and create a multi-manager Swarm Cluster and join all worker nodes and display the
cluster status at the end.


## License

`go-swarm` is licsned under the terms of the [AGPLv3](/LICENSE)