Entry Reveal example
```graphql endpoint doc
{
    factomNodeName
    identityChainID 
    event {
        ... on EntryReveal {
            timestamp
            entityState
            entry {
                version
                hash
                externalIDs 
                content
                chainID
            }
        }
    }
} 
```