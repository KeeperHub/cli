package khhttp_test

import (
	"testing"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/stretchr/testify/assert"
)

func TestBuildBaseURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bare hostname adds https scheme",
			input: "localhost:3000",
			want:  "https://localhost:3000",
		},
		{
			name:  "http scheme preserved unchanged",
			input: "http://127.0.0.1:54321",
			want:  "http://127.0.0.1:54321",
		},
		{
			name:  "https scheme preserved unchanged",
			input: "https://app.keeperhub.io",
			want:  "https://app.keeperhub.io",
		},
		{
			name:  "bare domain adds https scheme",
			input: "app.keeperhub.io",
			want:  "https://app.keeperhub.io",
		},
		{
			name:  "bare hostname trailing slash stripped",
			input: "localhost:3000/",
			want:  "https://localhost:3000",
		},
		{
			name:  "http scheme with trailing slash stripped",
			input: "http://localhost:3000/",
			want:  "http://localhost:3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := khhttp.BuildBaseURL(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
