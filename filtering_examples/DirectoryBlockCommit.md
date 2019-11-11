Directory Block Commit example
```graphql endpoint doc
{
    factomNodeName
    identityChainID 
    event {
        ... on DirectoryBlockCommit {
            directoryBlock {
                header {
                    version
                    timestamp
                    blockHeight
                    blockCount
                    networkID
                    bodyMerkleRoot
                    previousKeyMerkleRoot
                    previousFullHash
                }
                entries {
                    chainID
                    keyMerkleRoot
                }
            }
            adminBlock {
                header {
                    blockHeight
                    previousBackRefHash
                    messageCount
                }
                entries {
                    adminBlockEntry {
                        ... on AddEfficiency {
                            identityChainID
                            efficiency
                        }
                        # ... on AddAuditServer { }
                        # ... on AddFactoidAddress { } 
                        # ... on AddFederatedServer { } 
                        # ... on AddFederatedServerBitcoinAnchorKey { } 
                        # ... on AddFederatedServerSigningKey { } 
                        # ... on AddReplaceMatryoshkaHash { } 
                        # ... on CancelCoinbaseDescriptor { } 
                        # ... on CoinbaseDescriptor { } 
                        # ... on DirectoryBlockSignatureEntry { } 
                        # ... on EndOfMinuteEntry { } 
                        # ... on ForwardCompatibleEntry { } 
                        # ... on IncreaseServerCount { } 
                        # ... on RemoveFederatedServer { } 
                        # ... on RevealMatryoshkaHash { } 
                        # ... on ServerFault { } 
                    }
                }
            }
            factoidBlock {
                bodyMerkleRoot
                previousKeyMerkleRoot
                previousLedgerKeyMerkleRoot
                exchangeRate
                blockHeight
                transactions {
                    transactionID
                    blockHeight
                    timestamp
                    factoidInputs {
                        amount
                        address
                    }
                    factoidOutputs {
                        amount
                        address
                    }
                    entryCreditOutputs {
                        amount
                        address
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
                }
            }
            entryCreditBlock {
                header {
                    bodyHash
                    previousHeaderHash
                    previousFullHash
                    blockHeight
                    objectCount
                }
                entries {
                        entryCreditBlockEntry { 
                        ... on ChainCommit {
                            version
                            timestamp
                            entityState
                            entryCreditPublicKey
                            signature
                            credits
                            entryHash
                            chainIDHash
                            weld
                        }
                        ... on EntryCommit {
                            version 
                            timestamp
                            entryHash
                            entityState 
                            credits 
                            entryCreditPublicKey 
                            signature 
                        }
                        ... on IncreaseBalance { 
                            entryCreditPublicKey
                            transactionID
                            index
                            amount
                        } 
                        ... on MinuteNumber { 
                            minuteNumber
                        } 
                        ... on ServerIndexNumber { 
                            serverIndexNumber
                        } 
                    }
                }
            }
            entryBlocks {
                header {
                    bodyMerkleRoot
                    chainID
                    previousFullHash
                    previousKeyMerkleRoot
                    blockHeight
                    blockSequence
                    entryCount
                }
                entryHashes
            }
            entryBlockEntries {
                version
                hash
                externalIDs 
                content
            }
        }
    }
}
```