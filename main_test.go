package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindPath(t *testing.T) {
	tests := map[string]struct {
		path string
		spec Spec
		want []ObjectWithPath
	}{
		"default": {
			path: "$.paths.*.*.requestBody.*.schema",
			spec: Spec{
				Object{
					"paths": Object{
						"foo": Object{
							"ping": Object{
								"requestBody": Object{
									"bar": Object{
										"schema": Object{
											"dave": Object{
												"lastName": "sirockin",
												"legs":     2,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []ObjectWithPath{
				{
					object: Object{
						"dave": Object{
							"lastName": "sirockin",
							"legs":     2,
						},
					},
					path: Path{"paths", "foo", "ping", "requestBody", "bar", "schema"},
				},
			},
		},
		"specify attribute type": {
			path: "$.components.schemas.*.[?(@type=='object')]",
			spec: Spec{
				Object{
					"components": Object{
						"schemas": Object{
							"topLevel": Object{
								"type": "object",
								"properties": Object{
									"name": Object{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
			want: []ObjectWithPath{
				{
					object: Object{
						"type": "object",
						"properties": Object{
							"name": Object{
								"type": "string",
							},
						},
					},
					path: Path{"components", "schemas", "topLevel"},
				},
			},
		},
		"one level down specify attribute type": {
			path: "$.components.schemas.*.*.*.[?(@type=='object')]",
			spec: Spec{
				Object{
					"components": Object{
						"schemas": Object{
							"topLevel": Object{
								"type": "object",
								"properties": Object{
									"name": Object{
										"type": "string",
									},
									"complex": Object{
										"type": "object",
										"properties": Object{
											"foo": "bar",
											"ping:": Object{
												"whizz": "bang",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []ObjectWithPath{
				{
					object: Object{
						"type": "object",
						"properties": Object{
							"foo": "bar",
							"ping:": Object{
								"whizz": "bang",
							},
						},
					},
					path: Path{"components", "schemas", "topLevel", "properties", "complex"},
				},
			},
		},	
		"arbitrary depth specify attribute type": {
			path: "$.components.schemas.*.*..[?(@type=='object')]",
			spec: Spec{
				Object{
					"components": Object{
						"schemas": Object{
							"topLevel": Object{
								"type": "object",
								"properties": Object{
									"name": Object{
										"type": "string",
									},
									"complex": Object{
										"type": "object",
										"properties": Object{
											"foo": "bar",
										},
									},
								},
							},
						},
					},
				},
			},
			want: []ObjectWithPath{
				{
					object: Object{
						"type": "object",
						"properties": Object{
							"foo": "bar",
						},
					},
					path: Path{"components", "schemas", "topLevel", "properties", "complex"},
				},
			},
		},				
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.spec.FindPath(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroupObjects(t *testing.T) {
	tests := map[string]struct {
		in   []ObjectWithPath
		want []ObjectWithPaths
	}{
		"empty list": {
			in:   []ObjectWithPath{},
			want: []ObjectWithPaths{},
		},
		"single object": {
			in: []ObjectWithPath{
				{
					object: Object{
						"foo": "bar",
					},
					path: Path{"a", "b", "c"},
				},
			},
			want: []ObjectWithPaths{
				{
					object: Object{
						"foo": "bar",
					},
					paths: []Path{{"a", "b", "c"}},
				},
			},
		},
		"identical objects": {
			in: []ObjectWithPath{
				{
					object: Object{
						"foo": "bar",
					},
					path: Path{"a", "b", "c"},
				},
				{
					object: Object{
						"foo": "bar",
					},
					path: Path{"d", "e", "f"},
				},
			},
			want: []ObjectWithPaths{
				{
					object: Object{
						"foo": "bar",
					},
					paths: []Path{
						{"a", "b", "c"},
						{"d", "e", "f"},
					},
				},
			},
		},
		"different objects": {
			in: []ObjectWithPath{
				{
					object: Object{
						"foo": "bar",
					},
					path: Path{"a", "b", "c"},
				},
				{
					object: Object{
						"foo": "ping",
					},
					path: Path{"d", "e", "f"},
				},
			},
			want: []ObjectWithPaths{
				{
					object: Object{
						"foo": "bar",
					},
					paths: []Path{
						{"a", "b", "c"},
					},
				},
				{
					object: Object{
						"foo": "ping",
					},
					paths: []Path{
						{"d", "e", "f"},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := GroupObjects(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GroupObjects() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nextSymbol(t *testing.T) {
	tests := map[string]struct {
		in   string
		want string
	}{
		"no suffix": {
			in:   "Foo",
			want: "Foo2",
		},
		"suffix 1": {
			in:   "Foo1",
			want: "Foo2",
		},
		"suffix another number": {
			in:   "Foo999",
			want: "Foo1000",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := nextSymbol(tt.in); got != tt.want {
				t.Errorf("nextSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
