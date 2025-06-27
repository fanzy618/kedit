
package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd"
)

func TestPruneCommand(t *testing.T) {
	// Helper to create a kubeconfig with various states.
	createKubeconfigForPrune := func(tempDir string, content string) string {
		kubeconfigPath := filepath.Join(tempDir, "config")
		err := ioutil.WriteFile(kubeconfigPath, []byte(content), 0644)
		assert.NoError(t, err)
		return kubeconfigPath
	}

	// Test pruning with unreferenced items.
	t.Run("prune unreferenced items", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-prune-unreferenced-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		kubeconfigPath := createKubeconfigForPrune(tempDir, `
apiVersion: v1
clusters:
- cluster:
    server: https://referenced-cluster
  name: referenced-cluster
- cluster:
    server: https://unreferenced-cluster
  name: unreferenced-cluster
contexts:
- context:
    cluster: referenced-cluster
    user: referenced-user
  name: referenced-context
current-context: referenced-context
kind: Config
preferences: {}
users:
- name: referenced-user
  user:
    token: referenced-token
- name: unreferenced-user
  user:
    token: unreferenced-token
`)

		output := executeCommandC(t, "prune", "--kubeconfig", kubeconfigPath)
		expectedOutput := "Pruned 1 cluster(s) (out of 2) and 1 user(s) (out of 2) from '" + kubeconfigPath + "'."
		assert.Equal(t, expectedOutput, output)

		// Verify the kubeconfig content.
		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.Clusters, 1)
		assert.NotNil(t, config.Clusters["referenced-cluster"])
		assert.Nil(t, config.Clusters["unreferenced-cluster"])
		assert.Len(t, config.AuthInfos, 1)
		assert.NotNil(t, config.AuthInfos["referenced-user"])
		assert.Nil(t, config.AuthInfos["unreferenced-user"])
		assert.Len(t, config.Contexts, 1)
		assert.NotNil(t, config.Contexts["referenced-context"])
	})

	// Test pruning with no unreferenced items.
	t.Run("prune no unreferenced items", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-prune-no-unreferenced-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		kubeconfigPath := createKubeconfigForPrune(tempDir, `
apiVersion: v1
clusters:
- cluster:
    server: https://referenced-cluster
  name: referenced-cluster
contexts:
- context:
    cluster: referenced-cluster
    user: referenced-user
  name: referenced-context
current-context: referenced-context
kind: Config
preferences: {}
users:
- name: referenced-user
  user:
    token: referenced-token
`)

		output := executeCommandC(t, "prune", "--kubeconfig", kubeconfigPath)
		expectedOutput := "No unreferenced clusters or users found in '" + kubeconfigPath + "'. Nothing to prune."
		assert.Equal(t, expectedOutput, output)

		// Verify the kubeconfig content (should be unchanged).
		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.Clusters, 1)
		assert.Len(t, config.AuthInfos, 1)
		assert.Len(t, config.Contexts, 1)
	})

	// Test pruning an empty kubeconfig.
	t.Run("prune empty kubeconfig", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-prune-empty-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Do not create the file. loadKubeconfig should return an empty, initialized config.
		kubeconfigPath := filepath.Join(tempDir, "config")

		output := executeCommandC(t, "prune", "--kubeconfig", kubeconfigPath)
		expectedOutput := "No contexts, clusters, or users found. Nothing to prune."
		assert.Equal(t, expectedOutput, output)

		// Verify the kubeconfig content (should be empty).
		config, err := loadKubeconfig(kubeconfigPath) // Use loadKubeconfig to get an empty config
		assert.NoError(t, err)
		assert.Len(t, config.Clusters, 0)
		assert.Len(t, config.AuthInfos, 0)
		assert.Len(t, config.Contexts, 0)
	})

	// Test pruning a kubeconfig with clusters and users but no contexts.
	t.Run("prune no contexts", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-prune-no-contexts-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		kubeconfigPath := createKubeconfigForPrune(tempDir, `
apiVersion: v1
clusters:
- cluster:
    server: https://cluster1
  name: cluster1
- cluster:
    server: https://cluster2
  name: cluster2
contexts: []
kind: Config
preferences: {}
users:
- name: user1
  user:
    token: token1
- name: user2
  user:
    token: token2
`)

		output := executeCommandC(t, "prune", "--kubeconfig", kubeconfigPath)
		expectedOutput := "No contexts found. Pruned 2 cluster(s) and 2 user(s) from '" + kubeconfigPath + "'."
		assert.Equal(t, expectedOutput, output)

		// Verify the kubeconfig content.
		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.Clusters, 0)
		assert.Len(t, config.AuthInfos, 0)
		assert.Len(t, config.Contexts, 0)
	})
}
