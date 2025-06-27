
package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
	// Create a temporary directory for the test.
	tempDir, err := ioutil.TempDir("", "kedit-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a dummy kubeconfig file.
	kubeconfigPath := filepath.Join(tempDir, "config")
	err = ioutil.WriteFile(kubeconfigPath, []byte(`
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

	// Test list clusters.
	t.Run("list clusters", func(t *testing.T) {
		output := executeCommandC(t, "list", "cluster", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Clusters:\n- cluster1\n- cluster2"
		assert.Equal(t, expectedOutput, output)
	})

	// Test list users.
	t.Run("list users", func(t *testing.T) {
		output := executeCommandC(t, "list", "user", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Users:\n- user1\n- user2"
		assert.Equal(t, expectedOutput, output)
	})

	// Test list contexts.
	t.Run("list contexts", func(t *testing.T) {
		output := executeCommandC(t, "list", "context", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Contexts:\n- context1\n- context2"
		assert.Equal(t, expectedOutput, output)
	})

	// Test list all.
	t.Run("list all", func(t *testing.T) {
		output := executeCommandC(t, "list", "all", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Clusters:\n- cluster1\n- cluster2\nUsers:\n- user1\n- user2\nContexts:\n- context1\n- context2"
		assert.Equal(t, expectedOutput, output)
	})

	// Test list with empty kubeconfig.
	t.Run("list with empty kubeconfig", func(t *testing.T) {
		// Create an empty kubeconfig file.
		emptyKubeconfigPath := filepath.Join(tempDir, "empty-config")
		err = ioutil.WriteFile(emptyKubeconfigPath, []byte(""), 0644)
		assert.NoError(t, err)

		output := executeCommandC(t, "list", "all", "--kubeconfig", emptyKubeconfigPath)
		expectedOutput := "No clusters found.\nNo users found.\nNo contexts found."
		assert.Equal(t, expectedOutput, output)
	})
}
