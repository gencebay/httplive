package types

// IPResponse ...
type IPResponse struct {
	Origin string `json:"origin"`
}

// UserAgentResponse ...
type UserAgentResponse struct {
	UserAgent string `json:"user-agent"`
}

// HeadersResponse ...
type HeadersResponse struct {
	Headers map[string]string `json:"headers"`
}

// CookiesResponse ...
type CookiesResponse struct {
	Cookies map[string]string `json:"cookies"`
}

// GetResponse ...
type GetResponse struct {
	Args map[string][]string `json:"args"`
	HeadersResponse
	IPResponse
	URL string `json:"url"`
}

// GzipResponse ...
type GzipResponse struct {
	HeadersResponse
	IPResponse
	Gzipped bool `json:"gzipped"`
}

// DeflateResponse ...
type DeflateResponse struct {
	HeadersResponse
	IPResponse
	Deflated bool `json:"deflated"`
}

// BasicAuthResponse ...
type BasicAuthResponse struct {
	Authenticated bool   `json:"authenticated"`
	User          string `json:"string"`
}
