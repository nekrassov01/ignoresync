package cmd

import (
	"io"
	"log/slog"

	"github.com/nekrassov01/ignoresync/internal/params"
	"github.com/nekrassov01/ignoresync/internal/version"
	"github.com/urfave/cli/v3"
)

var (
	// loglevel is the flag for log level.
	loglevel = &cli.StringFlag{
		Name:    "log-level",
		Aliases: []string{"l"},
		Usage:   "set log level",
		Sources: cli.EnvVars(params.EnvLogLevel),
		Value:   slog.LevelInfo.String(),
	}

	// profile is the flag for AWS profile.
	profile = &cli.StringFlag{
		Name:    "profile",
		Aliases: []string{"p"},
		Usage:   "set aws profile",
		Sources: cli.EnvVars(params.EnvAWSProfile),
	}

	// region is the flag for AWS region.
	region = &cli.StringFlag{
		Name:    "region",
		Aliases: []string{"r"},
		Usage:   "set aws region",
		Sources: cli.EnvVars(params.EnvAWSRegion),
	}

	// remote is the flag for git remote name.
	remote = &cli.StringFlag{
		Name:    "remote",
		Aliases: []string{"R"},
		Usage:   "set git remote name",
		Sources: cli.EnvVars(params.EnvRemoteName),
		Value:   params.DefaultRemoteName,
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

// New creates a new CLI command.
func New(w, ew io.Writer) *cli.Command {
	return &cli.Command{
		Name:                  params.CommandName,
		Version:               version.Version(),
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
				Flags:       []cli.Flag{loglevel, profile, region},
			},
			{
				Name:        "activate",
				Usage:       "Activate the local machine.",
				Description: "Activate the local machine using the specified credential.\nPass a known ID to be stored in the local keystore.",
				Category:    categoryLocal,
				Before:      before,
				Action:      activate,
				Flags:       []cli.Flag{loglevel, profile, region},
			},
			{
				Name:        "deactivate",
				Usage:       "Deactivate the local machine.",
				Description: "Deactivate the specified credential for the local machine.\nDelete known ID stored in the local keystore.",
				Category:    categoryLocal,
				Before:      before,
				Action:      deactivate,
				Flags:       []cli.Flag{loglevel, profile, region},
			},
			{
				Name:        "list",
				Usage:       "List the key IDs.",
				Description: "List the key IDs from credentials stored in the local keystore.\nYou can also check whether they are active or not.",
				Category:    categoryLocal,
				Before:      before,
				Action:      list,
				Flags:       []cli.Flag{loglevel, profile, region},
			},
			{
				Name:        "rotate",
				Usage:       "Rotate the credential.",
				Description: "Rotate the credential in the local keystore. This process\ngenerates a new credential and updates the state accordingly.",
				Category:    categoryLocal,
				Before:      before,
				Action:      rotate,
				Flags:       []cli.Flag{loglevel, profile, region},
			},
			{
				Name:        "leave",
				Usage:       "Leave the environment.",
				Description: "Leave the environment by deleting the state. This process\nremoves all the credentials from the local keystore.",
				Category:    categoryLocal,
				Before:      before,
				Action:      leave,
				Flags:       []cli.Flag{loglevel, profile, region},
			},
			{
				Name:        "push",
				Usage:       "Push the local files.",
				Description: "Push the local files to the environment. Bundle the local files into a tar.gz,\nmultiple encrypted on the client side, and uploaded to the environment.",
				Category:    categoryRepository,
				Before:      before,
				Action:      push,
				Flags:       []cli.Flag{loglevel, profile, region, remote, dryrun},
			},
			{
				Name:        "pull",
				Usage:       "Pull the remote files.",
				Description: "Pull the remote files from the environment. The files are decrypted and extracted\nfrom tar.gz to the current directory. If there are differences in the same file,\nit will prompt you to confirm whether to overwrite it.",
				Category:    categoryRepository,
				Before:      before,
				Action:      pull,
				Flags:       []cli.Flag{loglevel, profile, region, remote, overwrite},
			},
			{
				Name:        "rm",
				Usage:       "Remove the remote files/patterns.",
				Description: "Remove the remote files/patterns uploaded to the environment.\nIf the key cannot be retrieved, the deletion will fail.",
				Category:    categoryRepository,
				Before:      before,
				Action:      rm,
				Flags:       []cli.Flag{loglevel, profile, region, remote},
			},
			{
				Name:        "set",
				Usage:       "Set the remote patterns.",
				Description: "Set the remote patterns uploaded to the environment.\nJSON array format is required.",
				Category:    categoryRepository,
				Before:      before,
				Action:      set,
				Flags:       []cli.Flag{loglevel, profile, region, remote},
			},
			{
				Name:        "preview",
				Usage:       "Preview the remote files/patterns.",
				Description: "Preview the remote files/patterns uploaded to the environment.\nThis command does not make any changes to the local files and\nis intended to be used for checking the remote state.",
				Category:    categoryRepository,
				Before:      before,
				Action:      preview,
				Flags:       []cli.Flag{loglevel, profile, region, remote},
			},
			{
				Name:        "rewrap",
				Usage:       "Rewrap the remote files/patterns.",
				Description: "Rewrap the remote files/patterns uploaded to the environment.\nRe-encrypt the existing files/patterns using the new key.",
				Category:    categoryRepository,
				Before:      before,
				Action:      rewrap,
				Flags:       []cli.Flag{loglevel, profile, region, remote},
			},
			{
				Name:        "clean",
				Usage:       "Clean up files in the current repository.",
				Description: "Clean up files in the current repository. This will only delete files\nthat match the remote target pattern in the local repository.",
				Category:    categoryRepository,
				Before:      before,
				Action:      clean,
				Flags:       []cli.Flag{loglevel, profile, region, remote},
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
