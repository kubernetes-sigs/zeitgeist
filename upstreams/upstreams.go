package upstreams

type Upstream interface {
	latestVersion() string
}

type UpstreamFlavour string

const (
	GitHub UpstreamFlavour = "github"
)
