package terraform

import (
	"io/ioutil"
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

	terraformFixture := "../../test/fixtures/terraform-basic-configuration"

	initializePluginsFolder, err := files.CopyTerraformFolderToTemp(terraformFixture, t.Name())
	require.NoError(t, err)

	testFolder, err := files.CopyTerraformFolderToTemp(terraformFixture, t.Name())
	require.NoError(t, err)

	terraformOptions := &Options{
		TerraformDir: initializePluginsFolder,
	}

	Init(t, terraformOptions)

	initializedPluginDir := initializePluginsFolder + "/.terraform/plugins"
	files.CopyFolderContents(initializedPluginDir, pluginDir)

	terraformOptionsPluginDir := &Options{
		TerraformDir: testFolder,
		PluginDir:    pluginDir,
	}

	initOutput := Init(t, terraformOptionsPluginDir)

	assert.Contains(t, initOutput, "(unauthenticated)")
}
