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
		filtering = nonFilteringQuery
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

var nonFilteringQuery = `{
    eventSource
    factomNodeName
    identityChainID {
      hashValue
    }
    value {
      ... on ChainCommit {
        chainIDHash {
          hashValue
        }
        credits
        entityState
        entryCreditPublicKey
        entryHash {
          hashValue
        }
        signature
        timestamp
        version
        weld {
          hashValue
        }
      }
      ... on EntryCommit {
        credits
        entityState
        entryCreditPublicKey
        entryHash {
          hashValue
        }
        signature
        timestamp
        version
      }
      ... on EntryReveal {
        chainID {
          hashValue
        }
        entityState
        entry {
          content {
            binaryValue
          }
          externalIDs {
            binaryValue
          }
          hash {
            hashValue
          }
          version
        }
        timestamp
      }
      ... on StateChange {
        blockHeight
        entityHash {
          hashValue
        }
        entityState
      }
      ... on DirectoryBlockCommit {
        adminBlock {
          entries {
            value {
              ... on AddAuditServer {
                blockHeight
                identityChainID {
                  hashValue
                }
              }
              ... on AddEfficiency {
                efficiency
                identityChainID {
                  hashValue
                }
              }
              ... on AddFactoidAddress {
                address {
                  hashValue
                }
                identityChainID {
                  hashValue
                }
              }
              ... on AddFederatedServer {
                blockHeight
                identityChainID {
                  hashValue
                }
              }
              ... on AddFederatedServerBitcoinAnchorKey {
                ecdsaPublicKey
                identityChainID {
                  hashValue
                }
                keyPriority
                keyType
              }
              ... on AddFederatedServerSigningKey {
                blockHeight
                identityChainID {
                  hashValue
                }
                keyPriority
                publicKey
              }
              ... on AddReplaceMatryoshkaHash {
                factoidOutputs {
                  address {
                    hashValue
                  }
                  amount
                }
                identityChainID {
                  hashValue
                }
                matryoshkaHash {
                  hashValue
                }
              }
              ... on CancelCoinbaseDescriptor {
                descriptorHeight
                descriptorIndex
              }
              ... on CoinbaseDescriptor {
                factoidOutputs {
                  address {
                    hashValue
                  }
                  amount
                }
              }
              ... on DirectoryBlockSignatureEntry {
                identityAdminChainID {
                  hashValue
                }
                previousDirectoryBlockSignature {
                  publicKey
                  signature
                }
              }
              ... on EndOfMinuteEntry {
                minuteNumber
              }
              ... on ForwardCompatibleEntry {
                data
                size
              }
              ... on IncreaseServerCount {
                amount
              }
              ... on RemoveFederatedServer {
                blockHeight
                identityChainID {
                  hashValue
                }
              }
              ... on RevealMatryoshkaHash {
                identityChainID {
                  hashValue
                }
                matryoshkaHash {
                  hashValue
                }
              }
              ... on ServerFault {
                auditServerID {
                  hashValue
                }
                blockHeight
                messageEntryHeight
                serverID {
                  hashValue
                }
                signatureList {
                  publicKey
                  signature
                }
                timestamp
                vmIndex
              }
            }
          }
          header {
            blockHeight
            bodySize
            headerExpansionArea
            headerExpansionSize
            messageCount
            previousBackRefHash {
              hashValue
            }
          }
        }
        directoryBlock {
          entries {
            chainID {
              hashValue
            }
            keyMerkleRoot {
              hashValue
            }
          }
          header {
            blockCount
            blockHeight
            bodyMerkleRoot {
              hashValue
            }
            networkID
            previousFullHash {
              hashValue
            }
            previousKeyMerkleRoot {
              hashValue
            }
            timestamp
            version
          }
        }
        entryBlockEntries {
          content {
            binaryValue
          }
          externalIDs {
            binaryValue
          }
          hash {
            hashValue
          }
          version
        }
        entryBlocks {
          entryHashes {
            hashValue
          }
          header {
            blockHeight
            blockSequence
            bodyMerkleRoot {
              hashValue
            }
            chainID {
              hashValue
            }
            entryCount
            previousFullHash {
              hashValue
            }
            previousKeyMerkleRoot {
              hashValue
            }
          }
        }
        entryCreditBlock {
          entries {
            value {
              ... on ChainCommit {
                chainIDHash {
                  hashValue
                }
                credits
                entityState
                entryCreditPublicKey
                entryHash {
                  hashValue
                }
                signature
                timestamp
                version
                weld {
                  hashValue
                }
              }
              ... on EntryCommit {
                credits
                entityState
                entryCreditPublicKey
                entryHash {
                  hashValue
                }
                signature
                timestamp
                version
              }
              ... on IncreaseBalance {
                amount
                entryCreditPublicKey
                index
                transactionID {
                  hashValue
                }
              }
              ... on MinuteNumber {
                minuteNumber
              }
              ... on ServerIndexNumber {
                serverIndexNumber
              }
            }
          }
          header {
            blockHeight
            bodyHash {
              hashValue
            }
            bodySize
            headerExpansionArea
            objectCount
            previousFullHash {
              hashValue
            }
            previousHeaderHash {
              hashValue
            }
          }
        }
        factoidBlock {
          blockHeight
          bodyMerkleRoot {
            hashValue
          }
          exchangeRate
          previousKeyMerkleRoot {
            hashValue
          }
          previousLedgerKeyMerkleRoot {
            hashValue
          }
          transactions {
            blockHeight
            entryCreditOutputs {
              address {
                hashValue
              }
              amount
            }
            factoidInputs {
              address {
                hashValue
              }
              amount
            }
            factoidOutputs {
              address {
                hashValue
              }
              amount
            }
            redeemConditionDataStructures {
              value {
                ... on RCD1 {
                  publicKey
                }
              }
            }
            signatureBlocks {
              signature {
                signatureValue
              }
            }
            timestamp
            transactionID {
              hashValue
            }
          }
        }
      }
      ... on ProcessMessage {
        level
        messageText
        processCode
      }
      ... on NodeMessage {
        level
        messageCode
        messageText
      }
    }
}`
