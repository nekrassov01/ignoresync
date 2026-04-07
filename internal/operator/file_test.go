package operator

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/nekrassov01/ignoresync/internal/testutil"
)

func TestOperator_bundleFiles(t *testing.T) {
	type fields struct {
		repo   *RepoInfo
		dryrun bool
		w      io.Writer
	}
	type want struct {
		files   map[string]string
		output  []string
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	type store struct {
		gitDir      string
		maxFileNum  int
		maxFileSize int64
		maxRepoSize int64
	}
	tmp := new(store)
	tests := []struct {
		name   string
		fields fields
		hook   hook
		want   want
	}{
		{
			name: "success",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"success.txt"}),
					targetPatterns: getPatterns([]string{"success.txt"}),
				},
				dryrun: false,
				w:      &bytes.Buffer{},
			},
			want: want{
				files:   map[string]string{"success.txt": "success\n"},
				output:  nil,
				isError: false,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "success.txt"), []byte("success\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
					if err := os.Remove(filepath.Join(testutil.RepoPath, "success.txt")); err != nil {
						t.Fatal(err)
					}
				},
			},
		},
		{
			name: "success at .git skip",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"success.txt", "dot_git/ignored.txt"}),
					targetPatterns: getPatterns([]string{"success.txt", "dot_git/ignored.txt"}),
				},
				dryrun: false,
				w:      &bytes.Buffer{},
			},
			want: want{
				files:   map[string]string{"success.txt": "success\n"},
				output:  nil,
				isError: false,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "success.txt"), []byte("success\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "dot_git/ignored.txt"), []byte("ignored\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
					if err := os.Remove(filepath.Join(testutil.RepoPath, "success.txt")); err != nil {
						t.Fatal(err)
					}
					if err := os.Remove(filepath.Join(testutil.RepoPath, "dot_git/ignored.txt")); err != nil {
						t.Fatal(err)
					}
				},
			},
		},
		{
			name: "success at dryrun",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"dryrun.txt"}),
					targetPatterns: getPatterns([]string{"dryrun.txt"}),
				},
				dryrun: true,
				w:      &bytes.Buffer{},
			},
			want: want{
				files:   map[string]string{},
				output:  []string{"push (dryrun):", "uploaded", "dryrun.txt (7 B)"},
				isError: false,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "dryrun.txt"), []byte("dryrun\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
					if err := os.Remove(filepath.Join(testutil.RepoPath, "dryrun.txt")); err != nil {
						t.Fatal(err)
					}
				},
			},
		},
		{
			name: "error at too large file",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"too-large.txt"}),
					targetPatterns: getPatterns([]string{"too-large.txt"}),
				},
				dryrun: false,
				w:      &bytes.Buffer{},
			},
			want: want{
				files:   nil,
				output:  nil,
				isError: true,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					maxFileSize = 1
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "too-large.txt"), []byte("large"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
					maxFileSize = tmp.maxFileSize
					if err := os.Remove(filepath.Join(testutil.RepoPath, "too-large.txt")); err != nil {
						t.Fatal(err)
					}
				},
			},
		},
		{
			name: "error at too many files",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"too-many-1.txt", "too-many-2.txt"}),
					targetPatterns: getPatterns([]string{"too-many-1.txt", "too-many-2.txt"}),
				},
				dryrun: false,
				w:      &bytes.Buffer{},
			},
			want: want{
				files:   nil,
				output:  nil,
				isError: true,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					maxFileNum = 1
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "too-many-1.txt"), []byte("1"), 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "too-many-2.txt"), []byte("2"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
					maxFileNum = tmp.maxFileNum
					if err := os.Remove(filepath.Join(testutil.RepoPath, "too-many-1.txt")); err != nil {
						t.Fatal(err)
					}
					if err := os.Remove(filepath.Join(testutil.RepoPath, "too-many-2.txt")); err != nil {
						t.Fatal(err)
					}
				},
			},
		},
		{
			name: "error at too large repository",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"repo-size-1.txt", "repo-size-2.txt"}),
					targetPatterns: getPatterns([]string{"repo-size-1.txt", "repo-size-2.txt"}),
				},
				dryrun: false,
				w:      &bytes.Buffer{},
			},
			want: want{
				files:   nil,
				output:  nil,
				isError: true,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					maxRepoSize = 1
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "repo-size-1.txt"), []byte("1"), 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "repo-size-2.txt"), []byte("2"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
					maxRepoSize = tmp.maxRepoSize
					if err := os.Remove(filepath.Join(testutil.RepoPath, "repo-size-1.txt")); err != nil {
						t.Fatal(err)
					}
					if err := os.Remove(filepath.Join(testutil.RepoPath, "repo-size-2.txt")); err != nil {
						t.Fatal(err)
					}
				},
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
			o := &Operator{
				repo:   test.fields.repo,
				dryrun: test.fields.dryrun,
				w:      test.fields.w,
			}
			files, err := testutil.ReadTarGz(t, o.bundleFiles())
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, files, test.want.files)
			testutil.CheckContains(t, test.fields.w.(*bytes.Buffer).String(), test.want.output)
		})
	}
}

func TestOperator_bundlePatterns(t *testing.T) {
	type args struct {
		patterns []string
	}
	type want struct {
		patterns []string
		isError  bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				patterns: []string{".env", "config/*.yml"},
			},
			want: want{
				patterns: []string{".env", "config/*.yml"},
			},
		},
		{
			name: "success at empty",
			args: args{
				patterns: []string{},
			},
			want: want{
				patterns: []string{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{}
			var got []string
			err := testutil.ReadGzJSON(t, o.bundlePatterns(test.args.patterns), &got)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.patterns)
		})
	}
}

func TestOperator_restoreFiles(t *testing.T) {
	type fields struct {
		repo      *RepoInfo
		preview   bool
		overwrite bool
		w         io.Writer
	}
	type args struct {
		r io.Reader
	}
	type want struct {
		rel     string
		body    []byte
		mode    os.FileMode
		output  []string
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
				repo:      &RepoInfo{path: t.TempDir()},
				preview:   false,
				overwrite: true,
				w:         &bytes.Buffer{},
			},
			args: args{
				r: testutil.NewTarGzReader(t, "restored.txt", 0o640, []byte("restored\n")),
			},
			want: want{
				rel:     "./restored.txt",
				body:    []byte("restored\n"),
				mode:    0o640,
				output:  []string{"restored", "restored.txt (9 B)"},
				isError: false,
			},
		},
		{
			name: "success at dryrun",
			fields: fields{
				repo:      &RepoInfo{path: t.TempDir()},
				preview:   true,
				overwrite: false,
				w:         &bytes.Buffer{},
			},
			args: args{
				r: testutil.NewTarGzReader(t, "preview.txt", 0o600, []byte("preview\n")),
			},
			want: want{
				rel:     "./preview.txt",
				body:    nil,
				mode:    0,
				output:  []string{"preview", "preview.txt (8 B)"},
				isError: false,
			},
		},
		{
			name: "error at invalid archive path",
			fields: fields{
				repo:      &RepoInfo{path: t.TempDir()},
				preview:   false,
				overwrite: false,
				w:         &bytes.Buffer{},
			},
			args: args{
				r: testutil.NewTarGzReader(t, "../outside.txt", 0o600, []byte("outside\n")),
			},
			want: want{
				rel:     "../outside.txt",
				body:    nil,
				mode:    0,
				output:  nil,
				isError: true,
			},
		},
		{
			name: "error at invalid gzip",
			fields: fields{
				repo:      &RepoInfo{path: t.TempDir()},
				preview:   false,
				overwrite: false,
				w:         &bytes.Buffer{},
			},
			args: args{
				r: bytes.NewReader([]byte("invalid gzip")),
			},
			want: want{
				rel:     "dummy.txt",
				body:    nil,
				mode:    0,
				output:  nil,
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{
				repo:      test.fields.repo,
				preview:   test.fields.preview,
				overwrite: test.fields.overwrite,
				w:         test.fields.w,
			}
			err := o.restoreFiles(test.args.r)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckContains(t, test.fields.w.(*bytes.Buffer).String(), test.want.output)
			testutil.CheckFile(t, filepath.Join(test.fields.repo.path, test.want.rel), test.want.body, test.want.mode)
		})
	}
}

func TestOperator_restorePatterns(t *testing.T) {
	type fields struct {
		repo    *RepoInfo
		preview bool
		w       io.Writer
	}
	type args struct {
		r io.Reader
	}
	type want struct {
		patterns   []string
		output     []string
		matches    [][]string
		mismatches [][]string
		isError    bool
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
				repo:    &RepoInfo{targetPatterns: getPatterns([]string{"before.txt"})},
				preview: false,
				w:       &bytes.Buffer{},
			},
			args: args{
				r: (&Operator{}).bundlePatterns([]string{".env", "config/*.yml"}),
			},
			want: want{
				patterns:   []string{".env", "config/*.yml"},
				matches:    [][]string{{".env"}, {"config", "app.yml"}},
				mismatches: [][]string{{"before.txt"}, {"README.md"}},
				isError:    false,
			},
		},
		{
			name: "success at empty",
			fields: fields{
				repo:    &RepoInfo{targetPatterns: getPatterns([]string{"before.txt"})},
				preview: false,
				w:       &bytes.Buffer{},
			},
			args: args{
				r: (&Operator{}).bundlePatterns([]string{}),
			},
			want: want{
				patterns:   []string{},
				mismatches: [][]string{{"before.txt"}, {".env"}},
				isError:    false,
			},
		},
		{
			name: "success at dryrun",
			fields: fields{
				repo:    &RepoInfo{targetPatterns: getPatterns([]string{"before.txt"})},
				preview: true,
				w:       &bytes.Buffer{},
			},
			args: args{
				r: (&Operator{}).bundlePatterns([]string{".env", "config/*.yml"}),
			},
			want: want{
				patterns:   []string{".env", "config/*.yml"},
				output:     []string{"pattern", ".env", "config/*.yml"},
				matches:    [][]string{{"before.txt"}},
				mismatches: [][]string{{".env"}, {"config", "app.yml"}},
				isError:    false,
			},
		},
		{
			name: "error at invalid gzip",
			fields: fields{
				repo:    &RepoInfo{targetPatterns: getPatterns([]string{"before.txt"})},
				preview: false,
				w:       &bytes.Buffer{},
			},
			args: args{
				r: bytes.NewReader([]byte("invalid gzip")),
			},
			want: want{
				matches:    [][]string{{"before.txt"}},
				mismatches: [][]string{{".env"}},
				isError:    true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{
				repo:    test.fields.repo,
				preview: test.fields.preview,
				w:       test.fields.w,
			}
			got, err := o.restorePatterns(test.args.r)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.patterns)
			testutil.CheckContains(t, test.fields.w.(*bytes.Buffer).String(), test.want.output)
			matcher := gitignore.NewMatcher(o.repo.targetPatterns)
			for _, tokens := range test.want.matches {
				testutil.CheckValue(t, matcher.Match(tokens, false), true)
			}
			for _, tokens := range test.want.mismatches {
				testutil.CheckValue(t, matcher.Match(tokens, false), false)
			}
		})
	}
}

func TestOperator_cleanupFiles(t *testing.T) {
	type fields struct {
		repo *RepoInfo
		w    io.Writer
	}
	type want struct {
		removed   []string
		remaining []string
		output    []string
		isError   bool
	}
	type hook struct {
		before func()
		after  func()
	}
	type store struct {
		gitDir string
	}
	tmp := new(store)
	tests := []struct {
		name   string
		fields fields
		hook   hook
		want   want
	}{
		{
			name: "success",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"cleanup.txt"}),
					targetPatterns: getPatterns([]string{"cleanup.txt"}),
				},
				w: &bytes.Buffer{},
			},
			want: want{
				removed: []string{"cleanup.txt"},
				output:  []string{"clean:", "cleaned", "cleanup.txt"},
				isError: false,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "cleanup.txt"), []byte("cleanup\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
				},
			},
		},
		{
			name: "success at multiple files",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"multi-1.txt", "multi-2.txt"}),
					targetPatterns: getPatterns([]string{"multi-1.txt", "multi-2.txt"}),
				},
				w: &bytes.Buffer{},
			},
			want: want{
				removed: []string{"multi-1.txt", "multi-2.txt"},
				output:  []string{"clean:", "cleaned", "multi-1.txt", "multi-2.txt"},
				isError: false,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "multi-1.txt"), []byte("1\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "multi-2.txt"), []byte("2\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
				},
			},
		},
		{
			name: "success at subdirectory",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"sub_dir/cleanup.txt"}),
					targetPatterns: getPatterns([]string{"sub_dir/cleanup.txt"}),
				},
				w: &bytes.Buffer{},
			},
			want: want{
				removed: []string{"sub_dir/cleanup.txt"},
				output:  []string{"clean:", "cleaned", "sub_dir/cleanup.txt"},
				isError: false,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "sub_dir/cleanup.txt"), []byte("sub\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
				},
			},
		},
		{
			name: "success at .git skip",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"cleanup.txt", "dot_git/inside.txt"}),
					targetPatterns: getPatterns([]string{"cleanup.txt", "dot_git/inside.txt"}),
				},
				w: &bytes.Buffer{},
			},
			want: want{
				removed:   []string{"cleanup.txt"},
				remaining: []string{"dot_git/inside.txt"},
				output:    []string{"clean:", "cleaned", "cleanup.txt"},
				isError:   false,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "cleanup.txt"), []byte("cleanup\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "dot_git/inside.txt"), []byte("inside\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
					_ = os.Remove(filepath.Join(testutil.RepoPath, "dot_git/inside.txt"))
				},
			},
		},
		{
			name: "success at no match by ignore pattern",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"other.txt"}),
					targetPatterns: getPatterns([]string{"notignored.txt"}),
				},
				w: &bytes.Buffer{},
			},
			want: want{
				remaining: []string{"notignored.txt"},
				isError:   false,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "notignored.txt"), []byte("data\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
					_ = os.Remove(filepath.Join(testutil.RepoPath, "notignored.txt"))
				},
			},
		},
		{
			name: "success at no match by target pattern",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"nottarget.txt"}),
					targetPatterns: getPatterns([]string{"other.txt"}),
				},
				w: &bytes.Buffer{},
			},
			want: want{
				remaining: []string{"nottarget.txt"},
				isError:   false,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
					if err := os.WriteFile(filepath.Join(testutil.RepoPath, "nottarget.txt"), []byte("data\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				},
				after: func() {
					gitDir = tmp.gitDir
					_ = os.Remove(filepath.Join(testutil.RepoPath, "nottarget.txt"))
				},
			},
		},
		{
			name: "success at no matching files",
			fields: fields{
				repo: &RepoInfo{
					path:           testutil.RepoPath,
					ignorePatterns: getPatterns([]string{"nonexistent.txt"}),
					targetPatterns: getPatterns([]string{"nonexistent.txt"}),
				},
				w: &bytes.Buffer{},
			},
			want: want{
				isError: false,
			},
			hook: hook{
				before: func() {
					gitDir = "dot_git"
				},
				after: func() {
					gitDir = tmp.gitDir
				},
			},
		},
		{
			name: "error at invalid path",
			fields: fields{
				repo: &RepoInfo{
					path:           filepath.Join(testutil.RepoPath, "nonexistent"),
					ignorePatterns: getPatterns([]string{"*.txt"}),
					targetPatterns: getPatterns([]string{"*.txt"}),
				},
				w: &bytes.Buffer{},
			},
			want: want{
				isError: true,
			},
			hook: hook{},
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
			o := &Operator{
				repo: test.fields.repo,
				w:    test.fields.w,
			}
			err := o.cleanupFiles()
			testutil.CheckError(t, err != nil, test.want.isError)
			for _, f := range test.want.removed {
				testutil.CheckFileNotExists(t, filepath.Join(o.repo.path, f))
			}
			for _, f := range test.want.remaining {
				testutil.CheckFileExists(t, filepath.Join(o.repo.path, f))
			}
			testutil.CheckContains(t, test.fields.w.(*bytes.Buffer).String(), test.want.output)
		})
	}
}

func TestOperator_compareHash(t *testing.T) {
	type args struct {
		path   string
		remote *diffInfo
	}
	type want struct {
		value   bool
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success equal",
			args: args{
				path: func() string {
					path := filepath.Join(t.TempDir(), "equal.txt")
					data := []byte("same")
					if err := os.WriteFile(path, data, 0o600); err != nil {
						t.Fatal(err)
					}
					return path
				}(),
				remote: &diffInfo{hash: sha256.Sum256([]byte("same"))},
			},
			want: want{
				value: true,
			},
		},
		{
			name: "success different",
			args: args{
				path: func() string {
					path := filepath.Join(t.TempDir(), "different.txt")
					data := []byte("left")
					if err := os.WriteFile(path, data, 0o600); err != nil {
						t.Fatal(err)
					}
					return path
				}(),
				remote: &diffInfo{hash: sha256.Sum256([]byte("right"))},
			},
			want: want{
				value: false,
			},
		},
		{
			name: "error at open file",
			args: args{
				path:   filepath.Join(t.TempDir(), "missing.txt"),
				remote: &diffInfo{},
			},
			want: want{
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{}
			got, err := o.compareHash(test.args.path, test.args.remote)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func TestOperator_reportDiff(t *testing.T) {
	type fields struct {
		w io.Writer
	}
	type args struct {
		rel    string
		abs    string
		tmp    *os.File
		local  diffInfo
		remote diffInfo
	}
	type want struct {
		values  []string
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
				w: &bytes.Buffer{},
			},
			args: args{
				rel: "file.txt",
				abs: func() string {
					abs := filepath.Join(t.TempDir(), "file.txt")
					if err := os.WriteFile(abs, []byte("left\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					return abs
				}(),
				tmp:    testutil.CreateTemp(t, "right\n"),
				local:  diffInfo{size: 5},
				remote: diffInfo{size: 6},
			},
			want: want{
				values: []string{
					"left",
					"right",
				},
				isError: false,
			},
		},
		{
			name: "diff too large",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				rel: "file.txt",
				abs: func() string {
					abs := filepath.Join(t.TempDir(), "file.txt")
					if err := os.WriteFile(abs, []byte("left\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					return abs
				}(),
				local:  diffInfo{size: maxDiffSize + 1},
				remote: diffInfo{size: 4},
			},
			want: want{
				values: []string{
					"pull:",
					"diff too large, skipping diff",
					"file.txt (4 B)",
				},
				isError: false,
			},
		},
		{
			name: "binary detected",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				rel: "file.bin",
				abs: func() string {
					abs := filepath.Join(t.TempDir(), "file.bin")
					if err := os.WriteFile(abs, []byte("\x00\x01\x02"), 0o600); err != nil {
						t.Fatal(err)
					}
					return abs
				}(),
				tmp:    testutil.CreateTemp(t, "\x00\x02\x03"),
				local:  diffInfo{size: 3},
				remote: diffInfo{size: 3},
			},
			want: want{
				values: []string{
					"pull:",
					"binary detected, skipping diff",
					"file.bin (3 B)",
				},
				isError: false,
			},
		},
		{
			name: "error at read local file",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				rel:    "missing.txt",
				abs:    filepath.Join(t.TempDir(), "missing.txt"),
				tmp:    testutil.CreateTemp(t, "left\n"),
				local:  diffInfo{size: 5},
				remote: diffInfo{size: 6},
			},
			want: want{
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{
				w: test.fields.w,
			}
			err := o.reportDiff(test.args.rel, test.args.abs, test.args.tmp, test.args.local, test.args.remote)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckContains(t, o.w.(*bytes.Buffer).String(), test.want.values)
		})
	}
}

func Test_isOutside(t *testing.T) {
	type args struct {
		path string
	}
	type want struct {
		value bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "relative path",
			args: args{
				path: "dir/file.txt",
			},
			want: want{
				value: false,
			},
		},
		{
			name: "parent path",
			args: args{
				path: "..",
			},
			want: want{
				value: true,
			},
		},
		{
			name: "traversal path",
			args: args{
				path: filepath.Join("..", "file.txt"),
			},
			want: want{
				value: true,
			},
		},
		{
			name: "absolute path",
			args: args{
				path: filepath.Join(string(filepath.Separator), "tmp", "file.txt"),
			},
			want: want{
				value: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isOutside(test.args.path)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func Test_isBinary(t *testing.T) {
	type args struct {
		data []byte
	}
	type want struct {
		value bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "text",
			args: args{
				data: []byte("hello"),
			},
			want: want{
				value: false,
			},
		},
		{
			name: "binary",
			args: args{
				data: []byte{0x00, 0x01, 0x02},
			},
			want: want{
				value: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isBinary(test.args.data)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}
