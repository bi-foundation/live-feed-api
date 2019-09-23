module github.com/FactomProject/live-feed-api

go 1.12

replace github.com/FactomProject/live-feed-api/EventRouter => ./EventRouter

require (
	github.com/FactomProject/live-feed-api/EventRouter v0.0.0-00010101000000-000000000000
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
)
