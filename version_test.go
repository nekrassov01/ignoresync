package ignoresync

import (
	"fmt"
	"testing"

	"github.com/nekrassov01/ignoresync/testutil"
)

func Test_getVersion(t *testing.T) {
	type hook struct {
		before func()
		after  func()
	}
	type store struct {
		revision string
	}
	tmp := new(store)
	tests := []struct {
		name string
		want string
		hook hook
	}{
		{
			name: "with revision",
			want: fmt.Sprintf("%s (revision: 1234567)", version),
			hook: hook{
				before: func() {
					tmp.revision = revision
					revision = "1234567"
				},
				after: func() {
					revision = tmp.revision
				},
			},
		},
		{
			name: "without revision",
			want: version,
			hook: hook{
				before: func() {
					tmp.revision = revision
					revision = ""
				},
				after: func() {
					revision = tmp.revision
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			defer func() {
				if test.hook.after != nil {
					test.hook.after()
				}
			}()
			got := Version()
			testutil.CheckValue(t, got, test.want)
		})
	}
}
