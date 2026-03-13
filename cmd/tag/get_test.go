package tag_test

import (
	"testing"

	"github.com/keeperhub/cli/cmd/tag"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCmd(t *testing.T) {
	tags := []map[string]interface{}{
		{
			"id":            "tag-001",
			"name":          "DeFi",
			"color":         "#6366f1",
			"workflowCount": 3,
			"createdAt":     "2026-01-01T00:00:00Z",
			"updatedAt":     "2026-02-01T00:00:00Z",
		},
	}
	server := makeTagsServer(t, tags)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"get", "tag-001"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "tag-001")
	assert.Contains(t, out, "DeFi")
}

func TestGetCmd_NotFound(t *testing.T) {
	tags := []map[string]interface{}{
		{"id": "tag-001", "name": "DeFi", "color": "#6366f1", "workflowCount": 0},
	}
	server := makeTagsServer(t, tags)
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"get", "tag-999"})
	err := cmd.Execute()
	require.Error(t, err)

	var notFound cmdutil.NotFoundError
	assert.ErrorAs(t, err, &notFound)
}

func TestGetCmd_JSON(t *testing.T) {
	tags := []map[string]interface{}{
		{
			"id":            "tag-001",
			"name":          "DeFi",
			"color":         "#6366f1",
			"workflowCount": 2,
			"createdAt":     "2026-01-01T00:00:00Z",
			"updatedAt":     "2026-01-01T00:00:00Z",
		},
	}
	server := makeTagsServer(t, tags)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"get", "tag-001", "--json"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, `"id"`)
	assert.Contains(t, out, "tag-001")
}
