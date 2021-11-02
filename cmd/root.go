package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	dataDir     string
	connTimeout int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "brbr",
	Short: "A command line tool for sending and receiving metagrams",
	Long: `brbr can establish end-to-end encrypted connections with other instances over the 
Tor network and send/receive messages called metagrams`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig, initDataDir)

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.brbr.json)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".brbr" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("json")
		viper.SetConfigName(".brbr")
	}

	viper.SetEnvPrefix("brbr") // prefix environment variables with BRBR_
	viper.AutomaticEnv()       // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// initDataDir set the data dir and creates it if needed
func initDataDir() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	dataDir = home + "/.brbr"

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "Creating data directory", dataDir)
		err := os.Mkdir(dataDir, 0700)
		cobra.CheckErr(err)
	}

	fmt.Fprintln(os.Stderr, "Using data dir:", dataDir)
}
