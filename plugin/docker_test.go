package plugin

import (
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestCommandBuilder(t *testing.T) {
	tests := []struct {
		name      string
		daemon    Daemon
		input     string
		wantedLen int
		skip      bool
		excuse    string
	}{
		{
			name:      "Single driver-opt value",
			daemon:    Daemon{},
			input:     "no_proxy=*.mydomain",
			wantedLen: 1,
		},
		{
			name:      "Single driver-opt value with comma",
			input:     "no_proxy=.mydomain,.sub.domain.com",
			wantedLen: 1,
			skip:      true,
			excuse:    "Can be enabled whenever #94 is fixed.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(fmt.Printf("%v skipped. %v", tt.name, tt.excuse))
			}
			// prepare test values to mock plugin call with settings
			t.Setenv("PLUGIN_BUILDKIT_DRIVEROPT", tt.input)

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
						Destination: &tt.daemon.BuildkitDriverOpt,
					},
				},
				Action: nil,
			}

			// need to run the app to resolve the flags
			_ = app.Run(context.Background(), nil)

			// call the commandBuilder to prepare the cmd with its args
			_ = commandBuilder(tt.daemon, "", false)

			assert.Len(t, tt.daemon.BuildkitDriverOpt, tt.wantedLen)
		})
	}
}

func TestCommandBuilderRemoteBuilders(t *testing.T) {
	tests := []struct {
		name   string
		daemon Daemon
		input  string
		skip   bool
		excuse string
	}{
		{
			name:  "Single 'local' remote builder",
			input: "local",
		},
		{
			name:  "Single remote builder",
			input: "root@example.org",
		},
		{
			name:  "Remote and local builders",
			input: "local,root@example.org",
		},
		{
			name:  "Two remote builders",
			input: "root@example.com,root@example.org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(fmt.Printf("%v skipped. %v", tt.name, tt.excuse))
			}
			// prepare test values to mock plugin call with settings
			t.Setenv("PLUGIN_REMOTE_BUILDERS", tt.input)

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
						Destination: &tt.daemon.RemoteBuilders,
					},
				},
				Action: nil,
			}

			// need to run the app to resolve the flags
			_ = app.Run(context.Background(), nil)

			var cmd *exec.Cmd
			for _, host := range tt.daemon.RemoteBuilders {
				cmd = commandBuilder(tt.daemon, host, false)
				assert.NotContains(t, cmd.Args, "local")
				assert.NotContains(t, cmd.Args, "--append")
				if host == "local" {
					assert.NotContains(t, cmd.Args, "--driver=docker-container")
				} else {
					assert.Contains(t, cmd.Args, "ssh://"+host)
				}
				cmd = commandBuilder(tt.daemon, host, true)
				assert.Contains(t, cmd.Args, "--append")
			}
		})
	}
}
