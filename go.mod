module github.com/FactomProject/live-feed-api

go 1.12

replace github.com/FactomProject/live-feed-api/EventRouter => ./EventRouter

require (
	github.com/FactomProject/live-feed-api/EventRouter v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
)
