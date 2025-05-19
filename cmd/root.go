package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	// cfgFile is the path to the kubeconfig file, set by the --kubeconfig flag.
	cfgFile string
	// resolvedKubeconfigPath is the actual path to be used by commands after resolving defaults and home dir.
	resolvedKubeconfigPath string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kedit",
	Short: "A CLI tool for managing kubectl config files.",
	Long: `kedit is a command-line utility designed to simplify the management of
Kubernetes configuration files (kubeconfig). It allows users to list
clusters, users, and contexts, delete specific entries, prune
unreferenced items, and merge contexts from other kubeconfig files.`,
	// This function runs before any subcommand's RunE, ensuring resolvedKubeconfigPath is set.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var path string
		var err error

		if cfgFile != "" {
			path = cfgFile
		} else {
			// Default to $HOME/.kube/config
			home, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return fmt.Errorf("failed to get user home directory: %w", homeErr)
			}
			path = filepath.Join(home, ".kube", "config")
		}

		// Expand path (e.g., ~ to actual home directory)
		resolvedKubeconfigPath, err = homedir.Expand(path)
		if err != nil {
			return fmt.Errorf("error expanding path '%s': %w", path, err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra already prints the error to stderr, so we just exit.
		os.Exit(1)
	}
}

func init() {
	// Register global persistent flag for --kubeconfig
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "kubeconfig", "k", "", "Path to the kubeconfig file (default is $HOME/.kube/config)")
}

// Utility functions for kubeconfig operations

// loadKubeconfig loads the configuration from the given file path.
// If the file does not exist, it returns a new empty, initialized config.
func loadKubeconfig(filePath string) (*api.Config, error) {
	config, err := clientcmd.LoadFromFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return a new empty config.
			// This is useful for 'merge' into a new file or if other commands
			// operate on a non-existent file (they'll typically report 'not found' or empty lists).
			return api.NewConfig(), nil
		}
		return nil, fmt.Errorf("failed to load kubeconfig from '%s': %w", filePath, err)
	}

	// Ensure maps are initialized, though clientcmd.LoadFromFile and api.NewConfig should handle this.
	// This is a defensive measure.
	if config.Clusters == nil {
		config.Clusters = make(map[string]*api.Cluster)
	}
	if config.AuthInfos == nil {
		config.AuthInfos = make(map[string]*api.AuthInfo)
	}
	if config.Contexts == nil {
		config.Contexts = make(map[string]*api.Context)
	}
	return config, nil
}

// saveKubeconfig saves the configuration to the given file path.
// It creates the parent directory if it doesn't exist.
func saveKubeconfig(config *api.Config, filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil { // 0755 for directory permissions
			return fmt.Errorf("failed to create directory '%s': %w", dir, mkdirErr)
		}
	}

	err := clientcmd.WriteToFile(*config, filePath)
	if err != nil {
		return fmt.Errorf("failed to save kubeconfig to '%s': %w", filePath, err)
	}
	return nil
}
