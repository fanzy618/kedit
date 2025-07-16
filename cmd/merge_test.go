package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd"
)



func TestMergeCommand(t *testing.T) {
	// Helper to create a base kubeconfig for target.
	createTargetKubeconfig := func(tempDir string) string {
		targetPath := filepath.Join(tempDir, "target-config")
		err := ioutil.WriteFile(targetPath, []byte(`
apiVersion: v1
clusters:
- cluster:
    server: https://target-cluster
  name: target-cluster
contexts:
- context:
    cluster: target-cluster
    user: target-user
  name: target-context
current-context: target-context
kind: Config
preferences: {}
users:
- name: target-user
  user:
    token: target-token
`), 0644)
		assert.NoError(t, err)
		return targetPath
	}

	// Helper to create a source kubeconfig.
	createSourceKubeconfig := func(tempDir string, content string) string {
		sourcePath := filepath.Join(tempDir, "source-config")
		err := ioutil.WriteFile(sourcePath, []byte(content), 0644)
		assert.NoError(t, err)
		return sourcePath
	}

	// Test merging a new context.
	t.Run("merge new context", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-merge-new-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		targetKubeconfigPath := createTargetKubeconfig(tempDir)
		sourceKubeconfigPath := createSourceKubeconfig(tempDir, `
apiVersion: v1
clusters:
- cluster:
    server: https://new-cluster
  name: new-cluster
contexts:
- context:
    cluster: new-cluster
    user: new-user
  name: new-context
current-context: new-context
kind: Config
preferences: {}
users:
- name: new-user
  user:
    token: new-token
`)

		output := executeCommandC(t, "merge", "new-context", "--from", sourceKubeconfigPath, "--kubeconfig", targetKubeconfigPath)
		expectedOutput := "Successfully merged context 'new-context' (with cluster 'new-cluster' and user 'new-user') from '" + sourceKubeconfigPath + "' into '" + targetKubeconfigPath + "'."
		assert.Equal(t, expectedOutput, output)

		// Verify the merged content.
		config, err := clientcmd.LoadFromFile(targetKubeconfigPath)
		assert.NoError(t, err)
		assert.NotNil(t, config.Contexts["new-context"])
		assert.NotNil(t, config.Clusters["new-cluster"])
		assert.NotNil(t, config.AuthInfos["new-user"])
		assert.Len(t, config.Contexts, 2)
		assert.Len(t, config.Clusters, 2)
		assert.Len(t, config.AuthInfos, 2)
	})

	// Test overwriting an existing context.
	t.Run("merge overwrite existing context", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-merge-overwrite-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		targetKubeconfigPath := createTargetKubeconfig(tempDir)
		sourceKubeconfigPath := createSourceKubeconfig(tempDir, `
apiVersion: v1
clusters:
- cluster:
    server: https://overwritten-cluster
  name: target-cluster
contexts:
- context:
    cluster: target-cluster
    user: target-user
  name: target-context
current-context: target-context
kind: Config
preferences: {}
users:
- name: target-user
  user:
    token: overwritten-token
`)

		output := executeCommandC(t, "merge", "target-context", "--from", sourceKubeconfigPath, "--kubeconfig", targetKubeconfigPath)
		expectedOutput := "Successfully merged context 'target-context' (with cluster 'target-cluster' and user 'target-user') from '" + sourceKubeconfigPath + "' into '" + targetKubeconfigPath + "'."
		assert.Equal(t, expectedOutput, output)

		// Verify the overwritten content.
		config, err := clientcmd.LoadFromFile(targetKubeconfigPath)
		assert.NoError(t, err)
		assert.Equal(t, "https://overwritten-cluster", config.Clusters["target-cluster"].Server)
		assert.Equal(t, "overwritten-token", config.AuthInfos["target-user"].Token)
	})

	// Test merging a context with non-existent cluster in source.
	t.Run("merge context with non-existent cluster", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-merge-no-cluster-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		targetKubeconfigPath := createTargetKubeconfig(tempDir)
		sourceKubeconfigPath := createSourceKubeconfig(tempDir, `
apiVersion: v1
clusters: []
contexts:
- context:
    cluster: non-existent-cluster
    user: new-user
  name: new-context
current-context: new-context
kind: Config
preferences: {}
users:
- name: new-user
  user:
    token: new-token
`)

		output := executeCommandC(t, "merge", "new-context", "--from", sourceKubeconfigPath, "--kubeconfig", targetKubeconfigPath)
		assert.Contains(t, output, "Error: cluster 'non-existent-cluster' (referenced by context 'new-context') not found in source kubeconfig '")
		// Ensure no changes were made to the target.
		config, err := clientcmd.LoadFromFile(targetKubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.Contexts, 1)
		assert.Len(t, config.Clusters, 1)
		assert.Len(t, config.AuthInfos, 1)
	})

	// Test merging a context with non-existent user in source.
	t.Run("merge context with non-existent user", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-merge-no-user-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		targetKubeconfigPath := createTargetKubeconfig(tempDir)
		sourceKubeconfigPath := createSourceKubeconfig(tempDir, `
apiVersion: v1
clusters:
- cluster:
    server: https://new-cluster
  name: new-cluster
contexts:
- context:
    cluster: new-cluster
    user: non-existent-user
  name: new-context
current-context: new-context
kind: Config
preferences: {}
users: []
`)

		output := executeCommandC(t, "merge", "new-context", "--from", sourceKubeconfigPath, "--kubeconfig", targetKubeconfigPath)
		assert.Contains(t, output, "Error: user 'non-existent-user' (referenced by context 'new-context') not found in source kubeconfig '")
		// Ensure no changes were made to the target.
		config, err := clientcmd.LoadFromFile(targetKubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.Contexts, 1)
		assert.Len(t, config.Clusters, 1)
		assert.Len(t, config.AuthInfos, 1)
	})

	// Test merging from a non-existent source kubeconfig.
	t.Run("merge from non-existent source", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-merge-no-source-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		targetKubeconfigPath := createTargetKubeconfig(tempDir)
		nonExistentSourcePath := filepath.Join(tempDir, "non-existent-source.yaml")

		output := executeCommandC(t, "merge", "any-context", "--from", nonExistentSourcePath, "--kubeconfig", targetKubeconfigPath)
		assert.Contains(t, output, "Error: source kubeconfig file '" + nonExistentSourcePath + "' not found")
		// Ensure no changes were made to the target.
		config, err := clientcmd.LoadFromFile(targetKubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.Contexts, 1)
		assert.Len(t, config.Clusters, 1)
		assert.Len(t, config.AuthInfos, 1)
	})

	// Test merging a context that does not exist in the source kubeconfig.
	t.Run("merge non-existent context in source", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-merge-source-no-context-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		targetKubeconfigPath := createTargetKubeconfig(tempDir)
		sourceKubeconfigPath := createSourceKubeconfig(tempDir, `
apiVersion: v1
clusters:
- cluster:
    server: https://some-cluster
  name: some-cluster
contexts: []
kind: Config
preferences: {}
users:
- name: some-user
  user:
    token: some-token
`)

		output := executeCommandC(t, "merge", "non-existent-context", "--from", sourceKubeconfigPath, "--kubeconfig", targetKubeconfigPath)
		assert.Contains(t, output, "Error: context 'non-existent-context' not found in source kubeconfig '")
		// Ensure no changes were made to the target.
		config, err := clientcmd.LoadFromFile(targetKubeconfigPath)
		assert.NoError(t, err)
		assert.Len(t, config.Contexts, 1)
		assert.Len(t, config.Clusters, 1)
		assert.Len(t, config.AuthInfos, 1)
	})

	// Test merging with a new name.
	t.Run("merge with new name", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "kedit-test-merge-new-name-")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		targetKubeconfigPath := createTargetKubeconfig(tempDir)
		sourceKubeconfigPath := createSourceKubeconfig(tempDir, `
apiVersion: v1
clusters:
- cluster:
    server: https://new-cluster
  name: new-cluster
contexts:
- context:
    cluster: new-cluster
    user: new-user
  name: new-context
current-context: new-context
kind: Config
preferences: {}
users:
- name: new-user
  user:
    token: new-token
`)

		output := executeCommandC(t, "merge", "new-context", "--from", sourceKubeconfigPath, "--name", "renamed", "--kubeconfig", targetKubeconfigPath)
		expectedOutput := "Successfully merged context 'renamed' (with cluster 'renamed' and user 'renamed') from '" + sourceKubeconfigPath + "' into '" + targetKubeconfigPath + "'."
		assert.Equal(t, expectedOutput, output)

		// Verify the merged and renamed content.
		config, err := clientcmd.LoadFromFile(targetKubeconfigPath)
		assert.NoError(t, err)
		assert.NotNil(t, config.Contexts["renamed"])
		assert.NotNil(t, config.Clusters["renamed"])
		assert.NotNil(t, config.AuthInfos["renamed"])
		assert.Len(t, config.Contexts, 2)
		assert.Len(t, config.Clusters, 2)
		assert.Len(t, config.AuthInfos, 2)
		assert.Equal(t, "renamed", config.Contexts["renamed"].Cluster)
		assert.Equal(t, "renamed", config.Contexts["renamed"].AuthInfo)
	})
}