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
	// keyConf is the key for the AWS config.
	keyConf = "config"

	// keyCred is the key for the credential.
	keyCred = "credential"
)
