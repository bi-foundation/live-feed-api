
# generate swagger from source

Generating a new swagger json is done with go-swagger. Go-swagger needs to be installed in $GOPATH/bin. When using go generate, the program is called to generate the spec from the comments in the code. Further information: https://goswagger.io/use/spec.html.  
```
go get -u github.com/go-swagger/go-swagger/cmd/swagger
go generate
```


# generate MYSQL database
config
```sql
# drivename: mysql
# dataSourceName: <user>:<password>>@tcp(host:port)/live_api
```

```sql
CREATE DATABASE IF NOT EXISTS live_api;

CREATE USER '<user>'@'<host>' IDENTIFIED BY <password>;
GRANT ALL PRIVILEGES ON live_api.* TO '<user>'@'<host>';

USE live_api
CREATE TABLE IF NOT EXISTS subscriptions (
	id INT AUTO_INCREMENT,
	callback VARCHAR(2083) NOT NULL,
	callback_type VARCHAR(20) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS filter (
	id INT AUTO_INCREMENT NOT NULL,
	subscription INT NOT NULL,
	filtering TEXT,
    PRIMARY KEY (id),
    FOREIGN KEY (subscription) REFERENCES subscriptions(id)
);

``` 