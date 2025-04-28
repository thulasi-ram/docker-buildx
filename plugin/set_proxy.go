package plugin

import (
	"fmt"
	"os"
	"strings"
)

func (p *Plugin) applyProxyConf() error {
	if p.Settings.ProxyConf.Http == "" &&
		p.Settings.ProxyConf.Https == "" &&
		p.Settings.ProxyConf.No == "" {
		return nil
	}

	// we set the environment for all commands we do exec
	if p.Settings.ProxyConf.Http != "" {
		if err := os.Setenv("HTTP_PROXY", p.Settings.ProxyConf.Http); err != nil {
			return fmt.Errorf("could not set HTTP_PROXY as environment variable: %w", err)
		}
	}
	if p.Settings.ProxyConf.Https != "" {
		if err := os.Setenv("HTTPS_PROXY", p.Settings.ProxyConf.Https); err != nil {
			return fmt.Errorf("could not set HTTPS_PROXY as environment variable: %w", err)
		}
	}
	if p.Settings.ProxyConf.No != "" {
		if err := os.Setenv("NO_PROXY", p.Settings.ProxyConf.No); err != nil {
			return fmt.Errorf("could not set NO_PROXY as environment variable: %w", err)
		}
	}

	// add driver-opt http config to tell buildkit + buildx to resolve external checksums through a proxy.
	if p.Settings.ProxyConf.Http != "" && !prefixExistInList(p.Settings.Daemon.BuildkitDriverOpt, "env.http_proxy=") {
		p.Settings.Daemon.BuildkitDriverOpt = append(p.Settings.Daemon.BuildkitDriverOpt, fmt.Sprintf("env.http_proxy=%s", p.Settings.ProxyConf.Http))
	}
	if p.Settings.ProxyConf.Https != "" && !prefixExistInList(p.Settings.Daemon.BuildkitDriverOpt, "env.https_proxy=") {
		p.Settings.Daemon.BuildkitDriverOpt = append(p.Settings.Daemon.BuildkitDriverOpt, fmt.Sprintf("env.https_proxy=%s", p.Settings.ProxyConf.Https))
	}
	if p.Settings.ProxyConf.No != "" && !prefixExistInList(p.Settings.Daemon.BuildkitDriverOpt, "env.no_proxy=") {
		p.Settings.Daemon.BuildkitDriverOpt = append(p.Settings.Daemon.BuildkitDriverOpt, fmt.Sprintf("\"env.no_proxy=%s\"", p.Settings.ProxyConf.No))
	}

	// passthrough proxy config to the build process and Dockerfile CMDs itself.
	if p.Settings.ProxyConf.Http != "" {
		p.Settings.Build.Args["HTTP_PROXY"] = p.Settings.ProxyConf.Http
	}
	if p.Settings.ProxyConf.Https != "" {
		p.Settings.Build.Args["HTTPS_PROXY"] = p.Settings.ProxyConf.Https
	}
	if p.Settings.ProxyConf.No != "" {
		p.Settings.Build.Args["NO_PROXY"] = p.Settings.ProxyConf.No
	}

	return nil
}

func prefixExistInList(list []string, prefix string) bool {
	for i := range list {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(list[i])), prefix) {
			return true
		}
	}
	return false
}
