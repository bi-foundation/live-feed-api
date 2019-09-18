module github.com/FactomProject/live-feed-api/EventRouter

go 1.12

require (
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc
	github.com/bi-foundation/protobuf-graphql-extension v1.0.19
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gogo/protobuf v1.3.0
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/graphql-go/graphql v0.7.8
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect

	github.com/proullon/ramsql v0.0.0-20181213202341-817cee58a244
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/swaggo/swag v1.6.2
	github.com/ziutek/mymysql v1.5.4 // indirect
	google.golang.org/appengine v1.6.1 // indirect

)

replace github.com/FactomProject/live-feed-api/EventRouter/api => ./api
