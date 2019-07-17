module github.com/FactomProject/LiveAPI

go 1.12

require (
	github.com/FactomProject/live-api/EventRouter v0.0.0-00010101000000-000000000000
	github.com/FactomProject/live-api/common v0.0.0-00010101000000-000000000000
)

replace (
	github.com/FactomProject/live-api/EventRouter => ./EventRouter
	github.com/FactomProject/live-api/common => ./common
)
