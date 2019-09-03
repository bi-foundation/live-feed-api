package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/eventmessages"
	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestQueryNodeMessage(t *testing.T) {
	// Query
	query := ` { 
		IdentityChainID {
			hashValue 
		}
		value {
			... on NodeMessage {
				nodeMessageCode
				messageText
			}
		}
	}`
	expectedJson := `{ 
		"event": {
			"IdentityChainID": {
				"hashValue": "OLqxRVt71+Xv0VxTx3fHnQyYjpIQ8dpJqZ2Vs6ZBe+k="
			},
			"value": {
				"messageText": "New minute [6]",
				"nodeMessageCode": "SYNC_COMPLETE"
			}
		}
	}`

	result, err := Filter(query, mockNodeMessage())
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJson, string(result))
}

func testQueryFilter(t *testing.T) {
	schema, err := queryScheme(mockNodeMessage())
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

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}

func mockNodeMessage() *eventmessages.FactomEvent {
	return &eventmessages.FactomEvent{
		IdentityChainID: &eventmessages.Hash{
			HashValue: []byte("OLqxRVt71+Xv0VxTx3fHnQyYjpIQ8dpJqZ2Vs6ZBe+k="),
		},
		Value: &eventmessages.FactomEvent_NodeEvent{
			NodeEvent: &eventmessages.NodeMessage{
				NodeMessageCode: 2,
				MessageText:     "New minute [6]",
			},
		},
	}
}
