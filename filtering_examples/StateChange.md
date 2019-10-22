State Change example
```graphql endpoint doc
{
    factomNodeName
    identityChainID 
    event {
        ... on StateChange {
            entityHash
            entityState 
            blockHeight
        }
    }
}
```