Commit Chain example
```graphql endpoint doc
{ 
    factomNodeName
    identityChainID 
    event {
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
    }
}
```