package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	sourceKubeconfigPath string // Flag for the source kubeconfig file path
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge <context_name> --from <source_kubeconfig_path>",
	Short: "Merge a context from another kubeconfig file",
	Long: `Import a specific context, along with its referenced cluster and user,
from another kubeconfig file into the current target kubeconfig file.
<context_name> is the name of the context to import.
--from (or -s) specifies the path to the source kubeconfig file.

If an item (context, cluster, or user) with the same name already exists
in the target kubeconfig, it will be overwritten by the item from the source.`,
	Args: cobra.ExactArgs(1), // Requires context_name
	RunE: func(cmd *cobra.Command, args []string) error {
		contextToMerge := args[0]

		if sourceKubeconfigPath == "" {
			return fmt.Errorf("flag --from <source_kubeconfig_path> is required for the merge command")
		}

		// Expand source path
		expandedSourcePath, err := homedir.Expand(sourceKubeconfigPath)
		if err != nil {
			return fmt.Errorf("error expanding source kubeconfig path '%s': %w", sourceKubeconfigPath, err)
		}

		// Load source kubeconfig. It must exist.
		if _, statErr := os.Stat(expandedSourcePath); os.IsNotExist(statErr) {
			return fmt.Errorf("source kubeconfig file '%s' not found", expandedSourcePath)
		}
		sourceConfig, err := clientcmd.LoadFromFile(expandedSourcePath) // Use direct LoadFromFile for source
		if err != nil {
			return fmt.Errorf("failed to load source kubeconfig from '%s': %w", expandedSourcePath, err)
		}

		// Find the context in the source config
		sourceContext, contextExists := sourceConfig.Contexts[contextToMerge]
		if !contextExists {
			return fmt.Errorf("context '%s' not found in source kubeconfig '%s'", contextToMerge, expandedSourcePath)
		}

		// Get referenced cluster and user names from the source context
		clusterNameFromSource := sourceContext.Cluster
		userNameFromSource := sourceContext.AuthInfo // AuthInfo field in Context struct stores the user name

		// A context must reference a cluster
		if clusterNameFromSource == "" {
			return fmt.Errorf("context '%s' in source kubeconfig '%s' does not reference a cluster", contextToMerge, expandedSourcePath)
		}

		// Get cluster details from source config
		sourceCluster, clusterExistsInSource := sourceConfig.Clusters[clusterNameFromSource]
		if !clusterExistsInSource {
			return fmt.Errorf("cluster '%s' (referenced by context '%s') not found in source kubeconfig '%s'", clusterNameFromSource, contextToMerge, expandedSourcePath)
		}

		// Get user details from source config (if a user is specified)
		var sourceUser *api.AuthInfo
		userExistsInSource := false
		if userNameFromSource != "" {
			sourceUser, userExistsInSource = sourceConfig.AuthInfos[userNameFromSource]
			if !userExistsInSource {
				return fmt.Errorf("user '%s' (referenced by context '%s') not found in source kubeconfig '%s'", userNameFromSource, contextToMerge, expandedSourcePath)
			}
		}

		// Load target kubeconfig (our helper `loadKubeconfig` creates an empty one if it doesn't exist)
		targetConfig, err := loadKubeconfig(resolvedKubeconfigPath)
		if err != nil {
			return fmt.Errorf("error loading target kubeconfig '%s': %w", resolvedKubeconfigPath, err)
		}
		// loadKubeconfig ensures maps are initialized in targetConfig

		// Add/Overwrite context, cluster, and user to the target config
		targetConfig.Contexts[contextToMerge] = sourceContext
		targetConfig.Clusters[clusterNameFromSource] = sourceCluster
		if userNameFromSource != "" && userExistsInSource { // Only add user if it was specified and found
			targetConfig.AuthInfos[userNameFromSource] = sourceUser
		}

		// Save the modified target kubeconfig
		if err := saveKubeconfig(targetConfig, resolvedKubeconfigPath); err != nil {
			return fmt.Errorf("error saving target kubeconfig '%s' after merge: %w", resolvedKubeconfigPath, err)
		}

		userMergeMsg := "no specific user"
		if userNameFromSource != "" {
			userMergeMsg = fmt.Sprintf("user '%s'", userNameFromSource)
		}
		fmt.Printf("Successfully merged context '%s' (with cluster '%s' and %s) from '%s' into '%s'.\n",
			contextToMerge, clusterNameFromSource, userMergeMsg, expandedSourcePath, resolvedKubeconfigPath)
		return nil
	},
}

func init() {
	mergeCmd.Flags().StringVarP(&sourceKubeconfigPath, "from", "s", "", "Path to the source kubeconfig file (required)")
	// MarkFlagRequired is an option, but manual check in RunE is also fine.
	// if err := mergeCmd.MarkFlagRequired("from"); err != nil {
	// 	 fmt.Fprintf(os.Stderr, "Error marking flag 'from' as required: %v\n", err)
	// 	 os.Exit(1)
	// }
	rootCmd.AddCommand(mergeCmd)
}
