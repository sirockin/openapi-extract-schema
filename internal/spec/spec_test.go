package spec

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_findPath(t *testing.T) {
	tests := map[string]struct {
		path string
		spec Spec
		want []objectWithPath
	}{
		"default": {
			path: "$.paths.*.*.requestBody.*.schema",
			spec: Spec{
				object{
					"paths": object{
						"foo": object{
							"ping": object{
								"requestBody": object{
									"bar": object{
										"schema": object{
											"dave": object{
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
			want: []objectWithPath{
				{
					object: object{
						"dave": object{
							"lastName": "sirockin",
							"legs":     2,
						},
					},
					path: _path{"paths", "foo", "ping", "requestBody", "bar", "schema"},
				},
			},
		},
		"specify attribute type": {
			path: "$.components.schemas.*.[?(@type=='object')]",
			spec: Spec{
				object{
					"components": object{
						"schemas": object{
							"topLevel": object{
								"type": "object",
								"properties": object{
									"name": object{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
			want: []objectWithPath{
				{
					object: object{
						"type": "object",
						"properties": object{
							"name": object{
								"type": "string",
							},
						},
					},
					path: _path{"components", "schemas", "topLevel"},
				},
			},
		},
		"one level down specify attribute type": {
			path: "$.components.schemas.*.*.*.[?(@type=='object')]",
			spec: Spec{
				object{
					"components": object{
						"schemas": object{
							"topLevel": object{
								"type": "object",
								"properties": object{
									"name": object{
										"type": "string",
									},
									"complex": object{
										"type": "object",
										"properties": object{
											"foo": "bar",
											"ping:": object{
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
			want: []objectWithPath{
				{
					object: object{
						"type": "object",
						"properties": object{
							"foo": "bar",
							"ping:": object{
								"whizz": "bang",
							},
						},
					},
					path: _path{"components", "schemas", "topLevel", "properties", "complex"},
				},
			},
		},
		"arbitrary depth specify attribute type": {
			path: "$.components.schemas.*.*..[?(@type=='object')]",
			spec: Spec{
				object{
					"components": object{
						"schemas": object{
							"topLevel": object{
								"type": "object",
								"properties": object{
									"name": object{
										"type": "string",
									},
									"complex": object{
										"type": "object",
										"properties": object{
											"foo": "bar",
										},
									},
								},
							},
						},
					},
				},
			},
			want: []objectWithPath{
				{
					object: object{
						"type": "object",
						"properties": object{
							"foo": "bar",
						},
					},
					path: _path{"components", "schemas", "topLevel", "properties", "complex"},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.spec.findStringPath(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_groupObjects(t *testing.T) {
	tests := map[string]struct {
		in   []objectWithPath
		want []objectWithPaths
	}{
		"empty list": {
			in:   []objectWithPath{},
			want: []objectWithPaths{},
		},
		"single object": {
			in: []objectWithPath{
				{
					object: object{
						"foo": "bar",
					},
					path: _path{"a", "b", "c"},
				},
			},
			want: []objectWithPaths{
				{
					object: object{
						"foo": "bar",
					},
					paths: []_path{{"a", "b", "c"}},
				},
			},
		},
		"identical objects": {
			in: []objectWithPath{
				{
					object: object{
						"foo": "bar",
					},
					path: _path{"a", "b", "c"},
				},
				{
					object: object{
						"foo": "bar",
					},
					path: _path{"d", "e", "f"},
				},
			},
			want: []objectWithPaths{
				{
					object: object{
						"foo": "bar",
					},
					paths: []_path{
						{"a", "b", "c"},
						{"d", "e", "f"},
					},
				},
			},
		},
		"different objects": {
			in: []objectWithPath{
				{
					object: object{
						"foo": "bar",
					},
					path: _path{"a", "b", "c"},
				},
				{
					object: object{
						"foo": "ping",
					},
					path: _path{"d", "e", "f"},
				},
			},
			want: []objectWithPaths{
				{
					object: object{
						"foo": "bar",
					},
					paths: []_path{
						{"a", "b", "c"},
					},
				},
				{
					object: object{
						"foo": "ping",
					},
					paths: []_path{
						{"d", "e", "f"},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := groupObjects(tt.in); !reflect.DeepEqual(got, tt.want) {
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

func Test_capitalizeFirst(t *testing.T) {
	tests := map[string]struct {
		in   string
		want string
	}{
		"all lower case": {
			in:   "foo",
			want: "Foo",
		},
		"already title case": {
			in:   "FooBar",
			want: "FooBar",
		},
		"camel case": {
			in:   "fooBar",
			want: "FooBar",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, capitalizeFirst(tt.in))
		})
	}
}
