Entry Reveal example
```graphql endpoint doc
{
    factomNodeName
    identityChainID {
        hashValue
    }
    value {
        ... on EntryReveal {
            timestamp
            entityState
            entry {
                version
                hash { hashValue }
                externalIDs { binaryValue } 
                content { binaryValue }
            }
            chainID { hashValue }
        }
    }
} 
```