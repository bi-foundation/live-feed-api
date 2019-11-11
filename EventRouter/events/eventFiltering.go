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
    event {
      ... on ChainCommit {
        chainIDHash
        credits
        entityState
        entryCreditPublicKey
        entryHash
        signature
        timestamp
        version
        weld
      }
      ... on EntryCommit {
        credits
        entityState
        entryCreditPublicKey
        entryHash
        signature
        timestamp
        version
      }
      ... on EntryReveal {
        entityState
        entry {
          chainID
          content
          externalIDs
          hash
          version
        }
        timestamp
      }
      ... on StateChange {
        blockHeight
        entityHash
        entityState
      }
      ... on DirectoryBlockCommit {
        adminBlock {
          entries {
            adminBlockEntry {
              ... on AddAuditServer {
                blockHeight
                identityChainID
              }
              ... on AddEfficiency {
                efficiency
                identityChainID
              }
              ... on AddFactoidAddress {
                address
                identityChainID
              }
              ... on AddFederatedServer {
                blockHeight
                identityChainID
              }
              ... on AddFederatedServerBitcoinAnchorKey {
                ecdsaPublicKey
                identityChainID
                keyPriority
                keyType
              }
              ... on AddFederatedServerSigningKey {
                blockHeight
                identityChainID
                keyPriority
                publicKey
              }
              ... on AddReplaceMatryoshkaHash {
                factoidOutputs {
                  address
                  amount
                }
                identityChainID
                matryoshkaHash
              }
              ... on CancelCoinbaseDescriptor {
                descriptorHeight
                descriptorIndex
              }
              ... on CoinbaseDescriptor {
                factoidOutputs {
                  address
                  amount
                }
              }
              ... on DirectoryBlockSignatureEntry {
                identityAdminChainID
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
                identityChainID
              }
              ... on RevealMatryoshkaHash {
                identityChainID
                matryoshkaHash
              }
              ... on ServerFault {
                auditServerID
                blockHeight
                messageEntryHeight
                serverID
                signatureList {
                  publicKey
                  signature
                }
                timestamp
                vmIndex
              }
            }
            adminIdType
          }
          header {
            blockHeight
            messageCount
            previousBackRefHash
          }
          keyMerkleRoot
        }
        directoryBlock {
          entries {
            chainID
            keyMerkleRoot
          }
          header {
            blockCount
            blockHeight
            bodyMerkleRoot
            networkID
            previousFullHash
            previousKeyMerkleRoot
            timestamp
            version
          }
        }
        entryBlockEntries {
          chainID
          content
          externalIDs
          hash
          version
        }
        entryBlocks {
          entryHashes
          header {
            blockHeight
            blockSequence
            bodyMerkleRoot
            chainID
            entryCount
            previousFullHash
            previousKeyMerkleRoot
          }
        }
        entryCreditBlock {
          entries {
            entryCreditBlockEntry {
              ... on ChainCommit {
                chainIDHash
                credits
                entityState
                entryCreditPublicKey
                entryHash
                signature
                timestamp
                version
                weld
              }
              ... on EntryCommit {
                credits
                entityState
                entryCreditPublicKey
                entryHash
                signature
                timestamp
                version
              }
              ... on IncreaseBalance {
                amount
                entryCreditPublicKey
                index
                transactionID
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
            bodyHash
            objectCount
            previousFullHash
            previousHeaderHash
          }
        }
        factoidBlock {
          blockHeight
          bodyMerkleRoot
          exchangeRate
          keyMerkleRoot
          previousKeyMerkleRoot
          previousLedgerKeyMerkleRoot
          transactionCount
          transactions {
            blockHeight
            entryCreditOutputs {
              address
              amount
            }
            factoidInputs {
              address
              amount
            }
            factoidOutputs {
              address
              amount
            }
            redeemConditionDataStructures {
              rcd {
                ... on RCD1 {
                  publicKey
                }
              }
            }
            signatureBlocks {
              signature
            }
            timestamp
            transactionID
          }
        }
      }
      ... on ProcessListEvent {
        processListEvent {
          ... on NewBlockEvent {
            newBlockHeight
          }
          ... on NewMinuteEvent {
            blockHeight
            newMinute
          }
        }
      }
      ... on NodeMessage {
        level
        messageCode
        messageText
      }
    }
    eventSource
    factomNodeName
    identityChainID
}`
