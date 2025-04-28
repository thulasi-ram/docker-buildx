package plugin

import (
	"fmt"
	"strings"
	"time"

	"github.com/6543/go-version"
)

// Labels returns list of labels to use for image
func (p *Plugin) Labels() []string {
	l := p.Settings.Build.Labels
	// As described in https://github.com/opencontainers/image-spec/blob/main/annotations.md
	l = append(l, fmt.Sprintf("org.opencontainers.image.created=%s", time.Now().UTC().Format(time.RFC3339)))
	if p.Settings.Build.Remote != "" {
		l = append(l, fmt.Sprintf("org.opencontainers.image.source=%s", p.Settings.Build.Remote))
	}
	if p.Metadata.Repository.Link != "" {
		l = append(l, fmt.Sprintf("org.opencontainers.image.url=%s", p.Metadata.Repository.Link))
	}
	if p.Metadata.Commit.Sha != "" {
		l = append(l, fmt.Sprintf("org.opencontainers.image.revision=%s", p.Metadata.Commit.Sha))
	}
	if p.Settings.Build.Ref != "" && strings.HasPrefix(p.Settings.Build.Ref, tagRefPrefix) {
		v, err := version.NewSemver(stripTagPrefix(p.Settings.Build.Ref))
		if err == nil && v != nil {
			l = append(l, fmt.Sprintf("org.opencontainers.image.version=%s", v.String()))
		}
	}
	return l
}
