module github.com/FactomProject/LiveAPI

go 1.12

require (
	github.com/FactomProject/live-api/EventRouter v0.0.0-00010101000000-000000000000
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.2
)

replace github.com/FactomProject/live-api/EventRouter => ./EventRouter
