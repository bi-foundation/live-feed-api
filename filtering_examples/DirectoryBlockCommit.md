Directory Block Commit example
```graphql endpoint doc
{
    factomNodeName
    identityChainID {
        hashValue
    }
    value {
        ... on DirectoryBlockCommit {
            directoryBlock {
                header {
                    version
                    timestamp
                    blockHeight
                    blockCount
                    networkID
                    bodyMerkleRoot { hashValue }
                    previousKeyMerkleRoot { hashValue }
                    previousFullHash { hashValue }
                }
                entries {
                    chainID { hashValue }
                    keyMerkleRoot { hashValue }
                }
            }
            adminBlock {
                header {
                    blockHeight
                    previousBackRefHash { hashValue }
                    headerExpansionSize
                    headerExpansionArea
                    messageCount
                    bodySize
                }
                entries {
                    value {
                        ... on AddEfficiency {
                            identityChainID { hashValue }
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
                bodyMerkleRoot { hashValue }
                previousKeyMerkleRoot { hashValue }
                previousLedgerKeyMerkleRoot { hashValue }
                exchangeRate
                blockHeight
                transactions {
                    transactionID { hashValue }
                    blockHeight
                    timestamp
                    factoidInputs {
                        amount
                        address { hashValue }
                    }
                    factoidOutputs {
                        amount
                        address { hashValue }
                    }
                    entryCreditOutputs {
                        amount
                        address { hashValue }
                    }
                    redeemConditionDataStructures {
                        value {
                            ... on RCD1 {
                                publicKey 
                            }
                        }
                    }
                    signatureBlocks {
                        signature { signatureValue }
                    } 
                }
            }
            entryCreditBlock {
                header {
                    bodyHash { hashValue }
                    previousHeaderHash { hashValue }
                    previousFullHash { hashValue }
                    blockHeight
                    headerExpansionArea
                    objectCount
                    bodySize
                }
                entries {
                    value { 
                        ... on ChainCommit {
                            version
                            timestamp
                            entityState
                            entryCreditPublicKey
                            signature
                            credits
                            entryHash { hashValue }
                            chainIDHash { hashValue }
                            weld { hashValue }
                        }
                        ... on EntryCommit {
                            version 
                            timestamp
                            entryHash { hashValue }
                            entityState 
                            credits 
                            entryCreditPublicKey 
                            signature 
                        }
                        ... on IncreaseBalance { 
                            entryCreditPublicKey
                            transactionID { hashValue }
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
                    bodyMerkleRoot { hashValue }
                    chainID { hashValue }
                    previousFullHash { hashValue }
                    previousKeyMerkleRoot { hashValue }
                    blockHeight
                    blockSequence
                    entryCount
                }
                entryHashes { hashValue }
            }
            entryBlockEntries {
                version
                hash { hashValue }
                externalIDs { binaryValue } 
                content { binaryValue }
            }
        }
    }
}
```