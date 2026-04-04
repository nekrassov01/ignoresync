package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/nekrassov01/ignoresync"
	"github.com/nekrassov01/ignoresync/color"
	"github.com/nekrassov01/ignoresync/config"
	"github.com/nekrassov01/ignoresync/env"
	"github.com/nekrassov01/ignoresync/health"
	"github.com/nekrassov01/ignoresync/manager"
	"github.com/nekrassov01/ignoresync/operator"
	"github.com/nekrassov01/ignoresync/prompt"
	"github.com/nekrassov01/logger/integrations/awssdk"
	"github.com/nekrassov01/logger/log"
	"github.com/urfave/cli/v3"
)

// logger is a global logger instance used across the application.
var logger = log.NewLogger(log.NewCLIHandler(io.Discard))

var (
	// loglevel is the flag for log level.
	loglevel = &cli.StringFlag{
		Name:    "log-level",
		Aliases: []string{"l"},
		Usage:   "set log level",
		Sources: cli.EnvVars(ignoresync.EnvLogLevel),
		Value:   slog.LevelInfo.String(),
	}

	// profile is the flag for AWS profile.
	profile = &cli.StringFlag{
		Name:    "profile",
		Aliases: []string{"p"},
		Usage:   "set aws profile",
		Sources: cli.EnvVars(ignoresync.EnvAWSProfile),
	}

	// region is the flag for AWS region.
	region = &cli.StringFlag{
		Name:    "region",
		Aliases: []string{"r"},
		Usage:   "set aws region",
		Sources: cli.EnvVars(ignoresync.EnvAWSRegion),
	}

	// remote is the flag for git remote name.
	remote = &cli.StringFlag{
		Name:    "remote",
		Aliases: []string{"R"},
		Usage:   "set git remote name",
		Sources: cli.EnvVars(ignoresync.EnvRemoteName),
		Value:   ignoresync.DefaultRemoteName,
	}

	// dryrun is the flag for dry run mode.
	dryrun = &cli.BoolFlag{
		Name:    "dry-run",
		Aliases: []string{"d"},
		Usage:   "run without processing files",
	}

	// overwrite is the flag for overwrite mode.
	overwrite = &cli.BoolFlag{
		Name:    "overwrite",
		Aliases: []string{"o"},
		Usage:   "force overwrite without confirmation",
	}
)

// newCmd creates a new CLI command.
func newCmd(w, ew io.Writer) *cli.Command {
	return &cli.Command{
		Name:                  ignoresync.CommandName,
		Version:               ignoresync.Version(),
		Usage:                 "Your shadow repository for ignored files.",
		Description:           "Sync files ignored in the repository across machines without configuration.",
		HideHelpCommand:       true,
		EnableShellCompletion: true,
		Writer:                w,
		ErrWriter:             ew,
		Metadata:              map[string]any{},
		Commands: []*cli.Command{
			{
				Name:        "bootstrap",
				Usage:       "Bootstrap the environment.",
				Description: "Build the AWS resources as environment and activate the local machine at the same.\nThis process creates an S3 bucket and a KMS key in your AWS account, generates an\ncredential tied to your account and region combination, and stores the state derived\nfrom that key in the local keystore.",
				Category:    categoryGlobal,
				Before:      before,
				Action:      bootstrap,
				Flags:       []cli.Flag{loglevel, profile, region},
			},
			{
				Name:        "check",
				Usage:       "Check the environment.",
				Description: "Check if the environment is accessible. If this check is not passed,\nthe environment is not available.",
				Category:    categoryGlobal,
				Before:      before,
				Action:      check,
				Flags:       []cli.Flag{loglevel, profile},
			},
			{
				Name:        "activate",
				Usage:       "Activate the local machine.",
				Description: "Activate the local machine using the specified credential.\nPass a known ID to be stored in the local keystore.",
				Category:    categoryLocal,
				Before:      before,
				Action:      activate,
				Flags:       []cli.Flag{loglevel, profile},
			},
			{
				Name:        "deactivate",
				Usage:       "Deactivate the local machine.",
				Description: "Deactivate the specified credential for the local machine.\nDelete known ID stored in the local keystore.",
				Category:    categoryLocal,
				Before:      before,
				Action:      deactivate,
				Flags:       []cli.Flag{loglevel, profile},
			},
			{
				Name:        "list",
				Usage:       "List the key IDs.",
				Description: "List the key IDs from credentials stored in the local keystore.\nYou can also check whether they are active or not.",
				Category:    categoryLocal,
				Before:      before,
				Action:      list,
				Flags:       []cli.Flag{loglevel, profile},
			},
			{
				Name:        "rotate",
				Usage:       "Rotate the credential.",
				Description: "Rotate the credential in the local keystore. This process\ngenerates a new credential and updates the state accordingly.",
				Category:    categoryLocal,
				Before:      before,
				Action:      rotate,
				Flags:       []cli.Flag{loglevel, profile},
			},
			{
				Name:        "leave",
				Usage:       "Leave the environment.",
				Description: "Leave the environment by deleting the state. This process\nremoves all the credentials from the local keystore.",
				Category:    categoryLocal,
				Before:      before,
				Action:      leave,
				Flags:       []cli.Flag{loglevel, profile},
			},
			{
				Name:        "push",
				Usage:       "Push the local files.",
				Description: "Push the local files to the environment. Bundle the local files into a tar.gz,\nmultiple encrypted on the client side, and uploaded to the environment.",
				Category:    categoryRepository,
				Before:      before,
				Action:      push,
				Flags:       []cli.Flag{loglevel, profile, remote, dryrun},
			},
			{
				Name:        "pull",
				Usage:       "Pull the remote files.",
				Description: "Pull the remote files from the environment. The files are decrypted and extracted\nfrom tar.gz to the current directory. If there are differences in the same file,\nit will prompt you to confirm whether to overwrite it.",
				Category:    categoryRepository,
				Before:      before,
				Action:      pull,
				Flags:       []cli.Flag{loglevel, profile, remote, overwrite},
			},
			{
				Name:        "rm",
				Usage:       "Remove the remote files/patterns.",
				Description: "Remove the remote files/patterns uploaded to the environment.\nIf the key cannot be retrieved, the deletion will fail.",
				Category:    categoryRepository,
				Before:      before,
				Action:      rm,
				Flags:       []cli.Flag{loglevel, profile, remote},
			},
			{
				Name:        "set",
				Usage:       "Set the remote patterns.",
				Description: "Set the remote patterns uploaded to the environment.\nJSON array format is required.",
				Category:    categoryRepository,
				Before:      before,
				Action:      set,
				Flags:       []cli.Flag{loglevel, profile, remote},
			},
			{
				Name:        "preview",
				Usage:       "Preview the remote files/patterns.",
				Description: "Preview the remote files/patterns uploaded to the environment.\nThis command does not make any changes to the local files and\nis intended to be used for checking the remote state.",
				Category:    categoryRepository,
				Before:      before,
				Action:      preview,
				Flags:       []cli.Flag{loglevel, profile, remote},
			},
			{
				Name:        "rewrap",
				Usage:       "Rewrap the remote files/patterns.",
				Description: "Rewrap the remote files/patterns uploaded to the environment.\nRe-encrypt the existing files/patterns using the new key.",
				Category:    categoryRepository,
				Before:      before,
				Action:      rewrap,
				Flags:       []cli.Flag{loglevel, profile, remote},
			},
			{
				Name:        "clean",
				Usage:       "Clean up files in the current repository.",
				Description: "Clean up files in the current repository. This will only delete files\nthat match the remote target pattern in the local repository.",
				Category:    categoryRepository,
				Before:      before,
				Action:      clean,
				Flags:       []cli.Flag{loglevel, profile, remote},
			},
			{
				Name:        "run",
				Usage:       "Run command with setup and cleanup.",
				Description: "Run command with setup and cleanup. Pull the files before execution.\nIf the command completes or is aborted, the pulled files are cleaned up.",
				Category:    categoryRepository,
				Before:      before,
				Action:      run,
				Flags:       []cli.Flag{loglevel, profile, region, remote},
			},
		},
	}
}

func before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	// Snapshot the original environment before any SDK loading
	cmd.Metadata[keyEnviron] = os.Environ()

	// Parse log level
	var level slog.Level
	if err := level.UnmarshalText([]byte(cmd.String(loglevel.Name))); err != nil {
		level = slog.LevelInfo
	}

	// Set logger style
	s := log.Style1()
	s.Caller.Fullpath = true

	// Create logger for application
	logger = log.NewLogger(log.NewCLIHandler(
		cmd.ErrWriter,
		log.WithLabel(ignoresync.LogLabel),
		log.WithLevel(level),
		log.WithCaller(level <= slog.LevelDebug),
		log.WithStyle(s),
	))

	// Load AWS config with the specified profile and region
	cfg, err := config.LoadAWSConfig(ctx, cmd.String(profile.Name), cmd.String(region.Name))
	if err != nil {
		return nil, err
	}

	// Create logger for AWS SDK
	cfg.Logger = awssdk.NewLogger(log.NewCLIHandler(
		cmd.ErrWriter,
		log.WithLabel("SDK"),
		log.WithLevel(level),
		log.WithCaller(level <= slog.LevelDebug),
		log.WithStyle(s),
	))
	cfg.ClientLogMode = aws.LogRequest | aws.LogResponse | aws.LogRetries | aws.LogSigning | aws.LogDeprecatedUsage

	// Warn if running in CI mode
	cred := os.Getenv(ignoresync.EnvCredential)
	if cred != "" {
		msg := fmt.Sprintf("ci mode: environment variable %q is set: this is not recommended except for CI use", ignoresync.EnvCredential)
		logger.Warn(msg)
	}

	// Set metadata for commands
	cmd.Metadata[keyConfig] = cfg
	cmd.Metadata[keyCredential] = cred

	return ctx, nil
}

func bootstrap(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		logger.Warn("ci mode: skip bootstrapping")
		return nil
	}
	logger.Info("bootstrap: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Check if state already exists to prevent accidental bootstrapping with another credential.
	exist, err := man.CheckStateExist()
	if err != nil {
		return err
	}
	if exist {
		msg := "skip bootstrapping: state already exists in the keyring"
		logger.Warn(msg)
		return nil
	}

	// Prompt for overwrite
	msg := fmt.Sprintf("confirm your profile: account=%s region=%s: ok?", man.Account, man.Region)
	if _, err := prompt.Confirm(cmd.Writer, msg, "canceled"); err != nil {
		_, _ = fmt.Fprintln(cmd.Writer, color.Warn(err.Error()))
		return nil
	}

	// Generate credential
	id, key, err := manager.GenerateCredential()
	if err != nil {
		return err
	}

	// Derive data from credential
	state, err := man.GenerateState(id, key)
	if err != nil {
		return err
	}

	// Create deployer
	d := env.New(cmd.Writer, cfg)

	// Deploy environment
	if err := d.Deploy(ctx, state); err != nil {
		return err
	}

	// Store state in the keyring
	if err := man.StoreState(state); err != nil {
		return err
	}

	// Expose credential
	pref := color.Important(" IMPORTANT ")
	cred := color.Underline(manager.EncodeCredential(id, key))
	_, _ = fmt.Fprintf(cmd.Writer, "%s store your credential securely: %s\n", pref, cred)

	logger.Info("bootstrap: finished")
	return nil
}

func check(ctx context.Context, cmd *cli.Command) error {
	logger.Info("check: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Load stored state
	state, err := man.EnsureState(cmd.Metadata[keyCredential].(string))
	if err != nil {
		return err
	}
	cfg.Region = state.Region

	// Check environment
	c := health.New(cmd.Writer, cfg)
	if err := c.Check(ctx, state); err != nil {
		return err
	}

	logger.Info("check: finished")
	return nil
}

func activate(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		logger.Warn("ci mode: skip activation")
		return nil
	}
	logger.Info("activate: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Get credential from user input
	cred, err := prompt.Secret(cmd.Writer, "enter your credential:", manager.ValidateCredential)
	if err != nil {
		return err
	}

	// Decode credential
	id, key, err := manager.DecodeCredential(cred)
	if err != nil {
		return err
	}

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Add new credential to state if state already exists
	exist, err := man.CheckStateExist()
	if err != nil {
		return err
	}
	if exist {
		return man.AddCredential(id, key)
	}

	// Generate state from credential
	state, err := man.GenerateState(id, key)
	if err != nil {
		return err
	}

	// Create deployer
	d := env.New(cmd.Writer, cfg)

	// Check if environment exists
	deployed, err := d.CheckDeployed(ctx, state)
	if err != nil {
		return err
	}
	if !deployed {
		err := errors.New("environment does not exist in your current profile: verify account/region is correct and bootstrapping is complete")
		return env.NewEnvError(err)
	}

	// Store state in the keyring
	if err := man.StoreState(state); err != nil {
		return err
	}

	logger.Info("activate: finished")
	return nil
}

func deactivate(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		logger.Warn("ci mode: skip deactivation")
		return nil
	}
	logger.Info("deactivate: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Get credential from user input
	cred, err := prompt.Secret(cmd.Writer, "enter your credential:", manager.ValidateCredential)
	if err != nil {
		return err
	}

	// Decode credential
	id, key, err := manager.DecodeCredential(cred)
	if err != nil {
		return err
	}

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Update state by removing credential
	if err := man.RemoveCredential(id, key); err != nil {
		return err
	}

	logger.Info("deactivate: finished")
	return nil
}

func list(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		logger.Warn("ci mode: skip getting key IDs")
		return nil
	}
	logger.Info("list: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// List credentials
	keys, err := man.ListCredentials()
	if err != nil {
		return err
	}

	// Display key list as JSON
	v, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(cmd.Writer, string(v))

	logger.Info("list: finished")
	return nil
}

func rotate(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		logger.Warn("ci mode: skip rotation")
		return nil
	}
	logger.Info("rotate: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Generate credential
	id, key, err := manager.GenerateCredential()
	if err != nil {
		return err
	}

	// Add new credential to state
	if err := man.AddCredential(id, key); err != nil {
		return err
	}

	// Expose credential
	pref := color.Important(" IMPORTANT ")
	cred := color.Underline(manager.EncodeCredential(id, key))
	warn := color.Warn("rotate successfully, but the object is still encrypted with the old key: rewrap required")
	_, _ = fmt.Fprintf(cmd.Writer, "%s store your credential securely: %s\n", pref, cred)
	_, _ = fmt.Fprintln(cmd.Writer, warn)

	logger.Info("rotate: finished")
	return nil
}

func leave(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		logger.Warn("ci mode: skip leaving environment")
		return nil
	}
	logger.Info("leave: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Delete state from the keyring
	if err := man.DeleteState(); err != nil {
		return err
	}

	logger.Info("leave: finished")
	return nil
}

func push(ctx context.Context, cmd *cli.Command) error {
	logger.Info("push: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Load stored state
	state, err := man.EnsureState(cmd.Metadata[keyCredential].(string))
	if err != nil {
		return err
	}
	cfg.Region = state.Region

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Create operator
	o, err := operator.New(cmd.Writer, cwd, cmd.String(remote.Name), cfg)
	if err != nil {
		return err
	}

	// Pull target patterns from S3
	if _, err := o.PullPatterns(ctx, state); err != nil {
		return err
	}

	// Set dry run mode if specified
	if cmd.Bool(dryrun.Name) {
		o.SetDryrun(true)
	}

	// Push target files to S3
	if err := o.PushFiles(ctx, state); err != nil {
		return err
	}

	logger.Info("push: finished")
	return nil
}

func pull(ctx context.Context, cmd *cli.Command) error {
	logger.Info("pull: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Load stored state
	state, err := man.EnsureState(cmd.Metadata[keyCredential].(string))
	if err != nil {
		return err
	}
	cfg.Region = state.Region

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Create operator
	o, err := operator.New(cmd.Writer, cwd, cmd.String(remote.Name), cfg)
	if err != nil {
		return err
	}

	// Set dryrun if specified
	if cmd.Bool(dryrun.Name) {
		o.SetDryrun(true)
	}

	// Set overwrite when in CI mode or specified
	if shouldOverwrite(cmd) {
		o.SetOverwrite(true)
	}

	// Pull target files from S3
	if err := o.PullFiles(ctx, state); err != nil {
		return err
	}

	logger.Info("pull: finished")
	return nil
}

func rm(ctx context.Context, cmd *cli.Command) error {
	logger.Info("rm: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Load stored state
	state, err := man.EnsureState(cmd.Metadata[keyCredential].(string))
	if err != nil {
		return err
	}
	cfg.Region = state.Region

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Create operator
	o, err := operator.New(cmd.Writer, cwd, cmd.String(remote.Name), cfg)
	if err != nil {
		return err
	}

	// Delete remote files, patterns, and keys from S3
	if err := o.Delete(ctx, state); err != nil {
		return err
	}

	logger.Info("rm: finished")
	return nil
}

func set(ctx context.Context, cmd *cli.Command) error {
	logger.Info("set: starting")

	// Get command arguments
	arg := cmd.Args().Get(0)
	if len(arg) == 0 {
		return nil
	}

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Load stored state
	state, err := man.EnsureState(cmd.Metadata[keyCredential].(string))
	if err != nil {
		return err
	}
	cfg.Region = state.Region

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Create operator
	o, err := operator.New(cmd.Writer, cwd, cmd.String(remote.Name), cfg)
	if err != nil {
		return err
	}

	// Parse target patterns
	var target []string
	if err := json.Unmarshal([]byte(arg), &target); err != nil {
		return err
	}

	// Set target patterns
	o.SetPatterns(target)

	// Push target patterns to S3
	if err := o.PushPatterns(ctx, state, target); err != nil {
		return err
	}

	logger.Info("set: finished")
	return nil
}

func preview(ctx context.Context, cmd *cli.Command) error {
	logger.Info("preview: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Load stored state
	state, err := man.EnsureState(cmd.Metadata[keyCredential].(string))
	if err != nil {
		return err
	}
	cfg.Region = state.Region

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Create operator
	o, err := operator.New(cmd.Writer, cwd, cmd.String(remote.Name), cfg)
	if err != nil {
		return err
	}

	// Set dry run mode
	o.SetDryrun(true)

	// Preview target files
	if _, err := o.PullPatterns(ctx, state); err != nil {
		return err
	}

	// Preview target patterns
	if err := o.PullFiles(ctx, state); err != nil {
		return err
	}

	logger.Info("preview: finished")
	return nil
}

func rewrap(ctx context.Context, cmd *cli.Command) error {
	logger.Info("rewrap: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Load stored state
	state, err := man.EnsureState(cmd.Metadata[keyCredential].(string))
	if err != nil {
		return err
	}
	cfg.Region = state.Region

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Create operator
	o, err := operator.New(cmd.Writer, cwd, cmd.String(remote.Name), cfg)
	if err != nil {
		return err
	}

	// Perform rewrap
	if err := o.Rewrap(ctx, state); err != nil {
		return err
	}

	logger.Info("rewrap: finished")
	return nil
}

func clean(ctx context.Context, cmd *cli.Command) error {
	logger.Info("clean: starting")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Load stored state
	state, err := man.EnsureState(cmd.Metadata[keyCredential].(string))
	if err != nil {
		return err
	}
	cfg.Region = state.Region

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Create operator
	o, err := operator.New(cmd.Writer, cwd, cmd.String(remote.Name), cfg)
	if err != nil {
		return err
	}

	// Pull remote patterns
	patterns, err := o.PullPatterns(ctx, state)
	if err != nil {
		return err
	}
	if len(patterns) > 0 {
		o.SetPatterns(patterns)
	}

	// Cleanup files
	if err := o.CleanupFiles(); err != nil {
		return err
	}

	logger.Info("clean: finished")
	return nil
}

func run(ctx context.Context, cmd *cli.Command) error {
	logger.Info("run: starting")

	// Get command arguments
	args := cmd.Args().Slice()
	if len(args) == 0 {
		return nil
	}
	command := strings.Join(args, " ")

	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return err
	}

	// Load stored state
	state, err := man.EnsureState(cmd.Metadata[keyCredential].(string))
	if err != nil {
		return err
	}

	// Set region from state if not specified in flags
	if cmd.String(region.Name) == "" {
		cfg.Region = state.Region
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Create operator
	o, err := operator.New(cmd.Writer, cwd, cmd.String(remote.Name), cfg)
	if err != nil {
		return err
	}

	// Pull remote patterns
	patterns, err := o.PullPatterns(ctx, state)
	if err != nil {
		return err
	}
	if len(patterns) > 0 {
		o.SetPatterns(patterns)
	}

	// Pull remote files
	o.SetOverwrite(true)
	if err := o.PullFiles(ctx, state); err != nil {
		return err
	}

	// Ensure cleanup runs on normal exit, error, or signal
	defer func() {
		_ = o.CleanupFiles()
		logger.Info("run: finished")
	}()

	// Execute command
	environ := cmd.Metadata[keyEnviron].([]string)
	return o.Run(command, environ)
}

// shouldOverwrite sets the overwrite mode based on the command flags and environment.
func shouldOverwrite(cmd *cli.Command) bool {
	if cmd.Bool(overwrite.Name) {
		return true
	}
	if cred := cmd.Metadata[keyCredential].(string); cred != "" {
		logger.Warn("ci mode: force overwrite enabled")
		return true
	}
	return false
}
