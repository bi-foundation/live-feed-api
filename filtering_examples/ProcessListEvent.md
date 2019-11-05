Process List Event example
```graphql endpoint doc
{
    factomNodeName
    identityChainID 
    event {
        ... on ProcessListEvent {
            processListEvent {
                ... on NewBlockEvent {
                    newBlockHeight
                }
                ... on NewMinuteEvent {
                    newMinute
                    blockHeight
                }
            }
        }
    }
}
```