package teamtailor

import "net/http"

// Region identifies a Teamtailor data region.
// Teamtailor hosts API endpoints in three regions; pick the one matching your account.
type Region string

const (
	// RegionEU is the European data region (default).
	RegionEU Region = "eu"
	// RegionNA is the North American data region.
	RegionNA Region = "na"
	// RegionAU is the Australian data region.
	RegionAU Region = "au"
)

// regionBaseURL maps a Region to its API origin.
var regionBaseURL = map[Region]string{
	RegionEU: "https://api.teamtailor.com",
	RegionNA: "https://api.na.teamtailor.com",
	RegionAU: "https://api.au.teamtailor.com",
}

// DefaultAPIVersion is the X-Api-Version sent when none is provided.
// Teamtailor uses dated API versions; bump this when adopting a newer one.
const DefaultAPIVersion = "20240904"

// Option configures the Client returned by NewClient.
type Option func(*options)

type options struct {
	httpClient           *http.Client
	region               Region
	baseURL              string
	apiVersion           string
	discardUnknownFields bool
}

func defaultOptions() *options {
	return &options{
		httpClient:           http.DefaultClient,
		region:               RegionEU,
		apiVersion:           DefaultAPIVersion,
		discardUnknownFields: false,
	}
}

// WithHTTPClient supplies a custom *http.Client (timeouts, transport, proxy, etc.).
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) { o.httpClient = c }
}

// WithRegion selects a Teamtailor data region. Defaults to RegionEU.
// Ignored if WithBaseURL is also set.
func WithRegion(r Region) Option {
	return func(o *options) { o.region = r }
}

// WithBaseURL overrides the API origin entirely (e.g. for a mock server).
// Takes precedence over WithRegion.
func WithBaseURL(u string) Option {
	return func(o *options) { o.baseURL = u }
}

// WithAPIVersion overrides the X-Api-Version header (default: DefaultAPIVersion).
func WithAPIVersion(v string) Option {
	return func(o *options) { o.apiVersion = v }
}

// WithDiscardUnknownFields silently ignores unknown JSON fields instead of erroring.
// Useful when Teamtailor adds new fields ahead of an SDK release.
func WithDiscardUnknownFields(discard bool) Option {
	return func(o *options) { o.discardUnknownFields = discard }
}

func (o *options) resolveBaseURL() string {
	if o.baseURL != "" {
		return o.baseURL
	}
	if url, ok := regionBaseURL[o.region]; ok {
		return url
	}
	return regionBaseURL[RegionEU]
}
