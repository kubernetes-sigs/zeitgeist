package upstreams

type Dummy struct {
	UpstreamBase
}

func (upstream Dummy) LatestVersion() (string, error) {
	return "1.0.0", nil
}
