# Live Feed API
The Live Feed API is a second layer application for the [factom](https://github.com/FactomProject/factomd) blockchain. The API emits live events from factomd to subscribers. Users can subscribe on receiving events.


## Contents
- [Getting started](#getting-started)
  - [Configuration](#configuration)
  - [Setup database](#setup-database)
  - [Starting Live Feed API](#starting-live-feed-api)
- [Using the Live Feed API](#using-the-live-feed-api)
  - [Subscriptions](#subscriptions)
- [Live Feed API Development](#live-feed-api-development)
	- [Generate sources](#generate-sources)

## Getting started

#### Prerequisites
install go 1.12 or higher


### Configuration
The configuration is stored in a file `factom-live-feed.conf` by default. The file can be stored in `/etc/factom-live-feed`, `$FACTOM_HOME` or `$FACTOM_HOME/factom-live-feed`. 
The configuration file can also be overridden using a command line flag 
Environment variables can be used to override properties of the configuration. For example, to set the receiver port use environment key: `FACTOM_LIVE_FEED_RECEIVER_PORT`.

| Property                   | Description                                                                         | Values      | Default |
| -------------------------- | ----------------------------------------------------------------------------------- | ------------------ | --------- |
| receiver / bindaddress         | The Network Interface address where the event listener needs to bind to.            | IP address         | 127.0.0.1
| receiver / port                | The event listener network port.                                                    | port number        | 8040
| receiver / protocol            | The network protocol that is used to receive event messages from the network.       | tcp                | tcp
| subscription / bindaddress     | The Network Interface address where the subscription API listener needs to bind to. | IP address         | 0.0.0.0 
| subscription / port            | The event listener network port.                                                    | port number        | 8700
| subscription / schemes         | The protocol schemes                                                                | HTTP or HTTPS | HTTP  
| subscription / certificatefile | Path to the certificate file to run the subscription api with TLS                   | /path/server.crt 
| subscription / privatekeyfile  | Path to the private key file corresponding to the certificate file                  | /path/server.key 
| database / database            | The type of database that will be used                                              | mysql or inmemory                  | mysql
| database / connectionString    | The connection string to connect to the database                                    | factom-live-api:<password>@tcp(<ip>:<port>)/<database> | 
| log / loglevel                 | The log level                                                                       | debug, info, warning, error, fatal | info


This is what factom-live-feed.conf looks like with the default settings:

```
[receiver]
  bindaddress = "127.0.0.1"
  port = "8040"
  protocol = "tcp"

[subscription]
  bindaddress = "0.0.0.0"
  port = "8700"
  schemes = "HTTPS"
  
[database]
  database = "mysql"
  connectionString = "factom-live-api:<password>@tcp(<ip>:<port>)/<database>"
  
[log]
  loglevel = "info"
```


### Setup Database
The Live Feed API needs to be able to store subscriptions in a database. An in-memory database can be used for rapid development. Note: this should not be used in production as after closing the application the subscriptions be lost. Alternative a MySQL database can be used.  

#### MYSQL database
Configuration for the mysql database.
```
# drivename: mysql
# dataSourceName: <user>:<password>>@tcp(host:port)/live_api
```

The following sql should be executed to create the tables in the database.
```mysql
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    failures int NOT NULL,
    callback VARCHAR(2083) NOT NULL,
    callback_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    info VARCHAR(200),
    access_token VARCHAR(255),
    username VARCHAR(255),
    password VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS filters (
    id SERIAL PRIMARY KEY,
    subscription BIGINT(20) REFERENCES subscriptions(id),
    event_type VARCHAR(20) NOT NULL,
    filtering TEXT
);
``` 

### Starting Live Feed API
Use go run to start the live feed API. To provide a custom configuration use the flag: --config-file "custom-configuration.conf".  
```shell script
go run ./live-feed-api.go
```

## Using the Live Feed API
Users can receive events from factomd by subscribing an application with the Live Feed API. 

A [swagger](EventRouter/swagger/swagger.yaml) is provided for the Subscription API. The swagger is also exposed at https://domain/live/feed/v0.1/swagger.json.

Users can receive the following event types:
* chain registration
* entry registration
* entry content registration
* block commit
* process message
* node message

Each of these event types can be filtered to reduce network traffic. Filtering is done with writing a query in [GraphQL](https://graphql.org/learn/).
```graphql endpoint doc
{ 
    identityChainID
    value { 
        ... on ProcessMessage { 
            messageCode
            messageText 
        }
    } 
}
```

Will result in the following event: 
```json
{
    "identityChainID": {
        "hashValue": "OLqxRVt71+Xv0VxTx3fHnQyYjpIQ8dpJqZ2Vs6ZBe+k="
    },
    "Value": {
        "ProcessMessage": {
            "messageCode": 2,
            "messageText": "New minute [6]"
        }
    }
}
``` 

### Subscriptions
Below is an example to create a subscription. In the example, the user registers the endpoint `https://server/events` to receive events. The user exposes the endpoint and has secured it with an API token. In the subscription request, the user sets the callback type on `BEARER_TOKEN` and sets the access token in the credentials field. As the user wants to receive all events it creates for each event type an entry in the filters field. The filtering itself is empty to receive the complete event. Users can filter the event with Graph QL to reduce the network traffic or receive only part of the events.   
```
POST /live/feed/v0.1/subscriptions
```
```json
{
  "callbackType": "BEARER_TOKEN", 
  "callbackUrl": "https://server/events", 
  "credentials": {
    "accessToken": "API_TOKEN_OF_THE_RECEIVING_ENDPOINT"
  }, 
  "filters": {
    "BLOCK_COMMIT": {
      "filtering": ""
    }, 
    "CHAIN_REGISTRATION": {
      "filtering": ""
    }, 
    "ENTRY_CONTENT_REGISTRATION": {
      "filtering": ""
    }, 
    "ENTRY_REGISTRATION": {
      "filtering": ""
    }, 
    "NODE_MESSAGE": {
      "filtering": ""
    }, 
    "PROCESS_MESSAGE": {
      "filtering": ""
    }
  }
}

```

## Live Feed API Development
The Live Feed API uses sources that are generated. The sources are provided but need to be updated if the API changes. If models are changed, these files needed to be regenerated. 

For the communication between factomd and the live-feed API the models are written as protobuf. The code for serialization and deserialization is generated with protoc. Further information: [golang/protobuf](https://github.com/golang/protobuf). 
As the API also provides event filtering with GraphQL a plugin is needed to generate the schemes of the models. The plugin that is used: [protobuf-graphql-extension](https://github.com/bi-foundation/protobuf-graphql-extension).

The swagger that established the contract for the subscription API is also generated. Generating the swagger is done with [swaggo/swag](https://github.com/swaggo/swag). In the code, comments provide information about the API. The comments are used as input for the swagger.
     
### Generate sources
Retrieving the develop dependencies will install swag, protoc-gen-go, protobuf-graphql-extension.

#### Prerequisites
A [protocol buffers](https://github.com/protocolbuffers/protobuf) compiler needs to be installed. Use the [install-protoc.sh](install-protoc.sh) to install protoc on linux or look at the [installation manual](https://github.com/protocolbuffers/protobuf#protocol-compiler-installation) for other platforms.   
```
make dev-deps
make generate
```

