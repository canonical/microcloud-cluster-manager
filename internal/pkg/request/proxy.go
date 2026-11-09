package request

import (
	"maps"
	"net/http"
	"strings"
)

// hopByHopHeaders are the headers that only apply to a single connection and must not be forwarded by a proxy.
var hopByHopHeaders = map[string]struct{}{
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"Proxy-Connection":    {},
	"Te":                  {},
	"Trailer":             {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
}

// CopyResponseHeaders copies the end-to-end headers of an upstream response to dst.
func CopyResponseHeaders(dst http.Header, src http.Header) {
	hopByHop := maps.Clone(hopByHopHeaders)

	// Headers named by the Connection header are hop-by-hop as well.
	for _, value := range src.Values("Connection") {
		for name := range strings.SplitSeq(value, ",") {
			name = http.CanonicalHeaderKey(strings.TrimSpace(name))
			if name != "" {
				hopByHop[name] = struct{}{}
			}
		}
	}

	for key, values := range src {
		_, skip := hopByHop[http.CanonicalHeaderKey(key)]
		if !skip {
			dst[key] = append([]string(nil), values...)
		}
	}
}
