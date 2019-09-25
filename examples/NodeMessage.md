Node Message example
```graphql endpoint doc
{
    factomNodeName
    identityChainID {
        hashValue
    }
    value {
        ... on NodeMessage {
            messageText 
            messageCode 
            level
        }
    }
}
```