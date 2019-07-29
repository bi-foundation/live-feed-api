package models

type Subscription struct {
	Id       string `json:"id"`
	Callback string `json:"callback"`
}
