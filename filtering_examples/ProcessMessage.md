Process Message example
```graphql endpoint doc
{
    factomNodeName
    identityChainID 
    event {
        ... on ProcessMessage {
            messageText 
            processCode 
            level
        }
    }
}
```