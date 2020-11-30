package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gitlab.mgt.aom.australiacloud.com.au/aom/swarm/internal"
)

var configFile string

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
		// set logging level
		if viper.GetBool("debug") {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
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
		&configFile, "config", "",
		"config file (default is $HOME/.swarm.yaml)",
	)

	RootCmd.PersistentFlags().BoolP(
		"debug", "D", false,
		"Enable debug logging",
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

	viper.BindPFlag("ssh-addr", RootCmd.PersistentFlags().Lookup("ssh-addr"))

	viper.BindPFlag("ssh-key", RootCmd.PersistentFlags().Lookup("ssh-key"))
	viper.SetDefault("ssh-key", internal.DefaultSSHKey)

	viper.BindPFlag("ssh-user", RootCmd.PersistentFlags().Lookup("ssh-user"))
	viper.SetDefault("ssh-user", internal.DefaultSSHUser)

	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.SetDefault("debug", false)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
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