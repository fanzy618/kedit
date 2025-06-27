package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd"
)

func TestRenameCommand(t *testing.T) {
	// Helper to create a kubeconfig for rename tests.
	createKubeconfigForRename := func(tempDir string) string {
		kubeconfigPath := filepath.Join(tempDir, "config")
		err := ioutil.WriteFile(kubeconfigPath, []byte(`
apiVersion: v1
clusters:
- cluster:
    server: https://old-cluster
  name: old-cluster
- cluster:
    server: https://another-cluster
  name: another-cluster
contexts:
- context:
    cluster: old-cluster
    user: old-user
  name: old-context
- context:
    cluster: another-cluster
    user: another-user
  name: another-context
current-context: old-context
kind: Config
preferences: {}
users:
- name: old-user
  user:
    token: old-token
- name: another-user
  user:
    token: another-token
`), 0644)
		assert.NoError(t, err)
		return kubeconfigPath
	}

	// Test renaming a cluster.
	t.Run("rename cluster", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-rename-cluster-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)
		kubeconfigPath := createKubeconfigForRename(tempDir)

		output := executeCommandC(t, "rename", "cluster", "old-cluster", "new-cluster", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Renamed cluster 'old-cluster' to 'new-cluster'.\nUpdated 1 context(s) to reference the new cluster name 'new-cluster'."
		assert.Equal(t, expectedOutput, output)

		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.Nil(t, config.Clusters["old-cluster"])
		assert.NotNil(t, config.Clusters["new-cluster"])
		assert.Equal(t, "new-cluster", config.Contexts["old-context"].Cluster)
	})

	// Test renaming a user.
	t.Run("rename user", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-rename-user-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)
		kubeconfigPath := createKubeconfigForRename(tempDir)

		output := executeCommandC(t, "rename", "user", "old-user", "new-user", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Renamed user 'old-user' to 'new-user'.\nUpdated 1 context(s) to reference the new user name 'new-user'."
		assert.Equal(t, expectedOutput, output)

		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.Nil(t, config.AuthInfos["old-user"])
		assert.NotNil(t, config.AuthInfos["new-user"])
		assert.Equal(t, "new-user", config.Contexts["old-context"].AuthInfo)
	})

	// Test renaming a context.
	t.Run("rename context", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-rename-context-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)
		kubeconfigPath := createKubeconfigForRename(tempDir)

		output := executeCommandC(t, "rename", "context", "old-context", "new-context", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Renamed context 'old-context' to 'new-context'.\nUpdated current-context from 'old-context' to 'new-context'."
		assert.Equal(t, expectedOutput, output)

		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.Nil(t, config.Contexts["old-context"])
		assert.NotNil(t, config.Contexts["new-context"])
		assert.Equal(t, "new-context", config.CurrentContext)
	})

	// Test renaming a non-existent item.
	t.Run("rename non-existent item", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-rename-nonexistent-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)
		kubeconfigPath := createKubeconfigForRename(tempDir)

		output := executeCommandC(t, "rename", "cluster", "non-existent-cluster", "new-name", "--kubeconfig", kubeconfigPath)
		assert.Contains(t, output, "Error: cluster 'non-existent-cluster' not found in '")

		// Ensure no changes were made.
		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.NotNil(t, config.Clusters["old-cluster"])
	})

	// Test renaming to an existing name.
	t.Run("rename to existing name", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-rename-existing-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)
		kubeconfigPath := createKubeconfigForRename(tempDir)

		output := executeCommandC(t, "rename", "cluster", "old-cluster", "another-cluster", "--kubeconfig", kubeconfigPath)
		assert.Contains(t, output, "Error: a cluster with the name 'another-cluster' already exists")

		// Ensure no changes were made.
		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.NotNil(t, config.Clusters["old-cluster"])
		assert.NotNil(t, config.Clusters["another-cluster"])
	})

	// Test renaming with same old and new name.
	t.Run("rename same name", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-rename-same-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)
		kubeconfigPath := createKubeconfigForRename(tempDir)

		output := executeCommandC(t, "rename", "cluster", "old-cluster", "old-cluster", "--kubeconfig", kubeconfigPath)
		expectedOutput := "The old name and new name are identical ('old-cluster'). No changes made."
		assert.Equal(t, expectedOutput, output)

		// Ensure no changes were made.
		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.NotNil(t, config.Clusters["old-cluster"])
	})
}