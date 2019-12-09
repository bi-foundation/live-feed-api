Directory Block Anchor example
```graphql endpoint doc
{
    factomNodeName
    identityChainID 
    event {
        ... on DirectoryBlockAnchor {
            blockHeight
            btcBlockHash
            btcBlockHeight
            btcConfirmed
            btcTxHash
            btcTxOffset
            directoryBlockHash
            directoryBlockMerkleRoot
            ethereumAnchorRecordEntryHash
            ethereumConfirmed
            timestamp
          }
    }
}
```