package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/eventmessages/generated/eventmessages"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"testing"
)

var randomizer = Randomizer{}

func TestNoFilteringQuery(t *testing.T) {
	// this test is used to verify if no filtering produces the full result.
	// if the protobuf changes, make sure the non filtering query is updated.
	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	schema, err := queryScheme(event)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	// build the query
	query := buildNonFilteringQuery(schema)

	assert.EqualValues(t, query, nonFilteringQuery)
}

func TestQueryNoFiltering(t *testing.T) {
	eventTypes := []models.EventType{models.ChainCommit, models.EntryCommit, models.EntryReveal, models.DirectoryBlockCommit, models.ProcessListEvent, models.NodeMessage}

	for _, eventType := range eventTypes {
		t.Run(string(eventType), func(t *testing.T) {
			event := createNewEvent(eventType)

			schema, err := queryScheme(event)
			if err != nil {
				log.Fatalf("failed to create new schema, error: %v", err)
			}

			// build the query
			query := buildNonFilteringQuery(schema)

			assert.EqualValues(t, query, nonFilteringQuery)

			queryResult, err := Filter(query, event)
			if err != nil {
				fmt.Printf("result: %s \n", jsonPrettyPrint(string(queryResult)))
				t.Fatalf("failed to marshal result: %v - %v", err, queryResult)
			}

			filterResult, err := Filter("", event)
			if err != nil {
				fmt.Printf("result: %s \n", jsonPrettyPrint(string(filterResult)))
				t.Fatalf("failed to marshal result: %v - %v", err, filterResult)
			}

			assert.JSONEq(t, string(queryResult), string(filterResult))
		})
	}
}

func TestQueryOnDifferentEvent(t *testing.T) {
	query := readQuery(t, "CommitChain.md")
	expectedJSON := `{
  		"event": {
			"factomNodeName": "1",
			"identityChainID": "\u0001",
			"event": {
			}
		}
	}`

	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	chainCommit := eventmessages.NewPopulatedFactomEvent_EntryCommit(randomizer, false)
	event.Event = chainCommit

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryCommitChain(t *testing.T) {
	query := readQuery(t, "CommitChain.md")
	expectedJSON := `{
  		"event": {
			"factomNodeName": "1",
			"identityChainID": "\u0001",
			"event": {
				"version": 1,
				"timestamp": 1000,
				"entityState": "ACCEPTED",
				"entryCreditPublicKey": "\u0001",
				"signature": "\u0001",
				"credits": 1,
				"entryHash": "\u0001",
				"chainIDHash": "\u0001",
				"weld": "\u0001"
			}
		}
	}`

	event := createNewEvent(models.ChainCommit)

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryCommitEntry(t *testing.T) {
	query := readQuery(t, "CommitEntry.md")
	expectedJSON := `{
  		"event": {
			"factomNodeName": "1",
			"identityChainID": "\u0001",
			"event": {
				"version": 1,
				"timestamp": 1000,
				"entityState": "ACCEPTED",
				"entryCreditPublicKey": "\u0001",
				"signature": "\u0001",
				"credits": 1,
				"entryHash": "\u0001"
			}
		}
	}`

	event := createNewEvent(models.EntryCommit)

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryEntryReveal(t *testing.T) {
	query := readQuery(t, "EntryReveal.md")
	expectedJSON := `{
  		"event": {
			"factomNodeName": "1",
			"identityChainID": "\u0001",
			"event": {
			  "entityState": "ACCEPTED",
			  "timestamp": 1000,
			  "entry": {
                "chainID": "\u0001", 
			    "hash": "\u0001", 
			    "externalIDs": ["\u0001"],
			    "content": "\u0001", 
			    "version": 1
			  }
			}
		  }
		}`

	event := createNewEvent(models.EntryReveal)

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryStateChange(t *testing.T) {
	query := readQuery(t, "StateChange.md")
	expectedJSON := `{
  		"event": {
			"factomNodeName": "1",
			"identityChainID": "\u0001",
			"event": {
			  "entityState": "ACCEPTED",
			  "blockHeight": 1,
			  "entityHash": "\u0001"
			}
		  }
		}`

	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	entryCommit := eventmessages.NewPopulatedFactomEvent_StateChange(randomizer, false)
	event.Event = entryCommit

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryDirectoryBlockCommit(t *testing.T) {
	expectedJSON := `{
	  "event": {
		"factomNodeName": "1", 
		"identityChainID": "\u0001", 
		"event": {
		  "adminBlock": {
			"entries": [
			  {
				"adminBlockEntry": {
					"efficiency": 1, 
					"identityChainID": "\u0001"
				}
			  }
			], 
			"header": {
			  "blockHeight": 1, 
               "messageCount":1,
               "previousBackRefHash": "\u0001"
			}
		  }, 
		  "directoryBlock": {
			"chainID":"\u0001",
			"keyMerkleRoot":"\u0001",
			"hash":"\u0001",
			"entries": [
			  {
				"chainID": "\u0001", 
				"keyMerkleRoot": "\u0001"
			  }
			], 
			"header": {
			  "blockCount": 1, 
			  "blockHeight": 1, 
			  "bodyMerkleRoot": "\u0001", 
			  "networkID": 1, 
			  "previousFullHash": "\u0001", 
			  "previousKeyMerkleRoot": "\u0001", 
			  "timestamp": 1000, 
			  "version": 1
			}
		  }, 
		  "entryBlockEntries": [
			{
			  "content": "\u0001", 
			  "externalIDs": ["\u0001"], 
			  "hash": "\u0001", 
			  "version": 1
			}
		  ], 
		  "entryBlocks": [
			{
			  "entryHashes": [
				"\u0001"
			  ], 
			  "header": {
				"blockHeight": 1, 
				"blockSequence": 1, 
				"bodyMerkleRoot": "\u0001", 
				"chainID": "\u0001", 
				"entryCount": 1, 
				"previousFullHash": "\u0001", 
				"previousKeyMerkleRoot": "\u0001"
			  }
			}
		  ], 
		  "entryCreditBlock": {
			"entries": [
			  {
				"entryCreditBlockEntry": {
				  "credits": 1, 
				  "entityState": "ACCEPTED", 
				  "entryCreditPublicKey": "\u0001", 
				  "entryHash": "\u0001", 
				  "signature": "\u0001", 
				  "timestamp": 1000, 
				  "version": 1
				}
			  }
			], 
			"header": {
			  "blockHeight": 1, 
			  "bodyHash": "\u0001", 
			  "objectCount": 1, 
			  "previousFullHash": "\u0001", 
			  "previousHeaderHash": "\u0001"
			}
		  }, 
		  "factoidBlock": {
			"blockHeight": 1, 
			"bodyMerkleRoot": "\u0001", 
			"exchangeRate": 1, 
			"previousKeyMerkleRoot": "\u0001", 
			"previousLedgerKeyMerkleRoot": "\u0001", 
			"transactions": [
			  {
				"blockHeight": 1, 
				"entryCreditOutputs": [
				  {
					"address": "\u0001", 
					"amount": 1
				  }
				], 
				"factoidInputs": [
				  {
					"address": "\u0001", 
					"amount": 1
				  }
				], 
				"factoidOutputs": [
				  {
					"address": "\u0001", 
					"amount": 1
				  }
				],
				"minuteNumber": 1,
				"redeemConditionDataStructures": [
				  {
					"rcd": {
                      "publicKey": "\u0001"
					}
				  }
				], 
				"signatureBlocks": [
				  {
					"signature": ["\u0001"]
				  }
				], 
				"timestamp": 1000, 
				"transactionID": "\u0001"
			  }
			]
		  }
		}
	  }
	}`
	query := readQuery(t, "DirectoryBlockCommit.md")

	event := createNewEvent(models.DirectoryBlockCommit)

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryDirectoryBlockAnchor(t *testing.T) {
	query := readQuery(t, "DirectoryBlockAnchor.md")
	expectedJSON := `{
		"event": {
			"event": {
				"blockHeight": 1,
				"btcBlockHash": "\u0001",
				"btcBlockHeight": 1,
				"btcConfirmed": false,
				"btcTxHash": "\u0001",
				"btcTxOffset": 1,
				"directoryBlockHash": "\u0001",
				"directoryBlockMerkleRoot": "\u0001",
				"ethereumAnchorRecordEntryHash": "\u0001",
				"ethereumConfirmed": false,
				"timestamp": 1000
			},
			"factomNodeName": "1",
			"identityChainID": "\u0001"
		}
	}`

	event := createNewEvent(models.DirectoryBlockAnchor)

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryProcessMessage(t *testing.T) {
	query := readQuery(t, "ProcessListEvent.md")
	expectedJSON := `{
  		"event": {
			"factomNodeName": "1",
			"identityChainID": "\u0001",
	  		"event": {
				"processListEvent": {
				  "blockHeight": 1,
				  "newMinute": 1
				}
			}
		  }
		}`

	event := createNewEvent(models.ProcessListEvent)

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryNodeMessage(t *testing.T) {
	query := readQuery(t, "NodeMessage.md")
	expectedJSON := `{ 
		"event": {
			"factomNodeName": "1",
			"identityChainID": "\u0001",
			"event": {
			  "messageText": "1",
			  "messageCode": "STARTED",
			  "level": "WARNING"
			}
		}
	}`

	event := createNewEvent(models.NodeMessage)

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryFilter(t *testing.T) {
	event := eventmessages.NewPopulatedFactomEvent(Randomizer{}, true)
	schema, err := queryScheme(event)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	// Query
	query := `
		query userModel {
			event { 
				eventSource
			}
		}`
	expectedJSON := `{
		"data": {
			"event": {
				"eventSource": "REPLAY_BOOT"
			}
		}
	}`

	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)

	fmt.Printf("%v\n", r)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)

	assert.JSONEq(t, expectedJSON, string(rJSON))
}

func TestInvalidQuery(t *testing.T) {
	event := eventmessages.NewPopulatedFactomEvent(Randomizer{}, true)
	query := `
		 {
			fieldNotExists
		}`

	_, err := Filter(query, event)

	assert.EqualError(t, err, `failed to execute graphql operation: [Cannot query field "fieldNotExists" on type "FactomEvent".]`)
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}

func readQuery(t testing.TB, filename string) string {
	data, err := ioutil.ReadFile("../../filtering_examples/" + filename)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	// search for the graphql code block in the markdown
	re := regexp.MustCompile("(?s)```graphql endpoint doc(.*?)```.*?")
	match := re.FindStringSubmatch(string(data))
	if len(match) < 2 {
		t.Fatalf("failed to find query in markdown: %v", match)
	}
	query := match[1]

	return query
}

// Randomizer that is not random
type Randomizer struct{}

func (Randomizer) Float32() float32 {
	return 1
}
func (Randomizer) Float64() float64 {
	return 1
}
func (Randomizer) Int63() int64 {
	return 1
}
func (Randomizer) Int31() int32 {
	return 1
}
func (Randomizer) Uint32() uint32 {
	return 1
}
func (Randomizer) Intn(n int) int {
	return 1 % n
}

func buildNonFilteringQuery(schema graphql.Schema) string {
	query := traverseType(schema, 4, "FactomEvent")
	return fmt.Sprintf("{\n%s}", query)
}

func traverseType(schema graphql.Schema, indent int, object string) string {
	var builder strings.Builder
	indentation := strings.Repeat(" ", indent)
	typeSchema := queryInfo(schema, object)
	if rootSchema, ok := typeSchema.(map[string]interface{}); ok {
		if tSchema, ok := rootSchema["__type"].(map[string]interface{}); ok {
			if fSchema, ok := tSchema["fields"].([]interface{}); ok {
				for _, fields := range fSchema {
					if field, ok := fields.(map[string]interface{}); ok {
						if fieldType, ok := field["type"].(map[string]interface{}); ok {
							if fieldType["kind"] == "OBJECT" || fieldType["kind"] == "UNION" {
								fmt.Fprintf(&builder, "%s%v", indentation, field["name"])
								query := traverseType(schema, indent+2, fieldType["name"].(string))
								fmt.Fprintf(&builder, " {\n%s%s}\n", query, indentation)
							} else if fieldType["kind"] == "LIST" {
								if listType, ok := fieldType["ofType"].(map[string]interface{}); ok {
									fmt.Fprintf(&builder, "%s%v", indentation, field["name"])
									query := traverseType(schema, indent+2, listType["name"].(string))
									if len(query) > 1 {
										fmt.Fprintf(&builder, " {\n%s%s}\n", query, indentation)
									} else {
										fmt.Fprintf(&builder, "\n")
									}
								}
							} else {
								fmt.Fprintf(&builder, "%s%s\n", indentation, field["name"])
							}
						}
					}
				}
			}
			// handle unions
			if pSchema, ok := tSchema["possibleTypes"].([]interface{}); ok {
				for _, unionType := range pSchema {
					if option, ok := unionType.(map[string]interface{}); ok {
						// fmt.Printf("%s...on %v: %v \n", indentation, option["name"], option["type"])
						if option["kind"] == "OBJECT" {
							fmt.Fprintf(&builder, "%s... on %v {\n", indentation, option["name"])
							query := traverseType(schema, indent+2, option["name"].(string))
							fmt.Fprintf(&builder, "%s%s}\n", query, indentation)
						} else {
							fmt.Fprintf(&builder, "%s... on %s\n", indentation, option["name"])
						}

					}
				}
			}
		}
	}
	return builder.String()
}

func queryInfo(schema graphql.Schema, object string) interface{} {
	query := `{
	  __type(name: "` + object + `") {
		name
		kind
		description
		fields {
		  name
		  description
		  type {
			name
			kind
			ofType {
			  name
			  kind
			}
		  }
		}
		possibleTypes {
		  name
		  kind
		  description
		}
	  }
	}`
	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}

	return r.Data
}

func BenchmarkFilters(b *testing.B) {
	eventTypes := []struct {
		EventType models.EventType
		Filtering string
	}{
		{models.DirectoryBlockAnchor, readQuery(b, "DirectoryBlockAnchor.md")},
		{models.DirectoryBlockCommit, readQuery(b, "DirectoryBlockCommit.md")},
		{models.ChainCommit, readQuery(b, "CommitChain.md")},
		{models.EntryCommit, readQuery(b, "CommitEntry.md")},
		{models.EntryReveal, readQuery(b, "EntryReveal.md")},
		{models.ProcessListEvent, readQuery(b, "ProcessListEvent.md")},
		{models.NodeMessage, readQuery(b, "NodeMessage.md")},
	}

	for _, benchmark := range eventTypes {
		b.Run(string(benchmark.EventType), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				event := createNewEvent(benchmark.EventType)
				result, err := Filter(benchmark.Filtering, event)

				if err != nil {
					b.Fatalf("failed to marshal result: %v - %v", err, jsonPrettyPrint(string(result)))
				}
			}
		})
	}
}

func createNewEvent(eventType models.EventType) *eventmessages.FactomEvent {
	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	switch eventType {
	case models.DirectoryBlockAnchor:
		event.Event = eventmessages.NewPopulatedFactomEvent_DirectoryBlockAnchor(randomizer, false)
	case models.DirectoryBlockCommit:
		event.Event = eventmessages.NewPopulatedFactomEvent_DirectoryBlockCommit(randomizer, false)
	case models.ChainCommit:
		event.Event = eventmessages.NewPopulatedFactomEvent_ChainCommit(randomizer, false)
	case models.EntryCommit:
		event.Event = eventmessages.NewPopulatedFactomEvent_EntryCommit(randomizer, false)
	case models.EntryReveal:
		event.Event = eventmessages.NewPopulatedFactomEvent_EntryReveal(randomizer, false)
	case models.ProcessListEvent:
		event.Event = eventmessages.NewPopulatedFactomEvent_ProcessListEvent(randomizer, false)
	case models.NodeMessage:
		event.Event = eventmessages.NewPopulatedFactomEvent_NodeMessage(randomizer, false)
	case models.StateChange:
		event.Event = eventmessages.NewPopulatedFactomEvent_StateChange(randomizer, false)
	}

	return event
}
