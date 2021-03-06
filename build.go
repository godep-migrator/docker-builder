package main

import (
	"github.com/rafecolton/docker-builder/builder"

	"github.com/codegangsta/cli"
)

func build(c *cli.Context) {
	builder.SkipPush = c.Bool("skip-push")
	builderfile := c.Args().First()
	if builderfile == "" {
		builderfile = "Bobfile"
	}

	bob, err := builder.NewBuilder(Logger, true)
	if err != nil {
		exitErr(61, "unable to build", err)
	}

	config, err := builder.NewTrustedFilePath(builderfile, ".")
	if err != nil {
		exitErr(1, "unable to create build config", err)
	}

	if err := bob.Build(config); err != nil {
		if builder.IsSanitizeError(err) {
			exitErr(err.ExitCode(), "unable to build", map[string]interface{}{
				"error":    err,
				"filename": err.(*builder.SanitizeError).Filename,
			})
		} else {
			exitErr(err.ExitCode(), "unable to build", err)
		}
	}
}
