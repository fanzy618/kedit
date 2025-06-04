package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	// Updated Use string to show type options directly
	Use:   "list (cluster|user|context|all)",
	Short: "List names of a specific type (cluster, user, context, or all)",
	Long: `List names of a specific item type within the target kubeconfig file.
The first argument specifies the type of item to list:
  cluster    List all cluster names.
  user       List all user names.
  context    List all context names.
  all        List all clusters, users and contexts.`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument which is the type
	RunE: func(cmd *cobra.Command, args []string) error {
		listType := args[0] // Will be "cluster", "user", "context", or "all"

		config, err := loadKubeconfig(resolvedKubeconfigPath)
		if err != nil {
			return fmt.Errorf("error loading kubeconfig from '%s': %w", resolvedKubeconfigPath, err)
		}

		switch listType {
		case "cluster":
			if len(config.Clusters) == 0 {
				fmt.Println("No clusters found.")
				return nil
			}
			names := make([]string, 0, len(config.Clusters))
			for name := range config.Clusters {
				names = append(names, name)
			}
			sort.Strings(names)
			fmt.Println("Clusters:")
			for _, name := range names {
				fmt.Printf("- %s\n", name)
			}
		case "user":
			if len(config.AuthInfos) == 0 {
				fmt.Println("No users found.")
				return nil
			}
			names := make([]string, 0, len(config.AuthInfos))
			for name := range config.AuthInfos {
				names = append(names, name)
			}
			sort.Strings(names)
			fmt.Println("Users:")
			for _, name := range names {
				fmt.Printf("- %s\n", name)
			}
		case "context":
			if len(config.Contexts) == 0 {
				fmt.Println("No contexts found.")
				return nil
			}
			names := make([]string, 0, len(config.Contexts))
			for name := range config.Contexts {
				names = append(names, name)
			}
			sort.Strings(names)
			fmt.Println("Contexts:")
			for _, name := range names {
				fmt.Printf("- %s\n", name)
			}
		case "all":
			if len(config.Clusters) == 0 {
				fmt.Println("No clusters found.")
			} else {
				clusterNames := make([]string, 0, len(config.Clusters))
				for name := range config.Clusters {
					clusterNames = append(clusterNames, name)
				}
				sort.Strings(clusterNames)
				fmt.Println("Clusters:")
				for _, name := range clusterNames {
					fmt.Printf("- %s\n", name)
				}
			}

			if len(config.AuthInfos) == 0 {
				fmt.Println("No users found.")
			} else {
				userNames := make([]string, 0, len(config.AuthInfos))
				for name := range config.AuthInfos {
					userNames = append(userNames, name)
				}
				sort.Strings(userNames)
				fmt.Println("Users:")
				for _, name := range userNames {
					fmt.Printf("- %s\n", name)
				}
			}

			if len(config.Contexts) == 0 {
				fmt.Println("No contexts found.")
			} else {
				contextNames := make([]string, 0, len(config.Contexts))
				for name := range config.Contexts {
					contextNames = append(contextNames, name)
				}
				sort.Strings(contextNames)
				fmt.Println("Contexts:")
				for _, name := range contextNames {
					fmt.Printf("- %s\n", name)
				}
			}
		default:
			// This case should ideally not be reached if Cobra validates based on Use line,
			// but good for robustness and if Use line is manually typed wrong by user.
			return fmt.Errorf("invalid type '%s'. Must be one of: cluster, user, context, all", listType)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
