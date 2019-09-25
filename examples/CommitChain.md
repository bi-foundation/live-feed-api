Commit Chain example
```graphql endpoint doc
{ 
    factomNodeName
    identityChainID {
        hashValue 
    }
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
    }
}
```