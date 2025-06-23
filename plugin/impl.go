package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/sirupsen/logrus"

	"codeberg.org/woodpecker-plugins/plugin-docker-buildx/utils"
)

var ErrTypeAssertionFailed = errors.New("type assertion failed")

func (l Login) anonymous() bool {
	return l.Username == "" || l.Password == ""
}

// Init initialise plugin settings
func (p *Plugin) InitSettings() error {
	if p.Settings.LoginsRaw != "" {
		if err := json.Unmarshal([]byte(p.Settings.LoginsRaw), &p.Settings.Logins); err != nil {
			return fmt.Errorf("could not unmarshal logins: %v", err)
		}
	}

	if err := p.applyProxyConf(); err != nil {
		return err
	}

	p.Settings.Build.Branch = p.Metadata.Commit.Branch
	p.Settings.Build.Ref = p.Metadata.Commit.Ref

	// check if any Login struct contains AWS credentials
	for _, login := range p.Settings.Logins {
		if strings.Contains(login.Registry, "amazonaws.com") {
			p.EcrInit()
		}
	}

	if p.Settings.DefaultLogin.Registry != "" && strings.Contains(p.Settings.DefaultLogin.Registry, "amazonaws.com") {
		p.EcrInit()
	}

	if p.Settings.AwsAccessKeyId != "" && p.Settings.AwsSecretAccessKey != "" {
		p.EcrInit()
	}

	if p.Settings.DefaultLogin.Registry != "" && p.Settings.Daemon.Mirror != "" {
		p.Settings.DefaultLogin.Mirrors = []string{p.Settings.Daemon.Mirror}
	}

	if len(p.Settings.Logins) == 0 {
		p.Settings.Logins = []Login{p.Settings.DefaultLogin}
	} else if !p.Settings.DefaultLogin.anonymous() || len(p.Settings.DefaultLogin.Mirrors) > 0 {
		p.Settings.Logins = prepend(p.Settings.Logins, p.Settings.DefaultLogin)
	}

	p.Settings.Daemon.Registry = p.Settings.Logins[0].Registry

	return nil
}

// Validate handles the settings validation of the plugin.
func (p *Plugin) Validate() error {
	if !isSingleTag(p.Settings.Build.TagsDefaultName) {
		return fmt.Errorf("'%s' is not a valid, single tag", p.Settings.Build.TagsDefaultName)
	}

	// beside the default login all other logins need to set either a username and password or custom mirrors
	for _, l := range p.Settings.Logins[1:] {
		if l.anonymous() && len(l.Mirrors) == 0 {
			return fmt.Errorf("beside the default login all other logins need to set either a username and password or custom mirrors")
		}
	}

	// overload tags flag with tags.file if set
	if p.Settings.Build.TagsFile != "" {
		tagsFile, err := os.ReadFile(p.Settings.Build.TagsFile)
		if err != nil {
			return fmt.Errorf("could not read tags file: %w", err)
		}

		// split file content into slice of tags
		tagsFileList := strings.Split(strings.TrimSpace(string(tagsFile)), "\n")
		// trim space of each tag
		tagsFileList = utils.Map(tagsFileList, func(s string) string { return strings.TrimSpace(s) })

		// finally overwrite
		p.Settings.Build.Tags = tagsFileList
	}

	if p.Settings.Build.TagsAuto {
		// we only generate tags on default branch or an tag event
		if UseDefaultTag(
			p.Settings.Build.Ref,
			p.Settings.Build.Branch,
		) {
			var tag []string
			var err error

			tag, err = DefaultTagSuffix(
				p.Settings.Build.Ref,
				p.Settings.Build.TagsDefaultName,
				p.Settings.Build.TagsSuffix,
			)
			if err != nil {
				logrus.Printf("cannot build docker image for %s, invalid semantic version", p.Settings.Build.Ref)
				return err
			}

			// include user supplied tags
			tag = append(tag, p.sanitizedUserTags()...)

			p.Settings.Build.Tags = tag
		} else {
			logrus.Printf("skipping automated docker build for %s", p.Settings.Build.Ref)
			return nil
		}
	} else {
		p.Settings.Build.Tags = p.sanitizedUserTags()
	}

	if p.Settings.Build.LabelsAuto {
		p.Settings.Build.Labels = p.Labels()
	}

	if err := p.generateBuildkitConfig(); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) sanitizedUserTags() []string {
	// ignore empty tags
	var tags []string
	for _, t := range p.Settings.Build.Tags {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

type BuildkitConfigTOML struct {
	Debug    bool                     `toml:"debug,omitempty"` // needs to be public for toml lib to use
	Registry map[string]*RegistryInfo `toml:"registry,omitempty"`
}

type RegistryInfo struct {
	Mirrors []string `toml:"mirrors,omitempty"`
	CA      []string `toml:"ca,omitempty"`
}

func (p *Plugin) generateBuildkitConfig() error {
	// no buildkit config, automatically generate buildkit configuration
	if p.Settings.Daemon.BuildkitConfig == "" {

		cfg := BuildkitConfigTOML{}
		cfg.Registry = make(map[string]*RegistryInfo)

		if p.Settings.Daemon.BuildkitDebug {
			cfg.Debug = p.Settings.Daemon.BuildkitDebug
			logrus.Println("buildkit debug enabled")
		}

		// use a custom CA certificate for each registry
		if p.Settings.Daemon.Registry != "" {
			for _, login := range p.Settings.Logins {
				if registry := login.Registry; registry != "" {
					u, err := url.Parse(registry)
					if err != nil {
						return fmt.Errorf("could not parse registry address: %s: %v", registry, err)
					}
					if u.Host != "" {
						registry = u.Host
					}

					// docker hub fix
					if registry == "index.docker.io" {
						registry = "docker.io"
					}

					caPath := fmt.Sprintf("%s/%s/ca.crt", p.Settings.CustomCertStore, registry)
					ca, err := os.Open(caPath)
					if err != nil && !os.IsNotExist(err) {
						logrus.Warnf("error reading %s: %v", caPath, err)
					} else if err == nil {
						if err := ca.Close(); err != nil {
							return fmt.Errorf("error closing ca file: %v", err)
						}
						logrus.Infof("found ca file for '%s' registry", registry)
						// add registry and ca path to buildkit.toml
						if cfg.Registry[registry] == nil {
							cfg.Registry[registry] = new(RegistryInfo)
						}
						cfg.Registry[registry].CA = []string{caPath}
					}

					if len(login.Mirrors) != 0 {
						if cfg.Registry[registry] == nil {
							cfg.Registry[registry] = new(RegistryInfo)
						}
						cfg.Registry[registry].Mirrors = login.Mirrors
					}
				}
			}
		}

		if cfg.Debug || len(cfg.Registry) > 0 {
			tomlData, err := toml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("error marshaling buildkit.toml: %s", err)
			} else {
				p.Settings.Daemon.BuildkitConfig = string(tomlData)
			}
		}
	}

	return nil
}

func (p *Plugin) writeBuildkitConfig() error {
	// save buildkit config as described
	if p.Settings.Daemon.BuildkitConfig != "" {
		err := os.WriteFile(buildkitConfig, []byte(p.Settings.Daemon.BuildkitConfig), 0o600)
		if err != nil {
			return fmt.Errorf("error writing buildkit.toml: %s", err)
		}
	}

	return nil
}

// Lookup Host keys and add them to known_hosts if new
func addToKnownHosts(host string) error {
	if host == "local" {
		return nil
	}
	// make sure we only use a host
	host_slice := strings.Split(host, "@")
	host = host_slice[len(host_slice)-1]

	if err := os.MkdirAll("/root/.ssh", 0o700); err != nil {
		return fmt.Errorf("error creating /root/.ssh dir: %s", err)
	}

	khFile, err := os.OpenFile("/root/.ssh/known_hosts", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("error opening known_hosts file: %s", err)
	}
	defer khFile.Close()

	cmd := exec.Command("/usr/bin/ssh-keyscan", host)
	cmd.Stdout = khFile
	cmd.Stderr = os.Stderr
	trace(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error writing known_hosts file: %s", err)
	}

	return nil
}

// Execute provides the implementation of the plugin.
func (p *Plugin) Execute(ctx context.Context) error {
	if err := p.InitSettings(); err != nil {
		return err
	}

	if err := p.Validate(); err != nil {
		return err
	}

	// start the Docker daemon server
	if !p.Settings.Daemon.Disabled {
		// If no custom DNS value set start internal DNS server
		if len(p.Settings.Daemon.DNS) == 0 {
			ip, err := getContainerIP()
			if err != nil {
				logrus.Warnf("error detecting IP address: %v", err)
			} else if ip != "" {
				p.startCoredns()
				p.Settings.Daemon.DNS = []string{ip}
			}
		}

		p.startDaemon()
	}

	// poll the docker daemon until it is started. This ensures the daemon is
	// ready to accept connections before we proceed.
	for i := 0; i < 15; i++ {
		cmd := commandInfo()
		err := cmd.Run()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}

	// Create Auth Config File
	if p.Settings.Logins[0].Config != "" {
		if err := os.MkdirAll(dockerHome, 0o600); err != nil {
			return fmt.Errorf("error creating %s dir: %s", dockerHome, err)
		}

		path := filepath.Join(dockerHome, "config.json")
		err := os.WriteFile(path, []byte(p.Settings.Logins[0].Config), 0o600)
		if err != nil {
			return fmt.Errorf("error writing config.json: %s", err)
		}
	}

	// login to the Docker registry
	if err := p.Login(); err != nil {
		return err
	}

	if err := p.writeBuildkitConfig(); err != nil {
		return err
	}

	switch {
	case p.Settings.Logins[0].Password != "":
		fmt.Println("Detected registry credentials")
	case p.Settings.Logins[0].Config != "":
		fmt.Println("Detected registry credentials file")
	default:
		fmt.Println("Registry credentials or Docker config not provided. Guest mode enabled.")
	}

	// add proxy build args
	addProxyBuildArgs(&p.Settings.Build)

	// add optional SSH key
	if p.Settings.Daemon.SshKey != "" {
		if err := os.MkdirAll("/root/.ssh", 0o700); err != nil {
			return fmt.Errorf("error creating /root/.ssh dir: %s", err)
		}
		// make sure the key ends with a newline
		if !strings.HasSuffix(p.Settings.Daemon.SshKey, "\n") {
			p.Settings.Daemon.SshKey = strings.Join(
				append(strings.Split(p.Settings.Daemon.SshKey, "\n"), "\n"),
				"\n",
			)
		}
		if err := os.WriteFile("/root/.ssh/id_rsa", []byte(p.Settings.Daemon.SshKey), 0o600); err != nil {
			return fmt.Errorf("error writing SSH key: %s", err)
		}
	}

	var cmds []*exec.Cmd
	cmds = append(cmds, commandVersion())           // docker version
	cmds = append(cmds, commandInfo())              // docker info
	if len(p.Settings.Daemon.RemoteBuilders) == 0 { // no remote builder, create a local one
		cmds = append(cmds, commandBuilder(p.Settings.Daemon, "local", false))
	} else {
		append_builder := false // don't append for the first builder
		for _, host := range p.Settings.Daemon.RemoteBuilders {
			if err := addToKnownHosts(host); err != nil {
				return err
			}
			cmds = append(cmds, commandBuilder(p.Settings.Daemon, host, append_builder))
			append_builder = true // to add subsequent builders we need to append
		}
	}
	cmds = append(cmds, commandBuildx())
	cmds = append(cmds, commandBuild(p.Settings.Build, p.Settings.Dryrun)) // docker build
	if !p.Settings.Dryrun && p.Settings.Build.Output != "" &&
		!strings.Contains(p.Settings.Build.Output, ",push=") {
		cmds = append(cmds, commandsPush(p.Settings.Build)...) // docker push
	}

	// execute all commands in batch mode.
	for _, cmd := range cmds {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func prepend[Type any](slice []Type, elems ...Type) []Type {
	return append(elems, slice...)
}
