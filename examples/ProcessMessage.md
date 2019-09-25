Process Message example
```graphql endpoint doc
{
    factomNodeName
    identityChainID {
        hashValue
    }
    value {
        ... on ProcessMessage {
            messageText 
            messageCode 
            level
        }
    }
}
```