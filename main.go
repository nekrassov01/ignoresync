package main

import (
	"context"
	"os"

	"github.com/nekrassov01/ignoresync/internal/cmd"
	"github.com/nekrassov01/ignoresync/internal/log"
)

func main() {
	ctx := context.Background()
	cli := cmd.New(os.Stdout, os.Stderr)
	if err := cli.Run(ctx, os.Args); err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}
}
