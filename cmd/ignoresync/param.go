package main

const (
	// categoryGlobal is the category for global commands that affect the entire environment.
	categoryGlobal = "GLOBAL COMMANDS"

	// categoryLocal is the category for local commands that affect the local machine.
	categoryLocal = "LOCAL COMMANDS"

	// categoryRepository is the category for repository commands that affect the repository.
	categoryRepository = "REPOSITORY COMMANDS"
)

var (
	// keyConfig is the key for the AWS config.
	keyConfig = "config"

	// keyCredential is the key for the credential.
	keyCredential = "credential"

	// keyEnviron is the key for the original environment snapshot.
	keyEnviron = "environ"
)
