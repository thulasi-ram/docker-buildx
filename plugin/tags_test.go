package plugin

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_stripTagPrefix(t *testing.T) {
	tests := []struct {
		name   string
		before string
		after  string
	}{
		{
			name:   "strip refs/tags prefix",
			before: "refs/tags/1.0.0",
			after:  "1.0.0",
		},
		{
			name:   "strip refs/tags prefix with v",
			before: "refs/tags/v1.0.0",
			after:  "v1.0.0",
		},
		{
			name:   "no prefix to strip",
			before: "v1.0.0",
			after:  "v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, want := stripTagPrefix(tt.before), tt.after
			if got != want {
				t.Errorf("Got tag %s, want %s", got, want)
			}
		})
	}
}

func TestDefaultTags(t *testing.T) {
	tests := []struct {
		name       string
		defaultTag string
		before     string
		after      []string
	}{
		{
			name:       "no tag event",
			defaultTag: "latest",
			before:     "",
			after:      []string{"latest"},
		},
		{
			name:       "master branch",
			defaultTag: "latest",
			before:     "refs/heads/master",
			after:      []string{"latest"},
		},
		{
			name:       "semver tag without v prefix",
			defaultTag: "latest",
			before:     "refs/tags/0.9.0",
			after:      []string{"0.9", "0.9.0"},
		},
		{
			name:       "semver tag without v prefix major version",
			defaultTag: "latest",
			before:     "refs/tags/1.0.0",
			after:      []string{"1", "1.0", "1.0.0"},
		},
		{
			name:       "semver tag with v prefix",
			defaultTag: "latest",
			before:     "refs/tags/v1.0.0",
			after:      []string{"1", "1.0", "1.0.0"},
		},
		{
			name:       "release candidate tag",
			defaultTag: "latest",
			before:     "refs/tags/v1.2.3-rc1",
			after:      []string{"1.2.3-rc1"},
		},
		{
			name:       "alpha release tag",
			defaultTag: "latest",
			before:     "refs/tags/v1.0.0-alpha.1",
			after:      []string{"1.0.0-alpha.1"},
		},
		{
			name:       "date version with v prefix",
			defaultTag: "latest",
			before:     "refs/tags/v20221221",
			after:      []string{"20221221", "20221221.0", "20221221.0.0"},
		},
		{
			name:       "date version with hyphens",
			defaultTag: "latest",
			before:     "refs/tags/v2022-12-21",
			after:      []string{"2022.0.0-12-21"},
		},
		{
			name:       "date tag without v prefix",
			defaultTag: "latest",
			before:     "refs/tags/20221221",
			after:      []string{"20221221"},
		},
		{
			name:       "date tag with hyphens",
			defaultTag: "latest",
			before:     "refs/tags/2022-12-21",
			after:      []string{"2022-12-21"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := DefaultTags(tt.before, tt.defaultTag)
			if err != nil {
				t.Error(err)

				return
			}

			got, want := tags, tt.after
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Got tag %v, want %v", got, want)
			}
		})
	}
}

func TestDefaultTagsError(t *testing.T) {
	tests := []struct {
		name       string
		defaultTag string
		before     string
	}{
		{
			name:       "invalid version prefix",
			defaultTag: "latest",
			before:     "refs/tags/x1.0.0",
		},
		{
			name:       "invalid version format",
			defaultTag: "latest",
			before:     "refs/tags/2a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := DefaultTags(tt.before, tt.defaultTag)
			if err == nil {
				t.Errorf("Expect tag error for %s, got tags %v", tt, tags)
			}
		})
	}
}

func TestDefaultTagsuffix(t *testing.T) {
	tests := []struct {
		name       string
		before     string
		suffix     string
		after      []string
		defaultTag string
	}{
		{
			name:       "Default tag without suffix",
			defaultTag: "latest",
			after:      []string{"latest"},
		},
		{
			name:       "Overridden default tag without suffix",
			defaultTag: "next",
			after:      []string{"next"},
		},
		{
			name:       "Generate version",
			defaultTag: "latest",
			before:     "refs/tags/v1.0.0",
			after: []string{
				"1",
				"1.0",
				"1.0.0",
			},
		},
		{
			name:       "Generate version with overridden default tag",
			defaultTag: "next",
			before:     "refs/tags/v1.0.0",
			after: []string{
				"1",
				"1.0",
				"1.0.0",
			},
		},
		{
			name:       "Default tag with suffix (linux-amd64)",
			defaultTag: "latest",
			suffix:     "linux-amd64",
			after:      []string{"linux-amd64"},
		},
		{
			name:       "Overridden default tag with suffix (linux-amd64)",
			defaultTag: "next",
			suffix:     "linux-amd64",
			after:      []string{"linux-amd64"},
		},
		{
			name:       "Generate version with suffix (linux-amd64)",
			defaultTag: "latest",
			before:     "refs/tags/v1.0.0",
			suffix:     "linux-amd64",
			after: []string{
				"1-linux-amd64",
				"1.0-linux-amd64",
				"1.0.0-linux-amd64",
			},
		},
		{
			name:       "Generate version with suffix (linux-amd64) and overridden default tag (next)",
			defaultTag: "next",
			before:     "refs/tags/v1.0.0",
			suffix:     "linux-amd64",
			after: []string{
				"1-linux-amd64",
				"1.0-linux-amd64",
				"1.0.0-linux-amd64",
			},
		},
		{
			name:       "Default tag with suffix (nanoserver)",
			defaultTag: "latest",
			suffix:     "nanoserver",
			after:      []string{"nanoserver"},
		},
		{
			name:       "Overridden default tag with suffix (nanoserver)",
			defaultTag: "next",
			suffix:     "nanoserver",
			after:      []string{"nanoserver"},
		},
		{
			name:       "Generate version with suffix (nanoserver)",
			defaultTag: "latest",
			before:     "refs/tags/v1.9.2",
			suffix:     "nanoserver",
			after: []string{
				"1-nanoserver",
				"1.9-nanoserver",
				"1.9.2-nanoserver",
			},
		},
		{
			name:       "Generate version with suffix (nanoserver) and overridden default tag (next)",
			defaultTag: "latest",
			before:     "refs/tags/v1.9.2",
			suffix:     "nanoserver",
			after: []string{
				"1-nanoserver",
				"1.9-nanoserver",
				"1.9.2-nanoserver",
			},
		},
		{
			name:       "Generate version with suffix (zero-padded version)",
			defaultTag: "latest",
			before:     "refs/tags/v18.06.0",
			suffix:     "nanoserver",
			after: []string{
				"18-nanoserver",
				"18.6-nanoserver",
				"18.6.0-nanoserver",
			},
		},
		{
			name:       "Generate version with suffix (zero-padded version) with overridden default tag (next)",
			defaultTag: "next",
			before:     "refs/tags/v18.06.0",
			suffix:     "nanoserver",
			after: []string{
				"18-nanoserver",
				"18.6-nanoserver",
				"18.6.0-nanoserver",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := DefaultTagSuffix(tt.before, tt.defaultTag, tt.suffix)
			if assert.NoError(t, err) {
				assert.EqualValues(t, tt.after, tags)
			}
		})
	}
}

func Test_stripHeadPrefix(t *testing.T) {
	type args struct {
		ref string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "strip refs/heads prefix",
			args: args{
				ref: "refs/heads/master",
			},
			want: "master",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripHeadPrefix(tt.args.ref); got != tt.want {
				t.Errorf("stripHeadPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUseDefaultTag(t *testing.T) {
	type args struct {
		ref           string
		defaultBranch string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "latest tag for default branch",
			args: args{
				ref:           "refs/heads/master",
				defaultBranch: "master",
			},
			want: true,
		},
		{
			name: "build from tags",
			args: args{
				ref:           "refs/tags/v1.0.0",
				defaultBranch: "master",
			},
			want: true,
		},
		{
			name: "skip build for not default branch",
			args: args{
				ref:           "refs/heads/develop",
				defaultBranch: "master",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UseDefaultTag(tt.args.ref, tt.args.defaultBranch); got != tt.want {
				t.Errorf("%q. UseDefaultTag() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_isSingleTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{
			name: "latest tag",
			tag:  "latest",
			want: true,
		},
		{
			name: "leading space",
			tag:  " latest",
			want: false,
		},
		{
			name: "mixed case with underscores",
			tag:  "LaTest__Hi",
			want: true,
		},
		{
			name: "all valid characters",
			tag:  "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ__.-0123456789",
			want: true,
		},
		{
			name: "special characters",
			tag:  "_wierd.but-ok1",
			want: true,
		},
		{
			name: "trailing space",
			tag:  "latest ",
			want: false,
		},
		{
			name: "multiple tags",
			tag:  "latest,next",
			want: false,
		},
		{
			name: "empty tag",
			tag:  "",
			want: true, // important to allow omitting 'latest' tag when using auto_tag: true
		},
		// more tests to be added, once the validation is more powerful
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSingleTag(tt.tag); got != tt.want {
				t.Errorf("%q. isSingleTag() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestTagsFile(t *testing.T) {
	tests := []struct {
		name         string
		envs         map[string]string
		tagsContent  string
		initialTags  []string
		expectedTags []string
		fileExists   bool
		wantErr      bool
	}{
		{
			name:         "tags on separate lines",
			tagsContent:  "tag1\ntag2\ntag3",
			expectedTags: []string{"tag1", "tag2", "tag3"},
			wantErr:      false,
		},
		{
			name:         "tags with extra whitespace",
			tagsContent:  "  tag1  \n tag2 \n\ntag3  ",
			expectedTags: []string{"tag1", "tag2", "tag3"},
			wantErr:      false,
		},
		{
			name:    "non-existent file",
			wantErr: true,
		},
		{
			name: "tags file overrides existing tags",
			envs: map[string]string{
				"PLUGIN_TAGS": "tag1,tag2",
			},
			tagsContent:  "new-tag1\nnew-tag2",
			expectedTags: []string{"new-tag1", "new-tag2"},
			wantErr:      false,
		},
		{
			name:         "tags file with commas",
			tagsContent:  "tag1,tag2,tag3",
			expectedTags: []string{"tag1,tag2,tag3"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}

			tempDir := t.TempDir()
			tmpFilePath := filepath.Join(tempDir, "tags-file.txt")

			if tt.tagsContent != "" {
				_ = os.WriteFile(tmpFilePath, []byte(tt.tagsContent), 0o644)
			}

			t.Setenv("PLUGIN_TAGS_FILE", tmpFilePath)

			got := setupPluginTest(t)

			_ = got.InitSettings()
			err := got.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTags, got.Settings.Build.Tags)
		})
	}
}
