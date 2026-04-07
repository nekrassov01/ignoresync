package testutil

import (
	"errors"
	"io"
)

// ReadErrorReadCloser is a mock implementation of io.ReadCloser.
// This causes an error while reading the response body use to test
// the scenario.
type ReadErrorReadCloser struct {
	io.Reader
}

// Read is the mock read method that always returns an error.
func (ReadErrorReadCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

// Close is the mock close method that does nothing.
func (ReadErrorReadCloser) Close() error {
	return nil
}

// CloseErrorReadCloser is a mock implementation of io.ReadCloser.
// This will cause an error while closing the response body use
// to test the scenario.
type CloseErrorReadCloser struct {
	io.Reader
}

// Read is the mock read method that reads from the embedded reader.
func (CloseErrorReadCloser) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

// Close is the mock close method that always returns an error.
func (CloseErrorReadCloser) Close() error {
	return errors.New("close error")
}

// IsClosedReadCloser is a mock implementation of io.ReadCloser.
// It records whether Close has been called.
type IsClosedReadCloser struct {
	io.Reader

	Closed *bool
}

// Read is the mock read method that reads from the embedded reader.
func (o IsClosedReadCloser) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

// Close records that Close has been called.
func (o IsClosedReadCloser) Close() error {
	if o.Closed != nil {
		*o.Closed = true
	}
	return nil
}
