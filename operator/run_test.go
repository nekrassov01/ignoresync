//go:build darwin || linux

package operator

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/nekrassov01/ignoresync/testutil"
)

func TestOperator_Run(t *testing.T) {
	type fields struct {
		repo    *RepoInfo
		workDir string
		w       io.Writer
	}
	type args struct {
		ctx     context.Context
		command string
		environ []string
	}
	type want struct {
		isError bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "success",
			fields: fields{
				repo: &RepoInfo{
					path: testutil.RepoPath,
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:     context.Background(),
				command: "echo hello",
				environ: nil,
			},
			want: want{
				isError: false,
			},
		},
		{
			name: "success at work dir",
			fields: fields{
				repo: &RepoInfo{
					path: testutil.RepoPath,
				},
				workDir: "sub_dir",
				w:       &bytes.Buffer{},
			},
			args: args{
				ctx:     context.Background(),
				command: "echo hello",
				environ: nil,
			},
			want: want{
				isError: false,
			},
		},
		{
			name: "success at environ",
			fields: fields{
				repo: &RepoInfo{
					path: testutil.RepoPath,
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:     context.Background(),
				command: "echo $TEST_VAR",
				environ: []string{"TEST_VAR=hello"},
			},
			want: want{
				isError: false,
			},
		},
		{
			name: "error at command failure",
			fields: fields{
				repo: &RepoInfo{
					path: testutil.RepoPath,
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:     context.Background(),
				command: "exit 1",
				environ: nil,
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at command not found",
			fields: fields{
				repo: &RepoInfo{
					path: testutil.RepoPath,
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:     context.Background(),
				command: "nonexistent_command_xyz_12345",
				environ: nil,
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at invalid workDir",
			fields: fields{
				repo: &RepoInfo{
					path: testutil.RepoPath,
				},
				workDir: "nonexistent",
				w:       &bytes.Buffer{},
			},
			args: args{
				ctx:     context.Background(),
				command: "echo hello",
				environ: nil,
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at cancelled context",
			fields: fields{
				repo: &RepoInfo{
					path: testutil.RepoPath,
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				command: "sleep 10",
				environ: nil,
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at context timeout with sigkill",
			fields: fields{
				repo: &RepoInfo{
					path: testutil.RepoPath,
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
					t.Cleanup(cancel)
					return ctx
				}(),
				command: "sleep 10",
				environ: nil,
			},
			want: want{
				isError: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Operator{
				repo:    tt.fields.repo,
				workDir: tt.fields.workDir,
				w:       tt.fields.w,
			}
			err := o.Run(tt.args.ctx, tt.args.command, tt.args.environ)
			testutil.CheckError(t, err != nil, tt.want.isError)
		})
	}
}
