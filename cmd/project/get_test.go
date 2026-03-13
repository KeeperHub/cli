package project_test

import (
	"testing"

	"github.com/keeperhub/cli/cmd/project"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCmd(t *testing.T) {
	projects := []map[string]interface{}{
		{
			"id":            "proj-001",
			"name":          "My Project",
			"description":   "A test project",
			"color":         "#6366f1",
			"workflowCount": 3,
			"createdAt":     "2026-01-01T00:00:00Z",
			"updatedAt":     "2026-02-01T00:00:00Z",
		},
	}
	server := makeProjectsServer(t, projects)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newProjectFactory(server, ios)

	cmd := project.NewProjectCmd(f)
	cmd.SetArgs([]string{"get", "proj-001"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "proj-001")
	assert.Contains(t, out, "My Project")
}

func TestGetCmd_NotFound(t *testing.T) {
	projects := []map[string]interface{}{
		{"id": "proj-001", "name": "My Project", "description": "", "workflowCount": 0},
	}
	server := makeProjectsServer(t, projects)
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newProjectFactory(server, ios)

	cmd := project.NewProjectCmd(f)
	cmd.SetArgs([]string{"get", "proj-999"})
	err := cmd.Execute()
	require.Error(t, err)

	var notFound cmdutil.NotFoundError
	assert.ErrorAs(t, err, &notFound)
}

func TestGetCmd_JSON(t *testing.T) {
	projects := []map[string]interface{}{
		{
			"id":            "proj-001",
			"name":          "My Project",
			"description":   "desc",
			"color":         "#6366f1",
			"workflowCount": 2,
			"createdAt":     "2026-01-01T00:00:00Z",
			"updatedAt":     "2026-01-01T00:00:00Z",
		},
	}
	server := makeProjectsServer(t, projects)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newProjectFactory(server, ios)

	cmd := project.NewProjectCmd(f)
	cmd.SetArgs([]string{"get", "proj-001", "--json"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, `"id"`)
	assert.Contains(t, out, "proj-001")
}
