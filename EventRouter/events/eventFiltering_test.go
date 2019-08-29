package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/gen/eventmessages"
	"github.com/graphql-go/graphql"
	"log"
	"testing"
)

func TestQueryFilter(t *testing.T) {
	schema, err := queryScheme()
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	// Query
	query := `
		query userModel {
			event { 
				eventSource
			}
		}
	`
	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)

	fmt.Printf("%v\n", r)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)
	fmt.Printf("query: %s \n", jsonPrettyPrint(query))
	fmt.Printf("result: %s \n", jsonPrettyPrint(string(rJSON)))
}

func queryScheme() (graphql.Schema, error) {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"event": &graphql.Field{
					Type: eventmessages.GraphQLFactomEventType,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return mockAnchorEvent(), nil
					},
				},
			},
		}),
	})
	return schema, err
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}
