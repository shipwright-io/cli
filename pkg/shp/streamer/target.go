package streamer

// Target represents the target POD to receive streamed data.
type Target struct {
	Namespace string // kubernetes namespace
	Pod       string // pod name
	Container string // container name
	BaseDir   string // base directory to store streamed data
}
