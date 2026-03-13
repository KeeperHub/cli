package template_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func TestDeployCmd(t *testing.T) {
	var receivedMethod, receivedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "wf-new-001",
			"name": "DeFi Monitor (Copy)",
		})
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTemplateFactory(server, ios)

	err := runTemplateViaParent(f, []string{"deploy", "tpl-001"})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, receivedMethod)
	assert.Equal(t, "/api/workflows/tpl-001/duplicate", receivedPath)
	assert.Contains(t, outBuf.String(), "wf-new-001")
}

func TestDeployCmd_CustomName(t *testing.T) {
	var requestBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requestBody = body
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "wf-new-002",
			"name": "My Custom Name",
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTemplateFactory(server, ios)

	err := runTemplateViaParent(f, []string{"deploy", "tpl-001", "--name", "My Custom Name"})
	require.NoError(t, err)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(requestBody, &body))
	assert.Equal(t, "My Custom Name", body["name"])
}

func TestDeployCmd_NoName_SendsEmptyBody(t *testing.T) {
	var requestBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := &bytes.Buffer{}
		_, _ = buf.ReadFrom(r.Body)
		requestBody = buf.Bytes()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "wf-new-003",
			"name": "Template (Copy)",
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTemplateFactory(server, ios)

	err := runTemplateViaParent(f, []string{"deploy", "tpl-001"})
	require.NoError(t, err)

	bodyStr := strings.TrimSpace(string(requestBody))
	assert.Equal(t, "{}", bodyStr, "expected empty JSON body when no --name flag")
}

func TestDeployCmd_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "wf-new-004",
			"name": "Template (Copy)",
		})
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTemplateFactory(server, ios)

	err := runTemplateViaParent(f, []string{"deploy", "tpl-001", "--json"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &result))
	assert.Equal(t, "wf-new-004", result["id"])
}
