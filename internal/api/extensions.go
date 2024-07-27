package api

// These are the extensions that are present when the daemon starts.
var extensions = []string{
	"remote_cluster_management",
}

// Extensions returns the list of Cluster Manager extensions.
func Extensions() []string {
	return extensions
}
