package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd"
)

func TestDeleteCommand(t *testing.T) {
	// Define a helper function to create a fresh kubeconfig for each test.
	createFreshKubeconfig := func(tempDir string) string {
		kubeconfigPath := filepath.Join(tempDir, "config")
		err := ioutil.WriteFile(kubeconfigPath, []byte(`
apiVersion: v1
clusters:
- cluster:
    server: https://cluster1
  name: cluster1
- cluster:
    server: https://cluster2
  name: cluster2
contexts:
- context:
    cluster: cluster1
    user: user1
  name: context1
- context:
    cluster: cluster2
    user: user2
  name: context2
current-context: context1
kind: Config
preferences: {}
users:
- name: user1
  user:
    token: token1
- name: user2
  user:
    token: token2
`), 0644)
		assert.NoError(t, err)
		return kubeconfigPath
	}

	// Test delete cluster.
	t.Run("delete cluster", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-delete-cluster-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)
		kubeconfigPath := createFreshKubeconfig(tempDir)

		// Execute the delete command.
		output := executeCommandC(t, "delete", "cluster", "cluster1", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Successfully deleted cluster 'cluster1' from '" + kubeconfigPath + "'."
		assert.Equal(t, expectedOutput, output)

		// Load the kubeconfig and assert that the cluster is deleted.
		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.Clusters, 1)
		assert.Nil(t, config.Clusters["cluster1"])
		assert.NotNil(t, config.Clusters["cluster2"])
	})

	// Test delete user.
	t.Run("delete user", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-delete-user-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)
		kubeconfigPath := createFreshKubeconfig(tempDir)

		// Execute the delete command.
		output := executeCommandC(t, "delete", "user", "user1", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Successfully deleted user 'user1' from '" + kubeconfigPath + "'."
		assert.Equal(t, expectedOutput, output)

		// Load the kubeconfig and assert that the user is deleted.
		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.AuthInfos, 1)
		assert.Nil(t, config.AuthInfos["user1"])
		assert.NotNil(t, config.AuthInfos["user2"])
	})

	// Test delete context.
	t.Run("delete context", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-delete-context-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)
		kubeconfigPath := createFreshKubeconfig(tempDir)

		// Execute the delete command.
		output := executeCommandC(t, "delete", "context", "context1", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Successfully deleted context 'context1' from '" + kubeconfigPath + "'."
		assert.Equal(t, expectedOutput, output)

		// Load the kubeconfig and assert that the context is deleted.
		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.Contexts, 1)
		assert.Nil(t, config.Contexts["context1"])
		assert.NotNil(t, config.Contexts["context2"])
		assert.Empty(t, config.CurrentContext) // Should clear current-context if deleted
	})

	// Test delete non-existent item.
	t.Run("delete non-existent item", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-delete-nonexistent-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)
		kubeconfigPath := createFreshKubeconfig(tempDir)

		// Execute the delete command for a non-existent cluster.
		output := executeCommandC(t, "delete", "cluster", "non-existent-cluster", "--kubeconfig", kubeconfigPath)
		expectedOutput := "cluster 'non-existent-cluster' not found in '" + kubeconfigPath + "'. Nothing to delete."
		assert.Equal(t, expectedOutput, output)

		// Load the kubeconfig and assert that no changes were made.
		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.Clusters, 2)
		assert.NotNil(t, config.Clusters["cluster1"])
		assert.NotNil(t, config.Clusters["cluster2"])
	})
}