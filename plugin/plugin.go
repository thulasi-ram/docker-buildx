package plugin

import (
	"codeberg.org/woodpecker-plugins/go-plugin"
	plugin_cli "codeberg.org/woodpecker-plugins/go-plugin/cli"
	"github.com/urfave/cli/v3"
)

// Daemon defines Docker daemon parameters.
type Daemon struct {
	Registry          string   // Docker registry
	Mirror            string   // Docker registry mirror
	Insecure          bool     // Docker daemon enable insecure registries
	StorageDriver     string   // Docker daemon storage driver
	StoragePath       string   // Docker daemon storage path
	Disabled          bool     // Docker daemon is disabled (already running)
	Debug             bool     // Docker daemon started in debug mode
	Bip               string   // Docker daemon network bridge IP address
	DNS               []string // Docker daemon dns server
	DNSSearch         []string // Docker daemon dns search domain
	MTU               string   // Docker daemon mtu setting
	IPv6              bool     // Docker daemon IPv6 networking
	Experimental      bool     // Docker daemon enable experimental mode
	BuildkitConfig    string   // Docker buildkit config
	BuildkitDriverOpt []string // Docker buildkit driveropt args
	BuildkitDebug     bool     // Docker buildkit debug setting
	SshKey            string   // Docker daemon SSH key
	RemoteBuilders    []string // Remote builders for buildx to use
}

// Login defines Docker login parameters.
type Login struct {
	// Generic
	Registry string   // Docker registry address
	Username string   // Docker registry username
	Password string   // Docker registry password
	Email    string   // Docker registry email
	Config   string   // Docker Auth Config
	Mirrors  []string // Docker registry mirrors
	// ECR
	Aws_access_key_id     string `json:"aws_access_key_id"`     // AWS access key id
	Aws_secret_access_key string `json:"aws_secret_access_key"` // AWS secret access key
	Aws_region            string `json:"aws_region"`            // AWS region
}

// Build defines Docker build parameters.
type Build struct {
	Remote          string            // Git remote URL
	Ref             string            // Git commit ref
	Branch          string            // Git repository branch
	Dockerfile      string            // Docker build Dockerfile
	Context         string            // Docker build context
	TagsAuto        bool              // Docker build auto tag
	TagsDefaultName string            // Docker build auto tag name override
	TagsSuffix      string            // Docker build tags with suffix
	Tags            []string          // Docker build tags
	TagsFile        string            // Docker build tags read from an file
	LabelsAuto      bool              // Docker build auto labels
	Labels          []string          // Docker build labels
	Platforms       []string          // Docker build target platforms
	Args            map[string]string // Docker build args
	ArgsEnv         []string          // Docker build args from env
	Secrets         []string          // Docker build secret
	Target          string            // Docker build target
	Output          string            // Docker build output
	Pull            bool              // Docker build pull
	CacheFrom       []string          // Docker build cache-from
	CacheTo         string            // Docker build cache-to
	CacheImages     []string          // Docker build cache images
	Compress        bool              // Docker build compress
	Repo            []string          // Docker build repository
	NoCache         bool              // Docker build no-cache
	AddHost         []string          // Docker build add-host
	Quiet           bool              // Docker build quiet
	Epoch           int64             // Docker build epoch
	Provenance      string            // Docker build provenance
	Sbom            string            // Docker build sbom
}

type ProxyConf struct {
	Http  string // HTTP_PROXY environment variable
	Https string // HTTPS_PROXY environment variable
	No    string // NO_PROXY environment variable
}

// Settings for the Plugin.
type Settings struct {
	// ECR
	AwsRegion           string `json:"aws_region"`            // AWS region
	EcrScanOnPush       bool   `json:"ecr_scan_on_push"`      // ECR scan on push
	EcrRepositoryPolicy string `json:"ecr_repository_policy"` // ECR repository policy
	EcrLifecyclePolicy  string `json:"ecr_lifecycle_policy"`  // ECR lifecycle policy
	EcrCreateRepository bool   `json:"ecr_create_repository"` // ECR create repository
	AwsAccessKeyId      string `json:"aws_access_key_id"`     // AWS access key id
	AwsSecretAccessKey  string `json:"aws_secret_access_key"` // AWS secret access key

	// Generic
	Daemon          Daemon
	Logins          []Login
	LoginsRaw       string
	DefaultLogin    Login
	Build           Build
	ProxyConf       ProxyConf
	Dryrun          bool
	Cleanup         bool
	CustomCertStore string // e.g. for "/etc/docker/certs.d/<registry>/ca.crt"
}

// Plugin implements drone.Plugin to provide the plugin implementation.
type Plugin struct {
	*plugin.Plugin
	Settings *Settings
}

// New initializes a plugin from the given Settings, Pipeline, and Network.
func New(e plugin.ExecuteFunc, version string) *Plugin {
	p := &Plugin{
		Settings: &Settings{
			CustomCertStore: "/etc/docker/certs.d/",
		},
	}

	options := plugin.Options{
		Name:        "plugin-docker-buildx",
		Description: "build docker container with DinD and buildx",
		Version:     version,
		Flags:       p.Flags(),
		Execute:     p.Execute,
	}

	if e != nil {
		options.Execute = e
	}

	p.Plugin = plugin.New(options)

	return p
}

func (p *Plugin) Flags() []cli.Flag {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:        "dry-run",
			Sources:     cli.EnvVars("PLUGIN_DRY_RUN"),
			Usage:       "disables docker push",
			Destination: &p.Settings.Dryrun,
		},
		&cli.StringFlag{
			Name:        "remote.url",
			Sources:     cli.EnvVars("CI_REMOTE_URL", "DRONE_REMOTE_URL"),
			Usage:       "sets the git remote url",
			Destination: &p.Settings.Build.Remote,
		},
		&cli.StringFlag{
			Name:        "daemon.mirror",
			Sources:     cli.EnvVars("PLUGIN_MIRROR", "DOCKER_PLUGIN_MIRROR"),
			Usage:       "sets a registry mirror to pull images",
			Destination: &p.Settings.Daemon.Mirror,
		},
		&cli.StringFlag{
			Name:        "daemon.storage-driver",
			Sources:     cli.EnvVars("PLUGIN_STORAGE_DRIVER"),
			Usage:       "sets the docker daemon storage driver",
			Destination: &p.Settings.Daemon.StorageDriver,
		},
		&cli.StringFlag{
			Name:        "daemon.storage-path",
			Sources:     cli.EnvVars("PLUGIN_STORAGE_PATH"),
			Usage:       "sets the docker daemon storage path",
			Value:       "/var/lib/docker",
			Destination: &p.Settings.Daemon.StoragePath,
		},
		&cli.StringFlag{
			Name:        "daemon.bip",
			Sources:     cli.EnvVars("PLUGIN_BIP"),
			Usage:       "allows the docker daemon to bride ip address",
			Destination: &p.Settings.Daemon.Bip,
		},
		&cli.StringFlag{
			Name:        "daemon.mtu",
			Sources:     cli.EnvVars("PLUGIN_MTU"),
			Usage:       "sets docker daemon custom mtu setting",
			Destination: &p.Settings.Daemon.MTU,
		},
		&cli.StringSliceFlag{
			Name:        "daemon.dns",
			Sources:     cli.EnvVars("PLUGIN_CUSTOM_DNS"),
			Usage:       "sets custom docker daemon dns server",
			Destination: &p.Settings.Daemon.DNS,
		},
		&cli.StringSliceFlag{
			Name:        "daemon.dns-search",
			Sources:     cli.EnvVars("PLUGIN_CUSTOM_DNS_SEARCH"),
			Usage:       "sets custom docker daemon dns search domain",
			Destination: &p.Settings.Daemon.DNSSearch,
		},
		&cli.BoolFlag{
			Name:        "daemon.insecure",
			Sources:     cli.EnvVars("PLUGIN_INSECURE"),
			Usage:       "allows the docker daemon to use insecure registries",
			Destination: &p.Settings.Daemon.Insecure,
		},
		&cli.BoolFlag{
			Name:        "daemon.ipv6",
			Sources:     cli.EnvVars("PLUGIN_IPV6"),
			Usage:       "enables docker daemon IPv6 support",
			Destination: &p.Settings.Daemon.IPv6,
		},
		&cli.BoolFlag{
			Name:        "daemon.experimental",
			Sources:     cli.EnvVars("PLUGIN_EXPERIMENTAL"),
			Usage:       "enables docker daemon experimental mode",
			Destination: &p.Settings.Daemon.Experimental,
		},
		&cli.BoolFlag{
			Name:        "daemon.debug",
			Sources:     cli.EnvVars("PLUGIN_DEBUG", "DOCKER_LAUNCH_DEBUG"),
			Usage:       "enables verbose debug mode for the docker daemon",
			Destination: &p.Settings.Daemon.Debug,
		},
		&cli.BoolFlag{
			Name:        "daemon.off",
			Sources:     cli.EnvVars("PLUGIN_DAEMON_OFF"),
			Usage:       "disables the startup of the docker daemon",
			Destination: &p.Settings.Daemon.Disabled,
		},
		&cli.StringFlag{
			Name:        "daemon.ssh-key",
			Sources:     cli.EnvVars("PLUGIN_SSH_KEY", "SSH_KEY"),
			Usage:       "adds a ssh key to be used by the daemon",
			Destination: &p.Settings.Daemon.SshKey,
		},
		&cli.StringSliceFlag{
			Name:        "daemon.remote-builder",
			Sources:     cli.EnvVars("PLUGIN_REMOTE_BUILDERS", "REMOTE_BUILDERS"),
			Usage:       "adds remote builders to be used by buildx",
			Destination: &p.Settings.Daemon.RemoteBuilders,
		},
		&cli.BoolFlag{
			Name:        "daemon.buildkit-debug",
			Sources:     cli.EnvVars("PLUGIN_BUILDKIT_DEBUG"),
			Usage:       "enables buildkit debug",
			Destination: &p.Settings.Daemon.BuildkitDebug,
		},
		&cli.StringFlag{
			Name:        "dockerfile",
			Sources:     cli.EnvVars("PLUGIN_DOCKERFILE"),
			Usage:       "sets dockerfile to use for the image build",
			Value:       "Dockerfile",
			Destination: &p.Settings.Build.Dockerfile,
		},
		&cli.StringFlag{
			Name:        "context",
			Sources:     cli.EnvVars("PLUGIN_CONTEXT"),
			Usage:       "sets the path of the build context to use",
			Value:       ".",
			Destination: &p.Settings.Build.Context,
		},
		&cli.StringSliceFlag{
			Name: "tags",
			Sources: cli.NewValueSourceChain(
				cli.File(".tags"),
				cli.EnvVar("PLUGIN_TAG"),
				cli.EnvVar("PLUGIN_TAGS"),
			),
			Usage:       "sets repository tags to use for the image",
			Value:       []string{"latest"},
			Destination: &p.Settings.Build.Tags,
		},
		&cli.StringFlag{
			Name:        "tags.file",
			Sources:     cli.EnvVars("PLUGIN_TAGS_FILE", "PLUGIN_TAG_FILE"),
			Usage:       "overwrites tags flag with values find in set file",
			Destination: &p.Settings.Build.TagsFile,
		},
		&cli.BoolFlag{
			Name:        "tags.auto",
			Sources:     cli.EnvVars("PLUGIN_AUTO_TAG"),
			Usage:       "generates tag names automatically based on git branch and git tag",
			Destination: &p.Settings.Build.TagsAuto,
		},
		&cli.StringFlag{
			Name:        "tags.defaultName",
			Sources:     cli.EnvVars("PLUGIN_DEFAULT_TAG"),
			Usage:       "allows setting an alternative to `latest` for the auto tag",
			Destination: &p.Settings.Build.TagsDefaultName,
			Value:       "latest",
		},
		&cli.StringFlag{
			Name:        "tags.suffix",
			Sources:     cli.EnvVars("PLUGIN_DEFAULT_SUFFIX", "PLUGIN_AUTO_TAG_SUFFIX"),
			Usage:       "generates tag names with the given suffix",
			Destination: &p.Settings.Build.TagsSuffix,
		},
		&cli.StringSliceFlag{
			Name:        "labels",
			Sources:     cli.EnvVars("PLUGIN_LABEL", "PLUGIN_LABELS"),
			Usage:       "sets labels to use for the image",
			Destination: &p.Settings.Build.Labels,
		},
		&cli.BoolFlag{
			Name:        "labels.auto",
			Sources:     cli.EnvVars("PLUGIN_DEFAULT_LABELS", "PLUGIN_AUTO_LABEL"),
			Usage:       "generates labels automatically based on git repository info",
			Value:       true,
			Destination: &p.Settings.Build.LabelsAuto,
		},
		&plugin_cli.StringMapFlag{
			Name:        "args",
			Sources:     cli.EnvVars("PLUGIN_BUILD_ARGS"),
			Usage:       "sets custom build arguments for the build",
			Destination: &p.Settings.Build.Args,
		},
		&cli.StringSliceFlag{
			Name:        "args-from-env",
			Sources:     cli.EnvVars("PLUGIN_BUILD_ARGS_FROM_ENV"),
			Usage:       "forwards environment variables as custom arguments to the build",
			Destination: &p.Settings.Build.ArgsEnv,
		},
		&plugin_cli.StringSliceFlag{
			Name:        "secrets",
			Sources:     cli.EnvVars("PLUGIN_SECRETS"),
			Usage:       "sets custom secret arguments for the build",
			Destination: &p.Settings.Build.Secrets,
			Config: plugin_cli.StringSliceConfig{
				Delimiter:    ",",
				EscapeString: "\\",
			},
		},
		&cli.BoolFlag{
			Name:        "quiet",
			Sources:     cli.EnvVars("PLUGIN_QUIET"),
			Usage:       "enables suppression of the build output",
			Destination: &p.Settings.Build.Quiet,
		},
		&cli.StringFlag{
			Name:        "target",
			Sources:     cli.EnvVars("PLUGIN_TARGET"),
			Usage:       "sets the build target to use",
			Destination: &p.Settings.Build.Target,
		},
		&plugin_cli.StringSliceFlag{
			Name:        "cache-from",
			Sources:     cli.EnvVars("PLUGIN_CACHE_FROM"),
			Usage:       "sets images to consider as cache sources",
			Destination: &p.Settings.Build.CacheFrom,
			Config: plugin_cli.StringSliceConfig{
				Delimiter:    ",",
				EscapeString: "\\",
			},
		},
		&cli.StringFlag{
			Name:        "cache-to",
			Sources:     cli.EnvVars("PLUGIN_CACHE_TO"),
			Usage:       "cache destination for the build cache",
			Destination: &p.Settings.Build.CacheTo,
		},
		&cli.StringSliceFlag{
			Name:        "cache-images",
			Sources:     cli.EnvVars("PLUGIN_CACHE_IMAGES"),
			Usage:       "list of images to use for build cache. applies both to and from flags for each image",
			Destination: &p.Settings.Build.CacheImages,
		},
		&cli.BoolFlag{
			Name:        "pull-image",
			Sources:     cli.EnvVars("PLUGIN_PULL_IMAGE"),
			Usage:       "enforces to pull base image at build time",
			Value:       true,
			Destination: &p.Settings.Build.Pull,
		},
		&cli.BoolFlag{
			Name:        "compress",
			Sources:     cli.EnvVars("PLUGIN_COMPRESS"),
			Usage:       "enables compression og the build context using gzip",
			Destination: &p.Settings.Build.Compress,
		},
		&cli.StringSliceFlag{
			Name:        "repo",
			Sources:     cli.EnvVars("PLUGIN_REPO"),
			Usage:       "sets repository name for the image",
			Destination: &p.Settings.Build.Repo,
		},
		&cli.StringFlag{
			Name:        "docker.registry",
			Sources:     cli.EnvVars("PLUGIN_REGISTRY", "DOCKER_REGISTRY"),
			Usage:       "sets docker registry to authenticate with",
			Value:       "https://index.docker.io/v1/",
			Destination: &p.Settings.DefaultLogin.Registry,
		},
		&cli.StringFlag{
			Name:        "docker.username",
			Sources:     cli.EnvVars("PLUGIN_USERNAME", "DOCKER_USERNAME"),
			Usage:       "sets username to authenticates with",
			Destination: &p.Settings.DefaultLogin.Username,
		},
		&cli.StringFlag{
			Name:        "docker.password",
			Sources:     cli.EnvVars("PLUGIN_PASSWORD", "DOCKER_PASSWORD"),
			Usage:       "sets password to authenticates with",
			Destination: &p.Settings.DefaultLogin.Password,
		},
		&cli.StringFlag{
			Name:        "docker.email",
			Sources:     cli.EnvVars("PLUGIN_EMAIL", "DOCKER_EMAIL"),
			Usage:       "sets email address to authenticates with",
			Destination: &p.Settings.DefaultLogin.Email,
		},
		&cli.StringFlag{
			Name:        "docker.config",
			Sources:     cli.EnvVars("PLUGIN_CONFIG", "DOCKER_PLUGIN_CONFIG"),
			Usage:       "sets content of the docker daemon json config",
			Destination: &p.Settings.DefaultLogin.Config,
		},
		&cli.StringFlag{
			Name:        "logins",
			Sources:     cli.EnvVars("PLUGIN_LOGINS"),
			Usage:       "list of login",
			Destination: &p.Settings.LoginsRaw,
			Value:       "[]",
		},
		&cli.BoolFlag{
			Name:        "docker.purge",
			Sources:     cli.EnvVars("PLUGIN_PURGE"),
			Usage:       "enables cleanup of the docker environment at the end of a build",
			Value:       true,
			Destination: &p.Settings.Cleanup,
		},
		&cli.BoolFlag{
			Name:        "no-cache",
			Sources:     cli.EnvVars("PLUGIN_NO_CACHE"),
			Usage:       "disables the usage of cached intermediate containers",
			Destination: &p.Settings.Build.NoCache,
		},
		&cli.StringSliceFlag{
			Name:        "add-host",
			Sources:     cli.EnvVars("PLUGIN_ADD_HOST"),
			Usage:       "sets additional host:ip mapping",
			Destination: &p.Settings.Build.AddHost,
		},
		&cli.StringSliceFlag{
			Name:        "platforms",
			Sources:     cli.EnvVars("PLUGIN_PLATFORMS"),
			Usage:       "sets target platform for build",
			Destination: &p.Settings.Build.Platforms,
		},
		&cli.StringFlag{
			Name:        "output",
			Sources:     cli.EnvVars("PLUGIN_OUTPUT"),
			Usage:       "sets build output type and destination configuration",
			Destination: &p.Settings.Build.Output,
		},
		&cli.StringFlag{
			Name:        "ecr.aws_access_key_id",
			Sources:     cli.EnvVars("PLUGIN_AWS_ACCESS_KEY_ID"),
			Usage:       "Access Key ID for AWS",
			Destination: &p.Settings.AwsAccessKeyId,
		},
		&cli.StringFlag{
			Name:        "ecr.aws_secret_access_key_id",
			Sources:     cli.EnvVars("PLUGIN_AWS_SECRET_ACCESS_KEY"),
			Usage:       "Secret Access Key for AWS",
			Destination: &p.Settings.AwsSecretAccessKey,
		},
		&cli.StringFlag{
			Name:        "ecr.aws_region",
			Sources:     cli.EnvVars("PLUGIN_AWS_REGION"),
			Usage:       "AWS region to use",
			Destination: &p.Settings.AwsRegion,
		},
		&cli.BoolFlag{
			Name:        "ecr.create_repository",
			Sources:     cli.EnvVars("PLUGIN_ECR_CREATE_REPOSITORY"),
			Usage:       "creates the ECR repository if it does not exist",
			Destination: &p.Settings.EcrCreateRepository,
		},
		&cli.StringFlag{
			Name:        "ecr.lifecycle_policy",
			Sources:     cli.EnvVars("PLUGIN_ECR_LIFECYCLE_POLICY"),
			Usage:       "AWS ECR lifecycle policy",
			Destination: &p.Settings.EcrLifecyclePolicy,
		},
		&cli.StringFlag{
			Name:        "ecr.repository_policy",
			Sources:     cli.EnvVars("PLUGIN_ECR_REPOSITORY_POLICY"),
			Usage:       "AWS ECR repository policy",
			Destination: &p.Settings.EcrRepositoryPolicy,
		},
		&cli.BoolFlag{
			Name:        "ecr.scan_on_push",
			Sources:     cli.EnvVars("PLUGIN_ECR_SCAN_ON_PUSH"),
			Usage:       "AWS: whether to enable image scanning on push",
			Destination: &p.Settings.EcrScanOnPush,
		},
		&cli.StringFlag{
			Name:        "provenance",
			Sources:     cli.EnvVars("PLUGIN_PROVENANCE"),
			Usage:       "defines provenance setting",
			Destination: &p.Settings.Build.Provenance,
		},
		&cli.StringFlag{
			Name:        "http_proxy",
			Sources:     cli.EnvVars("PLUGIN_HTTP_PROXY", "HTTP_PROXY"),
			Usage:       "sets the HTTP_PROXY env",
			Destination: &p.Settings.ProxyConf.Http,
		},
		&cli.StringFlag{
			Name:        "https_proxy",
			Sources:     cli.EnvVars("PLUGIN_HTTPS_PROXY", "HTTPS_PROXY"),
			Usage:       "sets the HTTPS_PROXY env",
			Destination: &p.Settings.ProxyConf.Https,
		},
		&cli.StringFlag{
			Name:        "no_proxy",
			Sources:     cli.EnvVars("PLUGIN_NO_PROXY", "NO_PROXY"),
			Usage:       "sets the NO_PROXY env",
			Destination: &p.Settings.ProxyConf.No,
		},
		&cli.StringFlag{
			Name:        "sbom",
			Sources:     cli.EnvVars("PLUGIN_SBOM"),
			Usage:       "defines sbom setting",
			Destination: &p.Settings.Build.Sbom,
		},
	}

	if Insecure {
		flags = append(flags,
			&cli.StringFlag{
				Name:        "daemon.buildkit-config",
				Sources:     cli.EnvVars("PLUGIN_BUILDKIT_CONFIG"),
				Usage:       "sets content of the docker buildkit json config",
				Destination: &p.Settings.Daemon.BuildkitConfig,
			},
			&cli.StringSliceFlag{
				Name:        "daemon.buildkit-driveropt",
				Sources:     cli.EnvVars("PLUGIN_BUILDKIT_DRIVEROPT"),
				Usage:       "adds optional driver-ops args like 'env.http_proxy'",
				Destination: &p.Settings.Daemon.BuildkitDriverOpt,
			})
	}

	return flags
}
