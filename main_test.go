package main

import (
	"reflect"
	"testing"
)

func TestFindPath(t *testing.T) {
	tests := map[string]struct {
		path string
		spec Object
		want []Object
	}{
		"default": {
			path: ".paths.*.requestBody.*.schema",
			spec: Object{
				"paths": Object{
					"foo": Object{
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
			want: []Object{{
				"dave": Object{
					"lastName": "sirockin",
					"legs":     2,
				},
			}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := FindPath(tt.path, tt.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
