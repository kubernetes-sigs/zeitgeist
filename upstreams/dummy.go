package upstreams

type DummyUpstream struct {
	AccessToken string
	URL         string
	Constraints string
}

func (upstream DummyUpstream) LatestVersion() string {
	return "1.0.0"
}
