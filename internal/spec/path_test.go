package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaths_requestSymbol(t *testing.T) {
	tests := map[string]struct {
		paths paths
		want  string
	}{
		"single request body": {
			paths: []_path{
				{"paths", "/v2/foo", "POST", "requestBody", "content", "application/json", "schema"},
			},
			want: "PostV2FooRequest",
		},
		"single request with dash in endpoint": {
			paths: []_path{
				{"paths", "/v2/mandate-imports", "POST", "requestBody", "content", "application/json", "schema"},
			},
			want: "PostV2MandateImportsRequest",
		},
		"common endpoint": {
			paths: []_path{
				{"paths", "/v2/foo", "POST", "requestBody", "content", "application/json", "schema"},
				{"paths", "/v2/foo", "GET", "requestBody", "content", "application/json", "schema"},
			},
			want: "CommonV2FooRequest",
		},
		"common verb": {
			paths: []_path{
				{"paths", "/v2/foo", "POST", "requestBody", "content", "application/json", "schema"},
				{"paths", "/v2/ping", "POST", "requestBody", "content", "application/json", "schema"},
			},
			want: "CommonPostRequest",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.paths.requestSymbol()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPaths_responseSymbol(t *testing.T) {
	tests := map[string]struct {
		paths paths
		want  string
	}{
		"single response": {
			paths: []_path{
				{"paths", "/v2/foo", "POST", "responses", "200"},
			},
			want: "PostV2Foo200Response",
		},
		"single response with dash in endpoint": {
			paths: []_path{
				{"paths", "/v2/mandate-imports", "POST", "responses", "200"},
			},
			want: "PostV2MandateImports200Response",
		},
		"all except verb the same": {
			paths: []_path{
				{"paths", "/v2/foo", "POST", "responses", "200"},
				{"paths", "/v2/foo", "GET", "responses", "200"},
			},
			want: "CommonV2Foo200Response",
		},
		"all except endpoint the same": {
			paths: []_path{
				{"paths", "/v2/foo", "POST", "responses", "200"},
				{"paths", "/v2/ping", "POST", "responses", "200"},
			},
			want: "CommonPost200Response",
		},
		"all except response code suffix the same": {
			paths: []_path{
				{"paths", "/v2/ping", "POST", "responses", "201"},
				{"paths", "/v2/ping", "POST", "responses", "200"},
			},
			want: "PostV2Ping2xxResponse",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.paths.responseSymbol()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
