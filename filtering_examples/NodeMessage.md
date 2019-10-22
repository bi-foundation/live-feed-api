Node Message example
```graphql endpoint doc
{
    factomNodeName
    identityChainID
    event {
        ... on NodeMessage {
            messageText 
            messageCode 
            level
        }
    }
}
```