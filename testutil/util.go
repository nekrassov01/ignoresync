package testutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/nekrassov01/ignoresync/params"
)

// RepoPath is the path to the test repository used in tests.
var RepoPath = filepath.Join("..", "testdata", "sandbox")

// GitDir is the path to the .git directory of the test repository.
var GitDir = map[string]string{
	"before": filepath.Join(RepoPath, "dot_git"),
	"after":  filepath.Join(RepoPath, ".git"),
}

// NewError creates a new generic error for testing.
func NewError() error {
	return errors.New("error")
}

// CheckError checks error mismatch in tests.
func CheckError(t *testing.T, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("\nerror mismatch\ngot:\n%v\nwant:\n%v\n", got, want)
	}
}

// CheckValue checks value mismatch in tests.
func CheckValue(t *testing.T, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\nvalue mismatch\ngot:\n%v\nwant:\n%v\n", got, want)
	}
}

// CheckHex checks hexadecimal value mismatch in tests.
func CheckHex(t *testing.T, got any, length int) {
	t.Helper()
	r := regexp.MustCompile(fmt.Sprintf("^[0-9a-fA-F]{%d}$", length))
	switch v := got.(type) {
	case string:
		if !r.MatchString(v) {
			t.Errorf("\nhex mismatch\ngot:\n%v\n", got)
		}
	case []byte:
		if !r.Match(v) {
			t.Errorf("\nhex mismatch\ngot:\n%v\n", got)
		}
	default:
		t.Fatalf("unsupported type: %T", got)
	}
}

// CheckContains checks if all expected substrings are present in the given string.
func CheckContains(t *testing.T, got string, wants []string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("\nvalue not contains\ngot:\n%v\nwants:\n%v\n", got, want)
		}
	}
}

// CheckFile checks if a file exists with the expected content and permissions.
func CheckFile(t *testing.T, path string, body []byte, mode os.FileMode) {
	t.Helper()
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, body) {
		t.Errorf("\nbody mismatch\ngot:\n%v\nwant:\n%v\n", got, body)
	}
	if fi.Mode().Perm() != mode {
		t.Errorf("\nmode mismatch\ngot:\n%v\nwant:\n%v\n", fi.Mode().Perm(), mode)
	}
}

// CheckFileExists checks that a file exists at the given path.
func CheckFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("\nfile does not exist:\n%s\n", path)
	}
}

// CheckFileNotExists checks that a file does not exist at the given path.
func CheckFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("\nfile still exists:\n%s\n", path)
	}
}

// ReadBody reads and closes a ReadCloser for testing.
func ReadBody(t *testing.T, body io.ReadCloser) []byte {
	t.Helper()
	if body == nil {
		return nil
	}
	b, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	_ = body.Close()
	return b
}

// ReadTarGz reads a tar.gz stream and returns file contents keyed by header name.
func ReadTarGz(t *testing.T, body io.ReadCloser) (map[string]string, error) {
	t.Helper()
	if body == nil {
		return map[string]string{}, nil
	}
	payload, err := io.ReadAll(body)
	defer func() {
		_ = body.Close()
	}()
	if err != nil {
		return nil, err
	}
	gr, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = gr.Close()
	}()
	tr := tar.NewReader(gr)
	files := make(map[string]string, 4)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		files[header.Name] = string(data)
	}
	return files, nil
}

// ReadGzJSON reads a gzip-compressed JSON stream into the provided value.
func ReadGzJSON(t *testing.T, body io.ReadCloser, v any) error {
	t.Helper()
	if body == nil {
		return nil
	}
	payload, err := io.ReadAll(body)
	defer func() {
		_ = body.Close()
	}()
	if err != nil {
		return err
	}
	gr, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = gr.Close()
	}()
	if err := json.NewDecoder(gr).Decode(v); err != nil {
		return err
	}
	return nil
}

// NewTarGzReader creates an in-memory tar.gz reader for testing.
func NewTarGzReader(t *testing.T, name string, mode int64, body []byte) io.Reader {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	h := &tar.Header{
		Name: name,
		Mode: mode,
		Size: int64(len(body)),
	}
	if err := tw.WriteHeader(h); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(buf.Bytes())
}

// CreateTemp creates a temporary file with the given input string for testing.
func CreateTemp(t *testing.T, input string) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), params.DefaultTempPattern)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(input); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = f.Close()
	})
	return f
}

// SetStdin sets the given file as os.Stdin for testing and restores it after the test.
func SetStdin(t *testing.T, f *os.File) {
	t.Helper()
	old := os.Stdin
	os.Stdin = f
	t.Cleanup(func() {
		os.Stdin = old
	})
}
