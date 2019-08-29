package upstreams

// Dummy upstream needs no parameters and always returns a latest version of 1.0.0. Can be used for testing.
type Dummy struct {
	UpstreamBase
}

// LatestVersion always returns 1.0.0
func (upstream Dummy) LatestVersion() (string, error) {
	return "1.0.0", nil
}
