package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/nekrassov01/ignoresync/color"
	"github.com/nekrassov01/ignoresync/config"
	"github.com/nekrassov01/ignoresync/env"
	"github.com/nekrassov01/ignoresync/health"
	"github.com/nekrassov01/ignoresync/log"
	"github.com/nekrassov01/ignoresync/manager"
	"github.com/nekrassov01/ignoresync/operator"
	"github.com/nekrassov01/ignoresync/params"
	"github.com/nekrassov01/ignoresync/prompt"
	"github.com/urfave/cli/v3"
)

func before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	// Snapshot the original environment before any SDK loading
	cmd.Metadata[keyEnviron] = os.Environ()

	// Set application logger with the specified log level
	log.SetAppLogger(cmd.ErrWriter, cmd.String(loglevel.Name))

	// Warn if running in CI mode
	cred := os.Getenv(params.EnvCredential)
	if cred != "" {
		msg := fmt.Sprintf("ci mode: environment variable %q is set: this is not recommended except for CI use", params.EnvCredential)
		log.Logger.Warn(msg)
	}
	cmd.Metadata[keyCredential] = cred

	// Load AWS config with the specified profile and region
	cfg, err := config.LoadAWSConfig(ctx, cmd.String(profile.Name), cmd.String(region.Name))
	if err != nil {
		return nil, err
	}
	log.SetSDKLogger(cmd.ErrWriter, cmd.String(loglevel.Name), &cfg)
	cmd.Metadata[keyConfig] = cfg

	return ctx, nil
}

func bootstrap(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		log.Logger.Warn("ci mode: skip bootstrapping")
		return nil
	}
	log.Logger.Info("bootstrap: starting")

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
		log.Logger.Warn(msg)
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
	cred := color.Private(manager.EncodeCredential(id, key))
	_, _ = fmt.Fprintf(cmd.Writer, "%s store your credential securely: %s\n", pref, cred)

	log.Logger.Info("bootstrap: finished")
	return nil
}

func check(ctx context.Context, cmd *cli.Command) error {
	log.Logger.Info("check: starting")

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

	// Check environment
	c := health.New(cmd.Writer, cfg)
	if err := c.Check(ctx, state); err != nil {
		return err
	}

	log.Logger.Info("check: finished")
	return nil
}

func activate(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		log.Logger.Warn("ci mode: skip activation")
		return nil
	}
	log.Logger.Info("activate: starting")

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
		if err := man.AddCredential(id, key); err != nil {
			return err
		}
		log.Logger.Info("activate: finished")
		return nil
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

	log.Logger.Info("activate: finished")
	return nil
}

func deactivate(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		log.Logger.Warn("ci mode: skip deactivation")
		return nil
	}
	log.Logger.Info("deactivate: starting")

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

	log.Logger.Info("deactivate: finished")
	return nil
}

func list(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		log.Logger.Warn("ci mode: skip getting key IDs")
		return nil
	}
	log.Logger.Info("list: starting")

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

	log.Logger.Info("list: finished")
	return nil
}

func rotate(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		log.Logger.Warn("ci mode: skip rotation")
		return nil
	}
	log.Logger.Info("rotate: starting")

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
	cred := color.Private(manager.EncodeCredential(id, key))
	warn := color.Warn("rotate successfully, but the object is still encrypted with the old key: rewrap required")
	_, _ = fmt.Fprintf(cmd.Writer, "%s store your credential securely: %s\n", pref, cred)
	_, _ = fmt.Fprintln(cmd.Writer, warn)

	log.Logger.Info("rotate: finished")
	return nil
}

func leave(ctx context.Context, cmd *cli.Command) error {
	// Check if running in CI mode
	if cmd.Metadata[keyCredential].(string) != "" {
		log.Logger.Warn("ci mode: skip leaving environment")
		return nil
	}
	log.Logger.Info("leave: starting")

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

	log.Logger.Info("leave: finished")
	return nil
}

func push(ctx context.Context, cmd *cli.Command) error {
	log.Logger.Info("push: starting")

	// Create operator and load state
	state, o, err := newOperator(ctx, cmd)
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

	log.Logger.Info("push: finished")
	return nil
}

func pull(ctx context.Context, cmd *cli.Command) error {
	log.Logger.Info("pull: starting")

	// Create operator and load state
	state, o, err := newOperator(ctx, cmd)
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

	log.Logger.Info("pull: finished")
	return nil
}

func rm(ctx context.Context, cmd *cli.Command) error {
	log.Logger.Info("rm: starting")

	// Create operator and load state
	state, o, err := newOperator(ctx, cmd)
	if err != nil {
		return err
	}

	// Delete remote files, patterns, and keys from S3
	if err := o.Delete(ctx, state); err != nil {
		return err
	}

	log.Logger.Info("rm: finished")
	return nil
}

func set(ctx context.Context, cmd *cli.Command) error {
	log.Logger.Info("set: starting")

	// Get command arguments
	arg := cmd.Args().Get(0)
	if len(arg) == 0 {
		return nil
	}

	// Create operator and load state
	state, o, err := newOperator(ctx, cmd)
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

	log.Logger.Info("set: finished")
	return nil
}

func preview(ctx context.Context, cmd *cli.Command) error {
	log.Logger.Info("preview: starting")

	// Create operator and load state
	state, o, err := newOperator(ctx, cmd)
	if err != nil {
		return err
	}

	// Set preview mode
	o.SetPreview(true)

	// Preview target files
	if _, err := o.PullPatterns(ctx, state); err != nil {
		return err
	}

	// Preview target patterns
	if err := o.PullFiles(ctx, state); err != nil {
		return err
	}

	log.Logger.Info("preview: finished")
	return nil
}

func rewrap(ctx context.Context, cmd *cli.Command) error {
	log.Logger.Info("rewrap: starting")

	// Create operator and load state
	state, o, err := newOperator(ctx, cmd)
	if err != nil {
		return err
	}

	// Perform rewrap
	if err := o.Rewrap(ctx, state); err != nil {
		return err
	}

	log.Logger.Info("rewrap: finished")
	return nil
}

func clean(ctx context.Context, cmd *cli.Command) error {
	log.Logger.Info("clean: starting")

	// Create operator and load state
	state, o, err := newOperator(ctx, cmd)
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

	log.Logger.Info("clean: finished")
	return nil
}

func run(ctx context.Context, cmd *cli.Command) error {
	log.Logger.Info("run: starting")

	// Get command arguments
	args := cmd.Args().Slice()
	if len(args) == 0 {
		return nil
	}
	command := strings.Join(args, " ")

	// Create operator and load state
	state, o, err := newOperator(ctx, cmd)
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
		log.Logger.Info("run: finished")
	}()

	// Execute command
	environ := cmd.Metadata[keyEnviron].([]string)
	return o.Run(ctx, command, environ)
}

// newOperator creates a new operator and loads the state from the keyring.
func newOperator(ctx context.Context, cmd *cli.Command) (*manager.State, *operator.Operator, error) {
	// Load AWS config
	cfg := cmd.Metadata[keyConfig].(aws.Config)

	// Create manager
	man, err := manager.New(ctx, cmd.Writer, cfg)
	if err != nil {
		return nil, nil, err
	}

	// Load stored state
	state, err := man.EnsureState(cmd.Metadata[keyCredential].(string))
	if err != nil {
		return nil, nil, err
	}

	// Set region from state if not specified in flags
	if cmd.String(region.Name) == "" {
		cfg.Region = state.Region
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}

	// Create operator
	o, err := operator.New(cmd.Writer, cwd, cmd.String(remote.Name), cfg)
	if err != nil {
		return nil, nil, err
	}

	return state, o, nil
}

// shouldOverwrite sets the overwrite mode based on the command flags and environment.
func shouldOverwrite(cmd *cli.Command) bool {
	if cmd.Bool(overwrite.Name) {
		return true
	}
	if cred := cmd.Metadata[keyCredential].(string); cred != "" {
		log.Logger.Warn("ci mode: force overwrite enabled")
		return true
	}
	return false
}
