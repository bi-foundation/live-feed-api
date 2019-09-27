package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/eventmessages/generated/eventmessages"
	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"testing"
)

var randomizer = Randomizer{}

func TestQueryNoFiltering(t *testing.T) {
	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	event.Value = eventmessages.NewPopulatedFactomEvent_DirectoryBlockCommit(randomizer, false)

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
}

func TestQueryOnDifferentEvent(t *testing.T) {
	query := readQuery(t, "CommitChain.md")
	expectedJSON := `{
  		"event": {
			"factomNodeName": "1",
			"identityChainID": {
				"hashValue": "\u0001"
			},
			"value": {
			}
		}
	}`

	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	chainCommit := eventmessages.NewPopulatedFactomEvent_EntryCommit(randomizer, false)
	event.Value = chainCommit

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
			"identityChainID": {
				"hashValue": "\u0001"
			},
			"value": {
				"version": 1,
				"timestamp": 1000,
				"entityState": "HOLDING_QUEUE",
				"entryCreditPublicKey": "\u0001",
				"signature": "\u0001",
				"credits": 1,
				"entryHash": {
					"hashValue": "\u0001"
				},
				"chainIDHash": {
					"hashValue": "\u0001"
				},
				"weld": {
					"hashValue": "\u0001"
				}
			}
		}
	}`

	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	chainCommit := eventmessages.NewPopulatedFactomEvent_ChainCommit(randomizer, false)
	event.Value = chainCommit

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
			"identityChainID": {
				"hashValue": "\u0001"
			},
			"value": {
				"version": 1,
				"timestamp": 1000,
				"entityState": "HOLDING_QUEUE",
				"entryCreditPublicKey": "\u0001",
				"signature": "\u0001",
				"credits": 1,
				"entryHash": {
					"hashValue": "\u0001"
				}
			}
		}
	}`

	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	entryCommit := eventmessages.NewPopulatedFactomEvent_EntryCommit(randomizer, false)
	event.Value = entryCommit

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
			"identityChainID": {
				"hashValue": "\u0001"
			},
			"value": {
			  "entityState": "HOLDING_QUEUE",
			  "timestamp": 1000,
			  "entry": { 
			    "hash": {
				  "hashValue": "\u0001"
			    }, 
			    "externalIDs": [{
				  "binaryValue": "\u0001"
			    }],
			    "content": {
				  "binaryValue": "\u0001"
			    }, 
			    "version": 1
			  }
			}
		  }
		}`

	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	entryCommit := eventmessages.NewPopulatedFactomEvent_EntryReveal(randomizer, false)
	event.Value = entryCommit

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
			"identityChainID": {
				"hashValue": "\u0001"
			},
			"value": {
			  "entityState": "HOLDING_QUEUE",
			  "blockHeight": 1,
			  "entityHash": { 
				"hashValue": "\u0001"
			  }
			}
		  }
		}`

	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	entryCommit := eventmessages.NewPopulatedFactomEvent_StateChange(randomizer, false)
	event.Value = entryCommit

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryDirectoryBlockCommit(t *testing.T) {
	query := readQuery(t, "DirectoryBlockCommit.md")
	expectedJSON := `{
	  "event": {
		"factomNodeName": "1", 
		"identityChainID": {
		  "hashValue": "\u0001"
		}, 
		"value": {
		  "adminBlock": {
			"entries": [
			  {
				"value": {
					"efficiency": 1, 
					"identityChainID": {
					  "hashValue": "\u0001"
					}
				}
			  }
			], 
			"header": {
			  "blockHeight": 1, 
			  "previousBackRefHash": {
				"hashValue": "\u0001"
			  }
			}
		  }, 
		  "directoryBlock": {
			"entries": [
			  {
				"chainID": {
				  "hashValue": "\u0001"
				}, 
				"keyMerkleRoot": {
				  "hashValue": "\u0001"
				}
			  }
			], 
			"header": {
			  "blockCount": 1, 
			  "blockHeight": 1, 
			  "bodyMerkleRoot": {
				"hashValue": "\u0001"
			  }, 
			  "networkID": 1, 
			  "previousFullHash": {
				"hashValue": "\u0001"
			  }, 
			  "previousKeyMerkleRoot": {
				"hashValue": "\u0001"
			  }, 
			  "timestamp": 1000, 
			  "version": 1
			}
		  }, 
		  "entryBlockEntries": [
			{
			  "content": {
				"binaryValue": "\u0001"
			  }, 
			  "externalIDs": [
				{
				  "binaryValue": "\u0001"
				}
			  ], 
			  "hash": {
				"hashValue": "\u0001"
			  }, 
			  "version": 1
			}
		  ], 
		  "entryBlocks": [
			{
			  "entryHashes": [
				{
				  "hashValue": "\u0001"
				}
			  ], 
			  "header": {
				"blockHeight": 1, 
				"blockSequence": 1, 
				"bodyMerkleRoot": {
				  "hashValue": "\u0001"
				}, 
				"chainID": {
				  "hashValue": "\u0001"
				}, 
				"entryCount": 1, 
				"previousFullHash": {
				  "hashValue": "\u0001"
				}, 
				"previousKeyMerkleRoot": {
				  "hashValue": "\u0001"
				}
			  }
			}
		  ], 
		  "entryCreditBlock": {
			"entries": [
			  {
				"value": {
				  "credits": 1, 
				  "entityState": "HOLDING_QUEUE", 
				  "entryCreditPublicKey": "\u0001", 
				  "entryHash": {
				    "hashValue": "\u0001"
				  }, 
				  "signature": "\u0001", 
				  "timestamp": 1000, 
				  "version": 1
				}
			  }
			], 
			"header": {
			  "blockHeight": 1, 
			  "bodyHash": {
				"hashValue": "\u0001"
			  }, 
			  "bodySize": 1, 
			  "headerExpansionArea": "\u0001", 
			  "objectCount": 1, 
			  "previousFullHash": {
				"hashValue": "\u0001"
			  }, 
			  "previousHeaderHash": {
				"hashValue": "\u0001"
			  }
			}
		  }, 
		  "factoidBlock": {
			"blockHeight": 1, 
			"bodyMerkleRoot": {
			  "hashValue": "\u0001"
			}, 
			"exchangeRate": 1, 
			"previousKeyMerkleRoot": {
			  "hashValue": "\u0001"
			}, 
			"previousLedgerKeyMerkleRoot": {
			  "hashValue": "\u0001"
			}, 
			"transactions": [
			  {
				"blockHeight": 1, 
				"entryCreditOutputs": [
				  {
					"address": {
					  "hashValue": "\u0001"
					}, 
					"amount": 1
				  }
				], 
				"factoidInputs": [
				  {
					"address": {
					  "hashValue": "\u0001"
					}, 
					"amount": 1
				  }
				], 
				"factoidOutputs": [
				  {
					"address": {
					  "hashValue": "\u0001"
					}, 
					"amount": 1
				  }
				], 
				"redeemConditionDataStructures": [
				  {
					"value": {
                      "publicKey": "\u0001"
					}
				  }
				], 
				"signatureBlocks": [
				  {
					"signature": [
					  {
						"signatureValue": "\u0001"
					  }
					]
				  }
				], 
				"timestamp": 1000, 
				"transactionID": {
				  "hashValue": "\u0001"
				}
			  }
			]
		  }
		}
	  }
	}`

	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	directoryBlockCommit := eventmessages.NewPopulatedFactomEvent_DirectoryBlockCommit(randomizer, false)
	event.Value = directoryBlockCommit

	result, err := Filter(query, event)
	if err != nil {
		fmt.Printf("query: %s \n", jsonPrettyPrint(query))
		fmt.Printf("result: %s \n", jsonPrettyPrint(string(result)))
		t.Fatalf("failed to marshal result: %v - %v", err, result)
	}

	assert.JSONEq(t, expectedJSON, string(result))
}

func TestQueryProcessMessage(t *testing.T) {
	query := readQuery(t, "ProcessMessage.md")
	expectedJSON := `{
  		"event": {
			"factomNodeName": "1",
			"identityChainID": {
				"hashValue": "\u0001"
			},
			"value": {
			  "messageText": "1",
			  "processCode": "NEW_MINUTE",
			  "level": "WARNING"
			}
		  }
		}`

	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	entryCommit := eventmessages.NewPopulatedFactomEvent_ProcessMessage(randomizer, false)
	event.Value = entryCommit

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
			"identityChainID": {
				"hashValue": "\u0001"
			},
			"value": {
			  "messageText": "1",
			  "messageCode": "STARTED",
			  "level": "WARNING"
			}
		}
	}`

	event := eventmessages.NewPopulatedFactomEvent(randomizer, false)
	entryCommit := eventmessages.NewPopulatedFactomEvent_NodeMessage(randomizer, false)
	event.Value = entryCommit

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

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}

func readQuery(t *testing.T, filename string) string {
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
						// fmt.Printf("%s%v : %v \n", indentation, field["name"], field["type"])
						if fieldType, ok := field["type"].(map[string]interface{}); ok {
							if fieldType["kind"] == "OBJECT" || fieldType["kind"] == "UNION" {
								fmt.Fprintf(&builder, "%s%v {\n", indentation, field["name"])
								query := traverseType(schema, indent+2, fieldType["name"].(string))
								fmt.Fprintf(&builder, "%s%s}\n", query, indentation)
							} else if fieldType["kind"] == "LIST" {
								if listType, ok := fieldType["ofType"].(map[string]interface{}); ok {
									fmt.Fprintf(&builder, "%s%v {\n", indentation, field["name"])
									query := traverseType(schema, indent+2, listType["name"].(string))
									fmt.Fprintf(&builder, "%s%s}\n", query, indentation)
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
