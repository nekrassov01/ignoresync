package operator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/nekrassov01/ignoresync/params"
	"github.com/nekrassov01/ignoresync/testutil"
)

func Test_newRepoInfo(t *testing.T) {
	type args struct {
		path   string
		remote string
	}
	type want struct {
		repo    *RepoInfo
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name string
		args args
		want want
		hook hook
	}{
		{
			name: "success",
			args: args{
				path:   testutil.RepoPath,
				remote: params.DefaultRemoteName,
			},
			want: want{
				repo: &RepoInfo{
					Name:   "example.com/localuser/repo",
					Hash:   "de8635e39fd183ba59e5d57ad1ce93c6e96ccd5cf7e80d67801e5ea730487c93",
					path:   testutil.RepoPath,
					remote: params.DefaultRemoteName,
					user:   "localuser",
				},
				isError: false,
			},
			hook: hook{
				before: func() {
					err := os.Rename(testutil.GitDir["before"], testutil.GitDir["after"])
					if err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					err := os.Rename(testutil.GitDir["after"], testutil.GitDir["before"])
					if err != nil {
						t.Fatal(err)
					}
				},
			},
		},
		{
			name: "error at opening repo",
			args: args{
				path:   t.TempDir(),
				remote: params.DefaultRemoteName,
			},
			want: want{
				repo:    nil,
				isError: true,
			},
		},
		{
			name: "error at getting remote URL",
			args: args{
				path: func() string {
					dir := t.TempDir()
					if _, err := git.PlainInit(dir, false); err != nil {
						t.Fatal(err)
					}
					return dir
				}(),
				remote: params.DefaultRemoteName,
			},
			want: want{
				repo:    nil,
				isError: true,
			},
		},
		{
			name: "error at getting repo name",
			args: args{
				path: func() string {
					dir := t.TempDir()
					repo, err := git.PlainInit(dir, false)
					if err != nil {
						t.Fatal(err)
					}
					cfg := &config.RemoteConfig{
						Name: params.DefaultRemoteName,
						URLs: []string{""},
					}
					if _, err := repo.CreateRemote(cfg); err != nil {
						t.Fatal(err)
					}
					return dir
				}(),
				remote: params.DefaultRemoteName,
			},
			want: want{
				repo:    nil,
				isError: true,
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
			got, err := newRepoInfo(test.args.path, test.args.remote)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.repo)
		})
	}
}

func Test_findRepoRoot(t *testing.T) {
	type args struct {
		path string
	}
	type want struct {
		dir     string
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name string
		args args
		want want
		hook hook
	}{
		{
			name: "repo root",
			args: args{
				path: testutil.RepoPath,
			},
			want: want{
				dir:     testutil.RepoPath,
				isError: false,
			},
			hook: hook{
				before: func() {
					err := os.Rename(testutil.GitDir["before"], testutil.GitDir["after"])
					if err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					err := os.Rename(testutil.GitDir["after"], testutil.GitDir["before"])
					if err != nil {
						t.Fatal(err)
					}
				},
			},
		},
		{
			name: "sub directory",
			args: args{
				path: filepath.Join(testutil.RepoPath, "sub_dir"),
			},
			want: want{
				dir:     testutil.RepoPath,
				isError: false,
			},
			hook: hook{
				before: func() {
					err := os.Rename(testutil.GitDir["before"], testutil.GitDir["after"])
					if err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					err := os.Rename(testutil.GitDir["after"], testutil.GitDir["before"])
					if err != nil {
						t.Fatal(err)
					}
				},
			},
		},
		{
			name: "no repo",
			args: args{
				path: t.TempDir(),
			},
			want: want{
				dir:     "",
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			if test.hook.after != nil {
				defer test.hook.after()
			}
			dir, _, err := findRepoRoot(test.args.path)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, dir, test.want.dir)
		})
	}
}
func Test_getRemoteURL(t *testing.T) {
	type args struct {
		repo *git.Repository
		name string
	}
	type want struct {
		value   string
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "origin",
			args: args{
				repo: func() *git.Repository {
					repo, err := git.PlainInit(t.TempDir(), false)
					if err != nil {
						t.Fatal(err)
					}
					remote := &config.RemoteConfig{
						Name: params.DefaultRemoteName,
						URLs: []string{
							"https://example.com/user/repo.git",
							"git@example.com:user/repo.git",
						},
					}
					if _, err := repo.CreateRemote(remote); err != nil {
						t.Fatal(err)
					}
					return repo
				}(),
				name: params.DefaultRemoteName,
			},
			want: want{
				value:   "https://example.com/user/repo.git",
				isError: false,
			},
		},
		{
			name: "not origin",
			args: args{
				repo: func() *git.Repository {
					repo, err := git.PlainInit(t.TempDir(), false)
					if err != nil {
						t.Fatal(err)
					}
					remote := &config.RemoteConfig{
						Name: "not-origin",
						URLs: []string{
							"https://example.com/user/repo.git",
							"git@example.com:user/repo.git",
						},
					}
					if _, err := repo.CreateRemote(remote); err != nil {
						t.Fatal(err)
					}
					return repo
				}(),
				name: "not-origin",
			},
			want: want{
				value:   "https://example.com/user/repo.git",
				isError: false,
			},
		},
		{
			name: "empty",
			args: args{
				repo: func() *git.Repository {
					repo, err := git.PlainInit(t.TempDir(), false)
					if err != nil {
						t.Fatal(err)
					}
					remote := &config.RemoteConfig{
						Name: params.DefaultRemoteName,
						URLs: []string{
							"https://example.com/user/repo.git",
							"git@example.com:user/repo.git",
						},
					}
					if _, err := repo.CreateRemote(remote); err != nil {
						t.Fatal(err)
					}
					return repo
				}(),
				name: "",
			},
			want: want{
				value:   "https://example.com/user/repo.git",
				isError: false,
			},
		},
		{
			name: "not found",
			args: args{
				repo: func() *git.Repository {
					repo, err := git.PlainInit(t.TempDir(), false)
					if err != nil {
						t.Fatal(err)
					}
					remote := &config.RemoteConfig{
						Name: params.DefaultRemoteName,
						URLs: []string{
							"https://example.com/user/repo.git",
							"git@example.com:user/repo.git",
						},
					}
					if _, err := repo.CreateRemote(remote); err != nil {
						t.Fatal(err)
					}
					return repo
				}(),
				name: "not-found",
			},
			want: want{
				value:   "https://example.com/user/repo.git",
				isError: false,
			},
		},
		{
			name: "no valid URL",
			args: args{
				repo: func() *git.Repository {
					repo, err := git.PlainInit(t.TempDir(), false)
					if err != nil {
						t.Fatal(err)
					}
					remote := &config.RemoteConfig{
						Name: params.DefaultRemoteName,
						URLs: []string{
							"",
							"",
						},
					}
					if _, err := repo.CreateRemote(remote); err != nil {
						t.Fatal(err)
					}
					return repo
				}(),
				name: params.DefaultRemoteName,
			},
			want: want{
				value:   "",
				isError: true,
			},
		},
		{
			name: "no repo",
			args: args{
				repo: nil,
				name: params.DefaultRemoteName,
			},
			want: want{
				value:   "",
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := getRemoteURL(test.args.repo, test.args.name)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func Test_getRepoName(t *testing.T) {
	type args struct {
		url string
	}
	type want struct {
		name    string
		hash    string
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "https",
			args: args{
				url: "https://example.com/user/repo.git",
			},
			want: want{
				name:    "example.com/user/repo",
				hash:    "5183e9f557bef70c519171a4693b10f979fe3eb562eaa83e4090419ca82bbef5",
				isError: false,
			},
		},
		{
			name: "ssh",
			args: args{
				url: "git@example.com:user/repo.git",
			},
			want: want{
				name:    "example.com/user/repo",
				hash:    "5183e9f557bef70c519171a4693b10f979fe3eb562eaa83e4090419ca82bbef5",
				isError: false,
			},
		},
		{
			name: "invalid",
			args: args{
				url: "example.com",
			},
			want: want{
				name:    "",
				hash:    "",
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			name, hash, err := getRepoName(test.args.url)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, name, test.want.name)
			testutil.CheckValue(t, hash, test.want.hash)
		})
	}
}

func Test_getPatterns(t *testing.T) {
	type args struct {
		patterns []string
	}
	type want struct {
		patterns []gitignore.Pattern
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "basic",
			args: args{
				patterns: []string{"**/.env*"},
			},
			want: want{
				patterns: []gitignore.Pattern{gitignore.ParsePattern("**/.env*", nil)},
			},
		},
		{
			name: "empty",
			args: args{
				patterns: []string{},
			},
			want: want{
				patterns: nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := getPatterns(test.args.patterns)
			testutil.CheckValue(t, got, test.want.patterns)
		})
	}
}

func Test_getUser(t *testing.T) {
	type args struct {
		repo *git.Repository
	}
	type want struct {
		value string
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name string
		args args
		want want
		hook hook
	}{
		{
			name: "from repo config",
			args: args{
				repo: func() *git.Repository {
					repo, err := git.PlainInit(t.TempDir(), false)
					if err != nil {
						t.Fatal(err)
					}
					cfg, err := repo.Config()
					if err != nil {
						t.Fatal(err)
					}
					cfg.User.Name = "repouser"
					if err := repo.SetConfig(cfg); err != nil {
						t.Fatal(err)
					}
					return repo
				}(),
			},
			want: want{
				value: "repouser",
			},
		},
		{
			name: "from environment",
			args: args{
				repo: nil,
			},
			want: want{
				value: "envuser",
			},
			hook: hook{
				before: func() {
					t.Setenv("GIT_AUTHOR_NAME", "envuser")
				},
				after: func() {
					t.Setenv("GIT_AUTHOR_NAME", "")
				},
			},
		},
		{
			name: "unknown",
			args: args{
				repo: nil,
			},
			want: want{
				value: params.DefaultUserName,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			if test.hook.after != nil {
				defer test.hook.after()
			}
			got := getUser(test.args.repo)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}
