package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	// api "k8s.io/client-go/tools/clientcmd/api" // Not directly used for type definitions here
)

// renameCmd represents the rename command
var renameCmd = &cobra.Command{
	Use:   "rename (cluster|user|context) <old_name> <new_name>",
	Short: "Rename a cluster, user, or context",
	Long: `Rename an existing cluster, user, or context in the kubeconfig file.

When renaming a cluster or a user, all contexts that reference the old name
will be updated to point to the new name.
When renaming a context that is the current-context, the current-context
field in the kubeconfig will also be updated to the new name.

Arguments:
  (cluster|user|context): The type of item to rename. Must be one of 'cluster', 'user', or 'context'.
  <old_name>:             The current name of the item to be renamed.
  <new_name>:             The desired new name for the item. The new name must not already exist for that item type.`,
	Args: cobra.ExactArgs(3), // Requires type, old_name, and new_name
	RunE: func(cmd *cobra.Command, args []string) error {
		itemType := args[0]
		oldName := args[1]
		newName := args[2]

		if oldName == newName {
			fmt.Printf("The old name and new name are identical ('%s'). No changes made.\n", oldName)
			return nil
		}

		config, err := loadKubeconfig(resolvedKubeconfigPath)
		if err != nil {
			// This error occurs if loading fails for reasons other than the file not existing.
			// If the file doesn't exist, loadKubeconfig returns an empty config,
			// and subsequent checks for oldName will correctly report "not found".
			return fmt.Errorf("error loading kubeconfig from '%s': %w", resolvedKubeconfigPath, err)
		}

		switch itemType {
		case "cluster":
			// 1. Check if the old cluster name exists.
			clusterToRename, ok := config.Clusters[oldName]
			if !ok {
				return fmt.Errorf("cluster '%s' not found in '%s'", oldName, resolvedKubeconfigPath)
			}

			// 2. Check if the new cluster name already exists.
			if _, exists := config.Clusters[newName]; exists {
				return fmt.Errorf("a cluster with the name '%s' already exists", newName)
			}

			// 3. Perform rename for the cluster entry.
			config.Clusters[newName] = clusterToRename
			delete(config.Clusters, oldName)
			fmt.Printf("Renamed cluster '%s' to '%s'.\n", oldName, newName)

			// 4. Update references in all contexts.
			updatedContextsCount := 0
			for _, contextDetails := range config.Contexts {
				if contextDetails.Cluster == oldName {
					// contextDetails is a pointer, so this modifies the original.
					contextDetails.Cluster = newName
					updatedContextsCount++
				}
			}
			if updatedContextsCount > 0 {
				fmt.Printf("Updated %d context(s) to reference the new cluster name '%s'.\n", updatedContextsCount, newName)
			}

		case "user":
			// 1. Check if the old user name exists.
			userToRename, ok := config.AuthInfos[oldName]
			if !ok {
				return fmt.Errorf("user '%s' not found in '%s'", oldName, resolvedKubeconfigPath)
			}

			// 2. Check if the new user name already exists.
			if _, exists := config.AuthInfos[newName]; exists {
				return fmt.Errorf("a user with the name '%s' already exists", newName)
			}

			// 3. Perform rename for the user entry.
			config.AuthInfos[newName] = userToRename
			delete(config.AuthInfos, oldName)
			fmt.Printf("Renamed user '%s' to '%s'.\n", oldName, newName)

			// 4. Update references in all contexts.
			updatedContextsCount := 0
			for _, contextDetails := range config.Contexts { // contextNameInMap not needed here for msg
				if contextDetails.AuthInfo == oldName {
					contextDetails.AuthInfo = newName
					updatedContextsCount++
				}
			}
			if updatedContextsCount > 0 {
				fmt.Printf("Updated %d context(s) to reference the new user name '%s'.\n", updatedContextsCount, newName)
			}

		case "context":
			// 1. Check if the old context name exists.
			contextToRename, ok := config.Contexts[oldName]
			if !ok {
				return fmt.Errorf("context '%s' not found in '%s'", oldName, resolvedKubeconfigPath)
			}

			// 2. Check if the new context name already exists.
			if _, exists := config.Contexts[newName]; exists {
				return fmt.Errorf("a context with the name '%s' already exists", newName)
			}

			// 3. Perform rename for the context entry.
			config.Contexts[newName] = contextToRename
			delete(config.Contexts, oldName)
			fmt.Printf("Renamed context '%s' to '%s'.\n", oldName, newName)

			// 4. Update current-context field if it was the renamed context.
			if config.CurrentContext == oldName {
				config.CurrentContext = newName
				fmt.Printf("Updated current-context from '%s' to '%s'.\n", oldName, newName)
			}

		default:
			return fmt.Errorf("invalid item type '%s'. Must be one of: cluster, user, context", itemType)
		}

		// Save the modified kubeconfig file.
		if err := saveKubeconfig(config, resolvedKubeconfigPath); err != nil {
			return fmt.Errorf("error saving kubeconfig to '%s' after renaming: %w", resolvedKubeconfigPath, err)
		}

		// General success message - specific actions already logged.
		// fmt.Printf("Successfully completed rename operation for %s '%s' to '%s' in '%s'.\n", itemType, oldName, newName, resolvedKubeconfigPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
