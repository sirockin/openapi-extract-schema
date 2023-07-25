package main

import (
	"reflect"
	"testing"
)

func TestFindPath(t *testing.T) {
	tests := map[string]struct {
		path string
		spec Object
		want []objectPath
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
			want: []objectPath{
				{
					object: Object{
						"dave": Object{
							"lastName": "sirockin",
							"legs":     2,
						},
					},
					path: Path{"paths", "foo", "requestBody", "bar", "schema" },
				},
			},
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
