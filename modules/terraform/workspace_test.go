package terraform

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/stretchr/testify/assert"
)

func TestWorkspaceNew(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-workspace", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	out := WorkspaceSelectOrNew(t, options, "terratest")

	assert.Equal(t, "terratest", out)
}

func TestWorkspaceIllegalName(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-workspace", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	out, err := WorkspaceSelectOrNewE(t, options, "###@@@&&&")

	assert.Error(t, err)
	assert.Equal(t, "", out, "%q should be an empty string", out)
}

func TestWorkspaceSelect(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-workspace", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	out := WorkspaceSelectOrNew(t, options, "terratest")
	assert.Equal(t, "terratest", out)

	out = WorkspaceSelectOrNew(t, options, "default")
	assert.Equal(t, "default", out)
}

func TestWorkspaceApply(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-workspace", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	WorkspaceSelectOrNew(t, options, "Terratest")
	out := InitAndApply(t, options)

	assert.Contains(t, out, "Hello, Terratest")
}

func TestIsExistingWorkspace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		out      string
		name     string
		expected bool
	}{
		{"  default\n* foo\n", "default", true},
		{"* default\n  foo\n", "default", true},
		{"  foo\n* default\n", "default", true},
		{"* foo\n  default\n", "default", true},
		{"  foo\n* bar\n", "default", false},
		{"* foo\n  bar\n", "default", false},
		{"  default\n* foobar\n", "foo", false},
		{"* default\n  foobar\n", "foo", false},
		{"  default\n* foo\n", "foobar", false},
		{"* default\n  foo\n", "foobar", false},
		{"* default\n  foo\n", "foo", true},
	}

	for _, testCase := range testCases {
		actual := isExistingWorkspace(testCase.out, testCase.name)
		assert.Equal(t, testCase.expected, actual, "Out: %q, Name: %q", testCase.out, testCase.name)
	}
}

func TestNameMatchesWorkspace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		workspace string
		expected  bool
	}{
		{"default", "  default", true},
		{"default", "* default", true},
		{"default", "", false},
		{"foo", "  foobar", false},
		{"foo", "* foobar", false},
		{"foobar", "  foo", false},
		{"foobar", "* foo", false},
		{"foo", "  foo", true},
		{"foo", "* foo", true},
	}

	for _, testCase := range testCases {
		actual := nameMatchesWorkspace(testCase.name, testCase.workspace)
		assert.Equal(t, testCase.expected, actual, "Name: %q, Workspace: %q", testCase.name, testCase.workspace)
	}
}

// Please note that this test depends on other functions that should be mocked to be a unit test.
func TestWorkspaceDeleteE(t *testing.T) {
	t.Parallel()

	// state describes an expected status when a given testCase begins
	type state struct {
		workspaces []string
		current    string
	}

	// testCase describes a named test case with a state, args and expcted results
	type testCase struct {
		name                 string
		initialState         state
		toDeleteWorkspace    string
		expectedCurrent      string
		expectedErrorMessage string
	}

	testCases := []testCase{
		{
			name: "delete another existing workspace and stay on current",
			initialState: state{
				workspaces: []string{"staging", "production"},
				current:    "staging",
			},
			toDeleteWorkspace:    "production",
			expectedCurrent:      "staging",
			expectedErrorMessage: "",
		},
		{
			name: "delete current workspace and switch to a specified",
			initialState: state{
				workspaces: []string{"staging", "production"},
				current:    "production",
			},
			toDeleteWorkspace:    "production",
			expectedCurrent:      "default",
			expectedErrorMessage: "",
		},
		{
			name: "delete a non existing workspace should trigger an error",
			initialState: state{
				workspaces: []string{"staging", "production"},
				current:    "staging",
			},
			toDeleteWorkspace:    "hellothere",
			expectedCurrent:      "staging",
			expectedErrorMessage: "The workspace \"hellothere\" does not exist.",
		},
		{
			name: "delete the default workspace triggers an error",
			initialState: state{
				workspaces: []string{"staging", "production"},
				current:    "staging",
			},
			toDeleteWorkspace:    "default",
			expectedCurrent:      "staging",
			expectedErrorMessage: "Deleting the workspace 'default' is not supported",
		},
	}

	for _, tt := range testCases {
		testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-workspace", tt.name)
		if err != nil {
			t.Fatal(err)
		}

		options := &Options{
			TerraformDir: testFolder,
		}

		// Set up pre-existing environment based on test case description
		for _, existingWorkspace := range tt.initialState.workspaces {
			_, err = RunTerraformCommandE(t, options, "workspace", "new", existingWorkspace)
			if err != nil {
				t.Fatal(err)
			}
		}
		// Switch to the specified workspace
		_, err = RunTerraformCommandE(t, options, "workspace", "select", tt.initialState.current)
		if err != nil {
			t.Fatal(err)
		}

		// Testing time, wooohoooo
		gotResult, gotErr := WorkspaceDeleteE(t, options, tt.toDeleteWorkspace)

		// Check for errors
		if tt.expectedErrorMessage != "" {
			assert.Error(t, gotErr)
			assert.Equal(t, tt.expectedErrorMessage, gotErr.Error())
		} else {
			assert.Nil(t, gotErr)
			// Check for results
			assert.Equal(t, tt.expectedCurrent, gotResult)
			assert.False(t, isExistingWorkspace(RunTerraformCommand(t, options, "workspace", "list"), tt.toDeleteWorkspace))
		}

	}
}

// Please note that this test depends on other functions that should be mocked to be a unit test.
func TestWorkspaceDelete(t *testing.T) {
	t.Parallel()

	// state describes an expected status when a given testCase begins
	type state struct {
		workspaces []string
		current    string
	}

	// testCase describes a named test case with a state, args and expcted results
	type testCase struct {
		name              string
		initialState      state
		toDeleteWorkspace string
		expectedCurrent   string
	}

	testCases := []testCase{
		{
			name: "delete another existing workspace and stay on current",
			initialState: state{
				workspaces: []string{"staging", "production"},
				current:    "staging",
			},
			toDeleteWorkspace: "production",
			expectedCurrent:   "staging",
		},
		{
			name: "delete current workspace and switch to a specified",
			initialState: state{
				workspaces: []string{"staging", "production"},
				current:    "production",
			},
			toDeleteWorkspace: "production",
			expectedCurrent:   "default",
		},
	}

	for _, tt := range testCases {
		testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-workspace", tt.name)
		if err != nil {
			t.Fatal(err)
		}

		options := &Options{
			TerraformDir: testFolder,
		}

		// Set up pre-existing environment based on test case description
		for _, existingWorkspace := range tt.initialState.workspaces {
			_, err = RunTerraformCommandE(t, options, "workspace", "new", existingWorkspace)
			if err != nil {
				t.Fatal(err)
			}
		}
		// Switch to the specified workspace
		_, err = RunTerraformCommandE(t, options, "workspace", "select", tt.initialState.current)
		if err != nil {
			t.Fatal(err)
		}

		// Testing time, wooohoooo
		gotResult := WorkspaceDelete(t, options, tt.toDeleteWorkspace)

		// Check for results
		assert.Equal(t, tt.expectedCurrent, gotResult)
		assert.False(t, isExistingWorkspace(RunTerraformCommand(t, options, "workspace", "list"), tt.toDeleteWorkspace))

	}
}
