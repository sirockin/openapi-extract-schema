package spec

import (
	"reflect"
	"testing"
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
