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
