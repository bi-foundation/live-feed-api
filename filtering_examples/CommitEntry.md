Commit Entry example
```graphql endpoint doc
{
    factomNodeName
    identityChainID 
    event {
        ... on EntryCommit {
            version
            timestamp
            entityState
            entryCreditPublicKey
            signature
            credits
            entryHash
        }
    }
}
```