package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestCommandBuilder(t *testing.T) {
	tests := []struct {
		Name      string
		Daemon    Daemon
		Input     string
		WantedLen int
		Skip      bool
		Excuse    string
	}{
		{
			Name:      "Single driver-opt value",
			Daemon:    Daemon{},
			Input:     "no_proxy=*.mydomain",
			WantedLen: 1,
		},
		{
			Name:      "Single driver-opt value with comma",
			Input:     "no_proxy=.mydomain,.sub.domain.com",
			WantedLen: 1,
			Skip:      true,
			Excuse:    "Can be enabled whenever #94 is fixed.",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.Skip {
				t.Skip(fmt.Printf("%v skipped. %v", test.Name, test.Excuse))
			}
			// prepare test values to mock plugin call with settings
			os.Setenv("PLUGIN_BUILDKIT_DRIVEROPT", test.Input)

			// create dummy cli app to reproduce the issue
			app := &cli.Command{
				Name:    "dummy App",
				Usage:   "testing inputs",
				Version: "0.0.1",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:        "daemon.buildkit-driveropt",
						Sources:     cli.EnvVars("PLUGIN_BUILDKIT_DRIVEROPT"),
						Usage:       "adds optional driver-ops args like 'env.http_proxy'",
						Destination: &test.Daemon.BuildkitDriverOpt,
					},
				},
				Action: nil,
			}

			// need to run the app to resolve the flags
			_ = app.Run(context.Background(), nil)

			// call the commandBuilder to prepare the cmd with its args
			_ = commandBuilder(test.Daemon, "", false)

			assert.Len(t, test.Daemon.BuildkitDriverOpt, test.WantedLen)
		})
	}
}

func TestCommandBuilderRemoteBuilders(t *testing.T) {
	tests := []struct {
		Name   string
		Daemon Daemon
		Input  string
		Skip   bool
		Excuse string
	}{
		{
			Name:  "Single 'local' remote builder",
			Input: "local",
		},
		{
			Name:  "Single remote builder",
			Input: "root@example.org",
		},
		{
			Name:  "Remote and local builders",
			Input: "local,root@example.org",
		},
		{
			Name:  "Two remote builders",
			Input: "root@example.com,root@example.org",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.Skip {
				t.Skip(fmt.Printf("%v skipped. %v", test.Name, test.Excuse))
			}
			// prepare test values to mock plugin call with settings
			os.Setenv("PLUGIN_REMOTE_BUILDERS", test.Input)

			// create dummy cli app to reproduce the issue
			app := &cli.Command{
				Name:    "dummy App",
				Usage:   "testing inputs",
				Version: "0.0.1",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:        "daemon.remote-builder",
						Sources:     cli.EnvVars("PLUGIN_REMOTE_BUILDERS"),
						Usage:       "adds remote builders to be used by buildx",
						Destination: &test.Daemon.RemoteBuilders,
					},
				},
				Action: nil,
			}

			// need to run the app to resolve the flags
			_ = app.Run(context.Background(), nil)

			var cmd *exec.Cmd
			for _, host := range test.Daemon.RemoteBuilders {
				cmd = commandBuilder(test.Daemon, host, false)
				assert.NotContains(t, cmd.Args, "local")
				assert.NotContains(t, cmd.Args, "--append")
				if host == "local" {
					assert.NotContains(t, cmd.Args, "--driver=docker-container")
				} else {
					assert.Contains(t, cmd.Args, "ssh://"+host)
				}
				cmd = commandBuilder(test.Daemon, host, true)
				assert.Contains(t, cmd.Args, "--append")
			}
		})
	}
}
