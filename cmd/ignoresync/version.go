package main

import "fmt"

// version is the current version.
const version = "0.0.1"

// revision is the git revision.
var revision = ""

// getVersion returns the version and revision.
func getVersion() string {
	if revision == "" {
		return version
	}
	return fmt.Sprintf("%s (revision: %s)", version, revision)
}
