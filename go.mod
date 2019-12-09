module github.com/FactomProject/live-feed-api

go 1.12

replace github.com/FactomProject/live-feed-api/EventRouter => ./EventRouter

require (
	github.com/FactomProject/live-feed-api/EventRouter v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/swaggo/swag v1.6.3 // indirect
	golang.org/x/tools v0.0.0-20191125144606-a911d9008d1f // indirect
)
