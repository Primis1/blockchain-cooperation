# My-own-blockchain

*If you want to read only purpose of functions, pls skip this boring introduction*

In brief. My idea was to finally answer "What the hell is even this blockchain is?". 

`Don't understand something. Build it.`

So, my goofy plan was: 

    1. Read A LOT OF DOCS 
    2. Write simple, but functionally complete blockchain in golang 
    3. Look for exaplme in Geth(go etherium); ask a team-lead; read repo again 
    4. Re-write it all with patterns
    5. Write tests
    6. Write fixes after tests
    7. Be happy with my work (never happen)

My next step is to make a blog page with Next.Js with complete overview for the entire project, but for now let it stay in the way it is

***

### `blockchain.go`
- `InitBlockchain(address, nodeId)`: Create initial blockchain with genesis block
- `ContinueBlockchain(nodeId)`: Restore existing blockchain
- `MineBlock(transaction)`: Create and add new block with transactions
- `VerifyTransaction(t *Transaction)`: Validate transaction integrity
- `FindUniqueTransaction(address)`: Retrieve transactions for specific address
- `GetBestHeightAndLastHash()`: Get current blockchain height and last block hash

### `block.go`
- `CreateBlock(txs, prevHash, height)`: Generate new block with transactions
- `CreateGenesis(coinbase)`: Create initial genesis block
- `HashTransactions()`: Generate Merkle root for block's transactions
- `Serialize()`: Convert block to byte array
- `DeserializeBlock(data)`: Reconstruct block from byte array

### `transaction.go`
- `NewTransaction(wallet, to, amount, UTXO)`: Create new transaction
- `Sign(privateKey, prevTransactions)`: Sign transaction with private key
- `Verify(prevTransactions)`: Validate transaction signatures
- `IsCoinbase()`: Check if transaction is a coinbase (mining reward)
- `CoinbaseTx(to, data)`: Create coinbase transaction for mining rewards

### `unspent.go`
- `Reindex()`: Rebuild UTXO set
- `Update(block)`: Update UTXO set after new block
- `FindSpendableOutputs(pubKeyHash, amount)`: Find unspent outputs for transaction
- `CountUnspentOuts()`: Count total unspent transaction outputs

### `proof.go`
- `NewProof(block)`: Create proof of work for block
- `Run()`: Mine block by finding valid nonce
- `Validate()`: Check if block's proof of work is valid


***

## wallet.go

### `CreateWallets(nodeId string) (*Wallets, error)` 
    Initializes a `Wallets` instance, loads wallet data from a file associated with the given `nodeId`, and returns the instance.

### `AddWallet() string` 
    Creates a new wallet with a unique private/public key pair, generates an address, and adds the wallet to the `Wallets` collection. Returns the generated address.

### `GetWallet(address string) *Wallet` 
    Retrieves a wallet from the `Wallets` collection using its address as the key.

### `GetAllAddresses() []string` 
    Returns a list of all wallet addresses stored in the `Wallets` collection.

### `loadFile(nodeId string) error` 
    Loads wallet data from a file associated with the given `nodeId`. If the file doesn't exist, it returns an error.

### `SaveFile(nodeId string)`  
    Saves the current `Wallets` collection to a file associated with the given `nodeId`.

***

## wallets.go

### `Address() []byte`  
    Generates a wallet address by applying a series of hashing algorithms (SHA-256 and RIPEMD-160) and encoding the result with Base58. Includes a version byte and a checksum.

### `newKeyPair() (ecdsa.PrivateKey, []byte)`  
    Generates a new key pair for a wallet. Uses the P256 elliptic curve to create a private key and derives the public key from it.

### `PublicKey(pubkey []byte) []byte`  
    Generates a public key hash by applying SHA-256 and RIPEMD-160 hashing algorithms to the provided public key.

### `checkSum(payload []byte) []byte`  
    Computes a checksum for a given payload by applying SHA-256 twice and returning the first 4 bytes of the result.

### `ValidateAddress(address string) bool`  
    Validates a wallet address by verifying its checksum and ensuring it matches the hash of the public key.

### `makeWallet() *Wallet`  
    Creates a new `Wallet` by generating a private/public key pair and initializing a `Wallet` instance with these keys.

***

# Resources used in development 
*Beyond good website that i wish i found earlier:*

https://learnmeabitcoin.com/

https://github.com/ethereum/go-ethereum
