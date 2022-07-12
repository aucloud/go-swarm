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

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aucloud/go-swarm"
	"github.com/aucloud/go-swarm/internal"
)

var (
	config  string
	manager *swarm.Manager
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "swarm",
	Version: internal.FullVersion(),
	Short:   "Swarm Manager CLI",
	Long: `This is a command-line Docker Swarm Manager

This tool is an implemtnation of the swarm management library used to help
facilitate and automation the creation and management of Docker Swarm Clusters.

Supported functions include:

- Creating a Swarm Clsuter
- Adding new worker or manager nodes
- Draining nodes
- Removing nodes
- Displaying cluster information`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error

		// set logging level
		if viper.GetBool("debug") {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		var switcher swarm.Switcher

		if viper.GetBool("use-local") {
			localSwitcher, err := swarm.NewLocalSwitcher()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error creating local switcher: %s\n", err)
				os.Exit(-1)
			}
			if localSwitcher.Switch(context.Background(), ""); err != nil {
				fmt.Fprintf(os.Stderr, "error switching to local node: %s\n", err)
				os.Exit(-1)
			}

			switcher = localSwitcher
		} else {
			timeout := viper.GetDuration("ssh-timeout")
			user := viper.GetString("ssh-user")
			addr := viper.GetString("ssh-addr")
			key := viper.GetString("ssh-key")
			useAgent := viper.GetBool("use-ssh-agent")

			var sshAuth swarm.SSHRunnerAuth
			if useAgent {
				sshAuth = swarm.NewAgentSSHRunnerAuth(os.Getenv("SSH_AUTH_SOCK"))
			} else {
				sshAuth = swarm.NewKeySSHRunnerAuth(key)
			}

			sshSwitcher, err := swarm.NewSSHSwitcher(user, addr, sshAuth, timeout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error creating ssh switcher: %s\n", err)
				os.Exit(-1)
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			if sshSwitcher.Switch(ctx, addr); err != nil {
				fmt.Fprintf(os.Stderr, "error switching to remote node %s: %s\n", addr, err)
				os.Exit(-1)
			}

			switcher = sshSwitcher
		}

		if manager, err = swarm.NewManager(switcher); err != nil {
			fmt.Fprintf(os.Stderr, "error creating manager: %s\n", err)
			os.Exit(-1)
		}
	},
}

// Execute adds all child commands to the root command
// and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(
		&config, "config", "",
		"config file (default is $HOME/.swarm.yaml)",
	)

	RootCmd.PersistentFlags().BoolP(
		"debug", "D", false,
		"Enable debug logging",
	)

	RootCmd.PersistentFlags().BoolP(
		"use-local", "L", false,
		"Use local socket (connects directly to Docker UNIX socket)",
	)

	RootCmd.PersistentFlags().DurationP(
		"ssh-timeout", "T", time.Minute*5,
		"Timeout to use for SSH connections before giving up (retries failed connections)",
	)

	RootCmd.PersistentFlags().StringP(
		"ssh-addr", "A", "",
		"SSH Address to connect to",
	)

	RootCmd.PersistentFlags().StringP(
		"ssh-key", "K", internal.DefaultSSHKey,
		"SSH Key to use for remote execution",
	)

	RootCmd.PersistentFlags().StringP(
		"ssh-user", "U", internal.DefaultSSHUser,
		"SSH User to use for remote execution",
	)

	RootCmd.PersistentFlags().StringP(
		"sock-path", "S", "/var/run/docker.sock",
		"Path to Docker UNIX Socket",
	)

	viper.BindPFlag("use-local", RootCmd.PersistentFlags().Lookup("use-local"))
	viper.SetDefault("use-local", false)

	viper.BindPFlag("ssh-addr", RootCmd.PersistentFlags().Lookup("ssh-addr"))

	viper.BindPFlag("ssh-key", RootCmd.PersistentFlags().Lookup("ssh-key"))
	viper.SetDefault("ssh-key", internal.DefaultSSHKey)

	viper.BindPFlag("ssh-user", RootCmd.PersistentFlags().Lookup("ssh-user"))
	viper.SetDefault("ssh-user", internal.DefaultSSHUser)

	viper.BindPFlag("sock-path", RootCmd.PersistentFlags().Lookup("sock-path"))
	viper.SetDefault("sock-path", internal.DefaultSockPath)

	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.SetDefault("debug", false)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if config != "" {
		// Use config file from the flag.
		viper.SetConfigFile(config)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".swarm.yaml")
	}

	// from the environment
	viper.SetEnvPrefix("SWARM")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
