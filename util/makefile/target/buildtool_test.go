package target

import (
	"fmt"
	"github.com/peter-mount/go-build/util/makefile"
	"reflect"
	"testing"
)

func Test_buildTool_Build(t *testing.T) {

	got := New().
		Target("builds/test").
		BuildTool("-test", "arg1", "arg2").
		Target("builds/dependency").
		BuildTool("-copy", "arg1", "arg2").
		Build(makefile.New()).
		Build()

	fmt.Println(got)
}

func Test_builder_BuildTool(t *testing.T) {
	type fields struct {
		parent   *builder
		target   *Target
		children []*builder
	}
	type args struct {
		flag string
		args []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Builder
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &builder{
				parent:   tt.fields.parent,
				target:   tt.fields.target,
				children: tt.fields.children,
			}
			if got := b.BuildTool(tt.args.flag, tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildTool() = %v, want %v", got, tt.want)
			}
		})
	}
}
