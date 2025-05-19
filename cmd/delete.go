package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	// Updated Use string to show type options directly
	Use:   "delete (cluster|user|context) <name>",
	Short: "Delete a specific item (cluster, user, context) by name",
	Long: `Delete a specific item from the kubeconfig file.
The first argument specifies the type of item to delete (cluster, user, or context).
The second argument is the name of the item to delete.

Valid types are:
  cluster    Delete a cluster.
  user       Delete a user.
  context    Delete a context.

Warning: Deleting a cluster or user that is currently referenced by one
or more contexts may break those contexts. kedit removes the specified item directly.`,
	Args: cobra.ExactArgs(2), // Requires type and name
	RunE: func(cmd *cobra.Command, args []string) error {
		itemType := args[0] // Will be "cluster", "user", or "context"
		itemName := args[1]

		config, err := loadKubeconfig(resolvedKubeconfigPath)
		if err != nil {
			return fmt.Errorf("error loading kubeconfig from '%s': %w", resolvedKubeconfigPath, err)
		}

		itemExisted := false
		switch itemType {
		case "cluster":
			if _, ok := config.Clusters[itemName]; ok {
				delete(config.Clusters, itemName)
				itemExisted = true
			}
		case "user":
			if _, ok := config.AuthInfos[itemName]; ok {
				delete(config.AuthInfos, itemName)
				itemExisted = true
			}
		case "context":
			if _, ok := config.Contexts[itemName]; ok {
				delete(config.Contexts, itemName)
				itemExisted = true
				if config.CurrentContext == itemName {
					config.CurrentContext = ""
				}
			}
		default:
			// This case should ideally not be reached.
			return fmt.Errorf("invalid type '%s'. Must be one of: cluster, user, context", itemType)
		}

		if !itemExisted {
			fmt.Printf("%s '%s' not found in '%s'. Nothing to delete.\n", itemType, itemName, resolvedKubeconfigPath)
			return nil
		}

		if err := saveKubeconfig(config, resolvedKubeconfigPath); err != nil {
			return fmt.Errorf("error saving kubeconfig to '%s' after deleting %s '%s': %w", resolvedKubeconfigPath, itemType, itemName, err)
		}

		fmt.Printf("Successfully deleted %s '%s' from '%s'.\n", itemType, itemName, resolvedKubeconfigPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
