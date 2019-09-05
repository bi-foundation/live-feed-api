module github.com/FactomProject/live-feed-api/EventRouter

go 1.12

require (
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gogo/protobuf v1.3.0
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/graphql-go/graphql v0.7.8
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/opsee/protobuf v0.0.0-20170203071455-928523252569
	github.com/proullon/ramsql v0.0.0-20181213202341-817cee58a244
	github.com/stretchr/testify v1.3.0
	github.com/ziutek/mymysql v1.5.4 // indirect
	google.golang.org/appengine v1.6.1 // indirect

)

replace github.com/FactomProject/live-feed-api/EventRouter/api => ./api
