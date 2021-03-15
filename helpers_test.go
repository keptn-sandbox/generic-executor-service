package main

import "testing"

func Test_replacePlaceHolderRecursively(t *testing.T) {
	type args struct {
		input   string
		keyPath string
		values  map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "replace placeholders",
			args: args{
				input:   "this should contain ${data.project} and ${data.my.property}",
				keyPath: "",
				values: map[string]interface{}{
					"data": map[string]interface{}{
						"project": "my-project",
						"my": map[string]interface{}{
							"property": "some-property",
						},
					},
				},
			},
			want: "this should contain my-project and some-property",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replacePlaceHolderRecursively(tt.args.input, tt.args.keyPath, tt.args.values); got != tt.want {
				t.Errorf("replacePlaceHolderRecursively() = %v, want %v", got, tt.want)
			}
		})
	}
}
