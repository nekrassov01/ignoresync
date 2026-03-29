package main

import (
	"testing"
)

// TestMain sets up the testing environment for command tests.
func TestMain(m *testing.M) {
	original := revision
	defer func() {
		revision = original
	}()
	m.Run()
}
