package api

// These are the extensions that are present when the daemon starts.
var extensions = []string{
	"site_management",
}

// Extensions returns the list of site manager extensions.
func Extensions() []string {
	return extensions
}
