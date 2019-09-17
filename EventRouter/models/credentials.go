package models

type Credentials struct {

	// Access token for setting the bearer token when authenticating on at the callback url. This is required when the callback type is set on BEARER_TOKEN.
	AccessToken string `json:"accessToken"`

	// Username for authenticating with basic authentication. This is required when the callback type is set on BASIC_AUTH.
	BasicAuthUsername string `json:"basicAuthUsername"`

	// Password for authenticating with basic authentication. This is required when the callback type is set on BASIC_AUTH.
	BasicAuthPassword string `json:"basicAuthPassword"`
}
