package upstreams

// The Dummy upstream needs no parameters and always returns a latest version of 1.0.0. Can be used for testing.
type Dummy struct {
	UpstreamBase
}

func (upstream Dummy) LatestVersion() (string, error) {
	return "1.0.0", nil
}
