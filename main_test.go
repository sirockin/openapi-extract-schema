package main

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
			path: ".paths.*.*.requestBody.*.schema",
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
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.spec.FindPath(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_generateSchemaNameFromRequest(t *testing.T) {
// 	tests := map[string]struct {
// 		in   Path
// 		want string
// 	}{
// 		"default": {
// 			in:   Path{"paths", "/v2/foo", "POST"},
// 			want: "postV2FooRequest",
// 		},
// 	}
// 	for k, tt := range tests {
// 		t.Run(k, func(t *testing.T) {
// 			if got := tt.in.generateSchemaNameFromRequest(); got != tt.want {
// 				t.Errorf("sanitizeURLPath() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

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
		in string
		want string
	}{
		"no suffix":
		{
			in: "Foo",
			want: "Foo2",
		},
		"suffix 1":
		{
			in: "Foo1",
			want: "Foo2",
		},
		"suffix another number":
		{
			in: "Foo999",
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
