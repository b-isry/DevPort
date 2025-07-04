// cmd/root.go
package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var logger *slog.Logger

var rootCmd = &cobra.Command{
	Use:   "devport",
	Short: "A blazingly fast dependency teleporter for Node.js",
	Long: `DevPort is a next-generation Node dependency engine that virtualizes node_modules.
It creates a shared, content-addressable cache to make dependency management
across machines and branches seamless and instant.`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")

		var logLevel slog.Level
		if verbose {
			logLevel = slog.LevelDebug
		} else {
			logLevel = slog.LevelInfo
		}
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
		logger = slog.New(handler)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./.devport.yaml or $HOME/.devport/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose logging")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".devport")	
		viper.SetConfigType("yaml")

	}
	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
}
	viper.SetConfigFile(".devport.secret.yaml")
	err := viper.MergeInConfig()
	if err == nil {
		fmt.Println("Using secret config file:", viper.ConfigFileUsed())
	}
	viper.SetDefault("root_directory", "node_modules")

	viper.SetDefault("s3.endpoint", "http://localhost:9000")
	viper.SetDefault("s3.bucket", "devport-cache")
	viper.SetDefault("s3.region", "us-east-1")
	viper.SetDefault("s3.access_key_id", "devport-admin")
	viper.SetDefault("s3.secret_access_key", "devport-password")
	viper.SetDefault("s3.use_ssl", false)
}
