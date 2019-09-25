Commit Entry example
```graphql endpoint doc
{
    factomNodeName
    identityChainID {
        hashValue
    }
    value {
        ... on EntryCommit {
            version
            timestamp
            entityState
            entryCreditPublicKey
            signature
            credits
            entryHash { hashValue }
        }
    }
}
```