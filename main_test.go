package main

import (
	"reflect"
	"testing"
)

func TestFindPath(t *testing.T) {
	tests := map[string]struct {
		path string
		spec Spec
		want []objectPath
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
			want: []objectPath{
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

func Test_generateSchemaNameFromRequest(t *testing.T) {
	tests := map[string]struct {
		in   Path
		want string
	}{
		"default": {
			in:   Path{"paths", "/v2/foo", "POST"},
			want: "postV2FooRequest",
		},
	}
	for k, tt := range tests {
		t.Run(k, func(t *testing.T) {
			if got := tt.in.generateSchemaNameFromRequest(); got != tt.want {
				t.Errorf("sanitizeURLPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
