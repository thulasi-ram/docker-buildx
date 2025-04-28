package plugin

import (
	"context"
	"io"
	"testing"

	"codeberg.org/6543/go-yaml2json"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func setupPluginTest(t *testing.T) *Plugin {
	t.Helper()

	cli.HelpPrinter = func(_ io.Writer, _ string, _ interface{}) {}
	got := New(func(_ context.Context) error { return nil }, "unknown")
	_ = got.App.Run(t.Context(), []string{"docker-buildx"})

	return got
}

func convertYamlToJson(t *testing.T, data string) string {
	t.Helper()

	r, err := yaml2json.Convert([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	return string(r)
}

func TestDefaultLogin(t *testing.T) {
	tests := []struct {
		name         string
		envs         map[string]string
		wantLogins   []string
		wantRegistry string
	}{
		{
			name: "default settings",
			wantLogins: []string{
				"https://index.docker.io/v1/",
			},
			wantRegistry: "https://index.docker.io/v1/",
		},
		{
			name: "only use login to auth to registries",
			envs: map[string]string{
				"PLUGIN_LOGINS": convertYamlToJson(t, `
- registry: https://index.docker.io/v1/
  username: docker_username
  password: docker_password
- registry: https://codeberg.org
  username: cb_username
  password: cb_password`),
			},
			wantLogins: []string{
				"https://index.docker.io/v1/",
				"https://codeberg.org",
			},
			wantRegistry: "https://index.docker.io/v1/",
		},
		{
			name: "mixed login settings",
			envs: map[string]string{
				"PLUGIN_LOGINS": convertYamlToJson(t, `
- registry: https://codeberg.org
  username: cb_username
  password: cb_password`),
				"DOCKER_REGISTRY": "https://quay.io",
			},
			wantLogins: []string{
				"https://codeberg.org",
			},
			wantRegistry: "https://quay.io",
		},
		{
			name: "ignore default registry",
			envs: map[string]string{
				"PLUGIN_LOGINS": convertYamlToJson(t, `
- registry: https://codeberg.org
  username: cb_username
  password: cb_password`),
				"DOCKER_REGISTRY": "",
			},
			wantLogins: []string{
				"https://codeberg.org",
			},
			wantRegistry: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}

			got := setupPluginTest(t)

			_ = got.InitSettings()
			err := got.Validate()

			assert.NoError(t, err)
			assert.Equal(t, tt.wantRegistry, got.Settings.DefaultLogin.Registry)

			for i, login := range got.Settings.Logins {
				assert.Equal(t, tt.wantLogins[i], login.Registry)
			}
		})
	}
}

func TestWriteBuildkitConfig(t *testing.T) {
	tests := []struct {
		name       string
		envs       map[string]string
		wantConfig string
	}{
		{
			name:       "default settings",
			wantConfig: "",
		},
		{
			name: "buildkit debug enabled",
			envs: map[string]string{
				"PLUGIN_BUILDKIT_DEBUG": "true",
			},
			wantConfig: "debug = true\n",
		},
		{
			name: "mirror configured",
			envs: map[string]string{
				"DOCKER_PLUGIN_MIRROR": "mirror.example.com",
			},
			wantConfig: "[registry]\n[registry.'docker.io']\nmirrors = ['mirror.example.com']\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}

			got := setupPluginTest(t)

			_ = got.InitSettings()
			err := got.Validate()

			assert.NoError(t, err)
			assert.Equal(t, tt.wantConfig, got.Settings.Daemon.BuildkitConfig)
		})
	}
}

func TestSecretsFlag(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want []string
	}{
		{
			name: "parse secrets list with escape",
			envs: map[string]string{
				"PLUGIN_SECRETS": "id=raw_file_secret\\,src=file.txt,id=SECRET_TOKEN",
			},
			want: []string{
				"id=raw_file_secret,src=file.txt",
				"id=SECRET_TOKEN",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}

			got := setupPluginTest(t)

			assert.EqualValues(t, tt.want, got.Settings.Build.Secrets)
		})
	}
}

func TestCacheFromFlag(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want []string
	}{
		{
			name: "simple escape",
			envs: map[string]string{
				"PLUGIN_CACHE_FROM": `type=registry\,ref=example,foo=bar`,
			},
			want: []string{
				"type=registry,ref=example",
				"foo=bar",
			},
		},
		{
			name: "double escape",
			envs: map[string]string{
				"PLUGIN_CACHE_FROM": "type=registry\\,ref=example,foo=bar",
			},
			want: []string{
				"type=registry,ref=example",
				"foo=bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}

			got := setupPluginTest(t)

			assert.ElementsMatch(t, tt.want, got.Settings.Build.CacheFrom)
		})
	}
}

func TestBuildArgsFlag(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want map[string]string
	}{
		{
			name: "not nil",
			envs: map[string]string{},
			want: map[string]string{},
		},
		{
			name: "parse args",
			envs: map[string]string{
				"PLUGIN_BUILD_ARGS": `{"arg1": "value1", "arg2": "value2"}`,
			},
			want: map[string]string{
				"arg1": "value1",
				"arg2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}

			got := setupPluginTest(t)

			assert.Equal(t, tt.want, got.Settings.Build.Args)
		})
	}
}
