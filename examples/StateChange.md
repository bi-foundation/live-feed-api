State Change example
```graphql endpoint doc
{
    factomNodeName
    identityChainID {
        hashValue
    }
    value {
        ... on StateChange {
            entityHash { hashValue }
            entityState 
            blockHeight
        }
    }
}
```