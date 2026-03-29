package operator

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/nekrassov01/ignoresync/color"
	"github.com/nekrassov01/ignoresync/prompt"
	"znkr.io/diff/textdiff"
	diffcolor "znkr.io/diff/textdiff/color"
)

// diffInfo represents the file information used for diffing local and remote files.
type diffInfo struct {
	size int64
	mode os.FileMode
	hash [32]byte
}

// bundleFiles creates a tar.gz archive of the repository files and returns it as an io.ReadCloser.
func (o *Operator) bundleFiles() io.ReadCloser {
	pr, pw := io.Pipe()

	go func() {
		var err error
		gw := gzip.NewWriter(pw)
		tw := tar.NewWriter(gw)

		defer func() {
			if closeErr := tw.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if closeErr := gw.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("failed to bundle files: %w", err))
			} else {
				_ = pw.Close()
			}
		}()

		var fileNum int
		var repoSize int64

		ignoreMatcher := gitignore.NewMatcher(o.repo.ignorePatterns)
		targetMatcher := gitignore.NewMatcher(o.repo.targetPatterns)

		err = filepath.WalkDir(o.repo.path, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return fmt.Errorf("failed to access local file: %w", walkErr)
			}
			if d.IsDir() && d.Name() == gitDir {
				return filepath.SkipDir
			}
			if d.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(o.repo.path, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			tokens := strings.Split(filepath.ToSlash(rel), "/")
			if !ignoreMatcher.Match(tokens, false) || !targetMatcher.Match(tokens, false) {
				return nil
			}

			fi, err := d.Info()
			if err != nil {
				return fmt.Errorf("failed to get file info: %w", err)
			}
			if fi.Mode()&os.ModeSymlink != 0 {
				return nil
			}

			fileSize := fi.Size()
			if fileSize > maxFileSize {
				return fmt.Errorf("file too large: %s, bytes=%d, max=%d", rel, fileSize, maxFileSize)
			}

			fileNum++
			if fileNum > maxFileNum {
				return fmt.Errorf("too many files: %s, max=%d", rel, maxFileNum)
			}

			repoSize += fileSize
			if repoSize > maxRepoSize {
				return fmt.Errorf("repository size exceeds limit: %s, max=%d", rel, maxRepoSize)
			}

			if o.dryrun {
				o.printStatus("pushed", rel, fi.Size(), "dryrun:")
				return nil
			}

			header, err := tar.FileInfoHeader(fi, "")
			if err != nil {
				return fmt.Errorf("failed to create tar header: %w", err)
			}
			header.Name = filepath.ToSlash(rel)

			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write tar header: %w", err)
			}

			if err := o.bundleFile(tw, rel, path); err != nil {
				return fmt.Errorf("failed to bundle file: %w", err)
			}

			return nil
		})
	}()

	return pr
}

// bundleFile writes the file content to the tar writer.
func (o *Operator) bundleFile(tw *tar.Writer, rel, abs string) error {
	f, err := os.Open(filepath.Clean(abs))
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	size, err := io.Copy(tw, f)
	if err != nil {
		return fmt.Errorf("failed to write file to tar: %w", err)
	}

	o.printStatus("pushed", rel, size, "state:")
	return nil
}

// bundlePatterns creates a gzip-compressed JSON array of target patterns and returns it as an io.ReadCloser.
func (o *Operator) bundlePatterns(patterns []string) io.ReadCloser {
	pr, pw := io.Pipe()

	go func() {
		var err error
		gw := gzip.NewWriter(pw)

		defer func() {
			if closeErr := gw.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("failed to bundle patterns: %w", err))
			} else {
				_ = pw.Close()
			}
		}()

		err = json.NewEncoder(gw).Encode(patterns)
	}()

	return pr
}

// restoreFiles extracts files from the given tar.gz reader into the repository path.
func (o *Operator) restoreFiles(r io.Reader) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if closeErr := gr.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar archive: %w", err)
		}

		err = func() error {
			// Even if a gosec warning occurs in a child function, it is sanitized here.
			rel := filepath.Clean(header.Name)
			abs := filepath.Join(o.repo.path, rel)

			if isOutside(rel) {
				return fmt.Errorf("invalid file path in archive: %s", header.Name)
			}

			if o.dryrun {
				o.printStatus("file", filepath.ToSlash(header.Name), header.Size, "preview:")
				return nil
			}

			var l, r diffInfo
			r.size = header.FileInfo().Size()
			r.mode = header.FileInfo().Mode().Perm()

			if header.Typeflag == tar.TypeDir {
				if err := os.MkdirAll(abs, r.mode); err != nil {
					return fmt.Errorf("failed to create directory: %w", err)
				}
			}

			tmp, err := os.CreateTemp("", "ignoresync-*")
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}
			defer func() {
				_ = tmp.Close()
				_ = os.Remove(tmp.Name())
			}()

			h := sha256.New()
			if _, err := io.CopyN(io.MultiWriter(tmp, h), tr, r.size); err != nil {
				return fmt.Errorf("failed to read remote file content: %w", err)
			}
			copy(r.hash[:], h.Sum(nil))

			if o.overwrite {
				return o.restoreFile(rel, abs, tmp, r.mode)
			}

			fi, err := os.Stat(abs)
			if os.IsNotExist(err) {
				return o.restoreFile(rel, abs, tmp, r.mode)
			}
			if err != nil || fi == nil {
				return fmt.Errorf("failed to stat local file: %w", err)
			}

			l.size = fi.Size()
			l.mode = fi.Mode().Perm()

			sizeEqual := l.size == r.size
			modeEqual := l.mode == r.mode
			bodyEqual := sizeEqual

			if sizeEqual {
				eq, err := o.compareHash(abs, &r)
				if err != nil {
					return fmt.Errorf("failed to compare files: %w", err)
				}
				bodyEqual = eq
			}

			if sizeEqual && modeEqual && bodyEqual {
				o.printStatus("no changes", rel, r.size, "state:")
				return nil
			}

			if !bodyEqual {
				o.printStatus("content changed", rel, r.size, "state:")
				if err := o.reportDiff(rel, abs, tmp, l, r); err != nil {
					return fmt.Errorf("failed to report diff: %w", err)
				}
			}

			if !modeEqual {
				o.printStatus(fmt.Sprintf("mode changed, local=%o, remote=%o", l.mode, r.mode), rel, r.size, "state:")
			}

			if _, err := prompt.Confirm(o.w, fmt.Sprintf("%s overwrite? %s %s (%s)", color.Mute("state:"), color.Mute("->"), rel, sizeString(r.size)), "skipped"); err != nil {
				o.printStatus(err.Error(), rel, r.size, "state:")
				return nil
			}

			return o.restoreFile(rel, abs, tmp, r.mode)
		}()

		if err != nil {
			return err
		}
	}
	return nil
}

// restoreFile writes the content from the temp file to the target path with the given mode.
func (o *Operator) restoreFile(rel, abs string, tmp *os.File, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(abs), 0o0750); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek remote file: %w", err)
	}

	f, err := os.OpenFile(abs, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	size, err := io.Copy(f, tmp)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := os.Chmod(abs, mode); err != nil {
		return fmt.Errorf("failed to chmod local file: %w", err)
	}

	o.printStatus("restored", rel, size, "state:")
	return nil
}

// restorePatterns decompresses and deserializes patterns from the given reader.
func (o *Operator) restorePatterns(r io.Reader) ([]string, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if closeErr := gr.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var patterns []string
	if err := json.NewDecoder(gr).Decode(&patterns); err != nil {
		return nil, fmt.Errorf("failed to decode patterns: %w", err)
	}

	if o.dryrun {
		for _, pattern := range patterns {
			o.printStatus("pattern", pattern, -1, "preview:")
		}
		return patterns, nil
	}

	if _, err = io.CopyN(io.Discard, gr, maxRepoSize); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to validate gzip footer checksum: %w", err)
	}

	o.repo.targetPatterns = getPatterns(patterns)
	return patterns, nil
}

// compareHash compares the SHA-256 hash of the local file with the remote file.
func (o *Operator) compareHash(path string, remote *diffInfo) (bool, error) {
	lf, err := os.Open(path) // #nosec G304
	if err != nil {
		return false, fmt.Errorf("failed to open local file: %w", err)
	}
	defer func() {
		_ = lf.Close()
	}()

	lh := sha256.New()
	if _, err := io.Copy(lh, lf); err != nil {
		return false, fmt.Errorf("failed to hash local file: %w", err)
	}

	return bytes.Equal(lh.Sum(nil), remote.hash[:]), nil
}

// reportDiff generates and displays the diff between the local and remote files.
func (o *Operator) reportDiff(rel, abs string, tmp *os.File, local, remote diffInfo) error {
	if local.size > maxDiffSize || remote.size > maxDiffSize {
		o.printStatus("diff too large, skipping diff", rel, remote.size, "state:")
		return nil
	}

	lf, err := os.ReadFile(abs) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek remote file: %w", err)
	}

	rf, err := io.ReadAll(tmp)
	if err != nil {
		return fmt.Errorf("failed to read remote file: %w", err)
	}

	if isBinary(lf) || isBinary(rf) {
		o.printStatus("binary detected, skipping diff", rel, remote.size, "state:")
		return nil
	}

	content := textdiff.Unified(
		string(lf),
		string(rf),
		textdiff.IndentHeuristic(),
		textdiff.TerminalColors(
			diffcolor.HunkHeaders(90),
			diffcolor.Matches(90),
			diffcolor.Inserts(1, 92),
			diffcolor.Deletes(1, 91),
		))
	_, _ = fmt.Fprintln(o.w, content)

	return nil
}

// isOutside checks if the given path is absolute or traverses outside.
func isOutside(path string) bool {
	prefix := ".." + string(filepath.Separator)
	return filepath.IsAbs(path) || path == ".." || strings.HasPrefix(path, prefix)
}

// isBinary checks if the given data is binary.
func isBinary(data []byte) bool {
	return http.DetectContentType(data) == "application/octet-stream"
}

// printStatus prints the status of the file operation in a formatted manner.
func (o *Operator) printStatus(msg, path string, size int64, prefix string) {
	if prefix != "" {
		_, _ = io.WriteString(o.w, color.Mute(prefix))
		_, _ = io.WriteString(o.w, " ")
	}
	_, _ = io.WriteString(o.w, msg)
	_, _ = io.WriteString(o.w, color.Mute(" -> "))
	_, _ = io.WriteString(o.w, path)
	if size >= 0 {
		_, _ = io.WriteString(o.w, " (")
		_, _ = io.WriteString(o.w, sizeString(size))
		_, _ = io.WriteString(o.w, ")")
	}
	_, _ = io.WriteString(o.w, "\n")
}

// sizeString formats the file size in a human-readable format, handling negative sizes as zero.
func sizeString(size int64) string {
	if size < 0 {
		return humanize.Bytes(0)
	}
	return humanize.Bytes(uint64(size))
}
