package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd/api"
)

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove unreferenced clusters and users",
	Long:  `Remove all clusters and users from the kubeconfig file that are not referenced by any existing context.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadKubeconfig(resolvedKubeconfigPath)
		if err != nil {
			return fmt.Errorf("error loading kubeconfig from '%s': %w", resolvedKubeconfigPath, err)
		}

		if len(config.Contexts) == 0 {
			// If there are no contexts, all clusters and users are unreferenced.
			originalClusterCount := len(config.Clusters)
			originalUserCount := len(config.AuthInfos)

			if originalClusterCount == 0 && originalUserCount == 0 {
				fmt.Println("No contexts, clusters, or users found. Nothing to prune.")
				return nil
			}

			config.Clusters = make(map[string]*api.Cluster)
			config.AuthInfos = make(map[string]*api.AuthInfo)

			if err := saveKubeconfig(config, resolvedKubeconfigPath); err != nil {
				return fmt.Errorf("error saving kubeconfig to '%s' after pruning all clusters/users: %w", resolvedKubeconfigPath, err)
			}
			fmt.Printf("No contexts found. Pruned %d cluster(s) and %d user(s) from '%s'.\n", originalClusterCount, originalUserCount, resolvedKubeconfigPath)
			return nil
		}

		// Collect all referenced cluster and user names
		referencedClusters := make(map[string]bool)
		referencedUsers := make(map[string]bool)
		for _, context := range config.Contexts {
			if context.Cluster != "" {
				referencedClusters[context.Cluster] = true
			}
			if context.AuthInfo != "" { // AuthInfo is the user name in a context
				referencedUsers[context.AuthInfo] = true
			}
		}

		// Prune unreferenced clusters
		originalClusterCount := len(config.Clusters)
		prunedClustersCount := 0
		newClusters := make(map[string]*api.Cluster)
		for name, cluster := range config.Clusters {
			if referencedClusters[name] {
				newClusters[name] = cluster
			} else {
				prunedClustersCount++
			}
		}
		config.Clusters = newClusters

		// Prune unreferenced users
		originalUserCount := len(config.AuthInfos)
		prunedUsersCount := 0
		newAuthInfos := make(map[string]*api.AuthInfo)
		for name, user := range config.AuthInfos {
			if referencedUsers[name] {
				newAuthInfos[name] = user
			} else {
				prunedUsersCount++
			}
		}
		config.AuthInfos = newAuthInfos

		if prunedClustersCount == 0 && prunedUsersCount == 0 {
			fmt.Printf("No unreferenced clusters or users found in '%s'. Nothing to prune.\n", resolvedKubeconfigPath)
			return nil
		}

		if err := saveKubeconfig(config, resolvedKubeconfigPath); err != nil {
			return fmt.Errorf("error saving kubeconfig to '%s' after pruning: %w", resolvedKubeconfigPath, err)
		}

		fmt.Printf("Pruned %d cluster(s) (out of %d) and %d user(s) (out of %d) from '%s'.\n",
			prunedClustersCount, originalClusterCount, prunedUsersCount, originalUserCount, resolvedKubeconfigPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)
}
