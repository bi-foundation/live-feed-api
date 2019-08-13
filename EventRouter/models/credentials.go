package models

// credentials of api to deliver events
// swagger:model
type Credentials struct {
	AccessToken string `json:"accessToken"`

	BasicAuthUsername string `json:"basicAuthUsername"`

	BasicAuthPassword string `json:"basicAuthPassword"`
}
