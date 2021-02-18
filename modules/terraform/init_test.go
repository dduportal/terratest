package terraform

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitBackendConfig(t *testing.T) {
	t.Parallel()

	stateDirectory, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	remoteStateFile := filepath.Join(stateDirectory, "backend.tfstate")

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-backend", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
		BackendConfig: map[string]interface{}{
			"path": remoteStateFile,
		},
	}

	InitAndApply(t, options)

	assert.FileExists(t, remoteStateFile)
}

func TestInitPluginDir(t *testing.T) {
	t.Parallel()

	pluginDir, err := ioutil.TempDir("", t.Name())
	require.NoError(t, err)
	defer os.RemoveAll(pluginDir)

	terraformFixture := "../../test/fixtures/terraform-basic-configuration"

	initializePluginsFolder, err := files.CopyTerraformFolderToTemp(terraformFixture, t.Name())
	require.NoError(t, err)
	defer os.RemoveAll(initializePluginsFolder)

	testFolder, err := files.CopyTerraformFolderToTemp(terraformFixture, t.Name())
	require.NoError(t, err)
	defer os.RemoveAll(testFolder)

	terraformOptions := &Options{
		TerraformDir: initializePluginsFolder,
	}

	terraformOptionsPluginDir := &Options{
		TerraformDir: testFolder,
		PluginDir:    pluginDir,
	}

	Init(t, terraformOptions)

	_, err = InitE(t, terraformOptionsPluginDir)
	require.Error(t, err)

	initializedPluginDir := initializePluginsFolder + "/.terraform/plugins"
	files.CopyFolderContents(initializedPluginDir, pluginDir)

	initOutput := Init(t, terraformOptionsPluginDir)

	assert.Contains(t, initOutput, "(unauthenticated)")
}
