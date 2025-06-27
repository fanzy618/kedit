
package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd/api"
)

// TestLoadKubeconfig_FileExists tests loading a valid kubeconfig file.
func TestLoadKubeconfig_FileExists(t *testing.T) {
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
    server: https://test-cluster
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
kind: Config
preferences: {}
users:
- name: test-user
  user:
    token: test-token
`), 0644)
	assert.NoError(t, err)

	// Load the kubeconfig.
	config, err := loadKubeconfig(kubeconfigPath)
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Assert that the loaded config is correct.
	assert.Equal(t, "test-context", config.CurrentContext)
	assert.Len(t, config.Clusters, 1)
	assert.Len(t, config.Contexts, 1)
	assert.Len(t, config.AuthInfos, 1)
	assert.Equal(t, "https://test-cluster", config.Clusters["test-cluster"].Server)
}

// TestLoadKubeconfig_FileDoesNotExist tests loading a non-existent kubeconfig file.
func TestLoadKubeconfig_FileDoesNotExist(t *testing.T) {
	// Create a temporary directory for the test.
	tempDir, err := ioutil.TempDir("", "kedit-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Attempt to load a non-existent kubeconfig file.
	kubeconfigPath := filepath.Join(tempDir, "non-existent-config")
	config, err := loadKubeconfig(kubeconfigPath)
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Assert that an empty config is returned.
	assert.Empty(t, config.CurrentContext)
	assert.Empty(t, config.Clusters)
	assert.Empty(t, config.Contexts)
	assert.Empty(t, config.AuthInfos)
}

// TestSaveKubeconfig tests saving a kubeconfig file.
func TestSaveKubeconfig(t *testing.T) {
	// Create a temporary directory for the test.
	tempDir, err := ioutil.TempDir("", "kedit-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new kubeconfig.
	config := api.NewConfig()
	config.CurrentContext = "test-context"
	config.Clusters["test-cluster"] = &api.Cluster{Server: "https://test-cluster"}
	config.Contexts["test-context"] = &api.Context{Cluster: "test-cluster", AuthInfo: "test-user"}
	config.AuthInfos["test-user"] = &api.AuthInfo{Token: "test-token"}

	// Save the kubeconfig.
	kubeconfigPath := filepath.Join(tempDir, "config")
	err = saveKubeconfig(config, kubeconfigPath)
	assert.NoError(t, err)

	// Read the saved file.
	savedConfig, err := ioutil.ReadFile(kubeconfigPath)
	assert.NoError(t, err)

	// Assert that the saved config is correct.
	expectedConfig := `
apiVersion: v1
clusters:
- cluster:
    server: https://test-cluster
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
kind: Config
preferences: {}
users:
- name: test-user
  user:
    token: test-token
`
	assert.YAMLEq(t, expectedConfig, string(savedConfig))
}
