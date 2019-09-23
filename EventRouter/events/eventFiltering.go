package events

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/eventmessages/generated/eventmessages"
	"github.com/graphql-go/graphql"
)

// Filter an event with the given GraphQl filtering
func Filter(filtering string, event *eventmessages.FactomEvent) ([]byte, error) {
	// generate graphql scheme for event
	schema, err := queryScheme(event)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %v", err)
	}

	if filtering == "" {
		// return complete event if there isn't any filtering
		resultJSON, err := json.Marshal(event)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %v", err)
		}
		return resultJSON, nil
	}

	// inject filtering in query
	query := fmt.Sprintf(`{ event %s }`, filtering)
	params := graphql.Params{Schema: schema, RequestString: query}
	result := graphql.Do(params)

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("failed to execute graphql operation: %v", result.Errors)
	}

	resultJSON, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %v", err)
	}

	return resultJSON, nil
}

func queryScheme(event interface{}) (graphql.Schema, error) {
	return graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"event": &graphql.Field{
					Type: eventmessages.GraphQLFactomEventType,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return event, nil
					},
				},
			},
		}),
	})
}
