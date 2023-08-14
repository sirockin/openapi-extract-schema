package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectWithPaths_requestSymbol(t *testing.T) {
	// ".paths.*.*.requestBody.content.*.schema"
	tests := map[string]struct {
		paths Paths
		want  string
	}{
		"single request body": {
			paths: []Path{
				{"paths", "/v2/foo", "POST", "requestBody", "content", "application/json", "schema"},
			},
			want: "PostV2FooRequest",
		},
		"single request with dash in endpoing": {
			paths: []Path{
				{"paths", "/v2/mandate-imports", "POST", "requestBody", "content", "application/json", "schema"},
			},
			want: "PostV2MandateImportsRequest",
		},
		"common endpoint": {
			paths: []Path{
				{"paths", "/v2/foo", "POST", "requestBody", "content", "application/json", "schema"},
				{"paths", "/v2/foo", "GET", "requestBody", "content", "application/json", "schema"},
			},
			want: "CommonV2FooRequest",
		},
		"common verb": {
			paths: []Path{
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

func TestObjectWithPaths_responseSymbol(t *testing.T) {
	// .paths.*.*.responses.*.content.*.schema	tests := map[string]struct {
	tests := map[string]struct {
		paths Paths
		want  string
	}{
		"single response": {
			paths: []Path{
				{"paths", "/v2/foo", "POST", "responses", "200"},
			},
			want: "PostV2Foo200Response",
		},
		"single response with dash in endpoint": {
			paths: []Path{
				{"paths", "/v2/mandate-imports", "POST", "responses", "200"},
			},
			want: "PostV2MandateImports200Response",
		},
		"all except verb the same": {
			paths: []Path{
				{"paths", "/v2/foo", "POST", "responses", "200"},
				{"paths", "/v2/foo", "GET", "responses", "200"},
			},
			want: "CommonV2Foo200Response",
		},
		"all except endpoint the same": {
			paths: []Path{
				{"paths", "/v2/foo", "POST", "responses", "200"},
				{"paths", "/v2/ping", "POST", "responses", "200"},
			},
			want: "CommonPost200Response",
		},
		"all except response code suffix the same": {
			paths: []Path{
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
