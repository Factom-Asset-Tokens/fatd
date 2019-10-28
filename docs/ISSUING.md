## Issuing a FAT Token

Using FAT you can issue your own token on the Factom Blockchain. There are two types of tokens you can create
* FAT-0: fungible tokens (like ERC-20)
* FAT-1: non-fungible tokens (like ERC-721)

The following steps show how you can:
* Initialize a token
* Distribute tokens to an address
* Transact your new token between Factom Addresses

### Prerequisites

Before you can create a token with FAT, you need to create a Factom Server Identity. The following link explains how to do that: https://docs.factomprotocol.org/authority-node-operators/ano-guides-and-tutorials/generating-your-server-identity

After running the `important.sh` script you should see the following:

Identity Private Keys:
Level 1: <SK1>
Level 2: <SK2>
Level 3: <SK3>
Level 4: <SK4>

Root Chain          : <ROOT CHAIN>
Management Chain    : <MANAGEMENT CHAIN>

While creating the Identity you would have created an EC Address, save the public and private key. You see the private key once you export the address.

EC Address      : <EC PUBLIC>
EC Address PK   : <EC PRIVATE KEY>
FA1             : <FA1>
FA2             : <FA2>

Now that you have all this done, you're ready to initialize your token.

### Initialize a token

Run the following command:
```
$ fat-cli issue -v --sk1 <REDACTED> \
        --ecadr EC3emTZegtoGuPz3MRA4uC8SU6Up52abqQUEKqW44TppGGAc4Vrq \
        --tokenid "SPARTA" \
        --identity 8888888de45074fb3505cfdc942f80f4c9ef1ddd5c4633cd21a940288ffc89f3 \
        --type FAT-0 \
        --symbol "SPT" \
        --precision 10 \
        --supply 3000000000000 \
        --metadata '{"fight":true}'
Fetching secret address... EC3emTZegtoGuPz3MRA4uC8SU6Up52abqQUEKqW44TppGGAc4Vrq
Token Chain ID: 56bba15293b1f24849a7c3205a30db5981022776bea711d217437a642e9e080c
Preparing Chain Creation Entry...
Preparing and signing Token Initialization Entry...
Checking chain existence...
Checking token chain status...
Fetching Identity Chain...
Verifying SK1 Key...
Checking EC balance...
New chain creation cost: 11 EC
Token Initialization Entry cost: 1 EC

Submitting the Chain Creation Entry to the Factom blockchain...
Chain Creation Entry Submitted
Chain ID:     56bba15293b1f24849a7c3205a30db5981022776bea711d217437a642e9e080c
Entry Hash:   ca6a22133e5ad6c96afd26567b52cc8437f7507f11833b5607cf532487fa6376
Factom Tx ID: eb9db70e8e854ff6ad7abc51d42a589c5c922345caf96508c0e2f80dd44f6e02

Submitting the Token Initialization Entry to the Factom blockchain...
Token Initialization Entry Submitted
Entry Hash:   271ade6467d44937406fe934e7d68cc44a18b61778cdd4a478b4353db4caaa5c
Factom Tx ID: 6df52ff302c6a03734ac94c0cacacabcbc9461ade1ff016e9d9e5866a11d0abb
```

The token symbol can be up to 4 letters. Once this transaction completes -
instantaneously on a private chain, about 10 minutes on public testnet, you
need to run the same command again.

The command will create the chain if it does not exist, and submit the issuance
entry. See `fat-cli help issue` for more details.

After about 10 minutes, the chain will be tracked by `fatd`. Information about
the chain can then be queried.

```
$ fat-cli get chains 56bba15293b1f24849a7c3205a30db5981022776bea711d217437a642e9e080c

Chain ID:        56bba15293b1f24849a7c3205a30db5981022776bea711d217437a642e9e080c
Issuer Identity: 8888888de45074fb3505cfdc942f80f4c9ef1ddd5c4633cd21a940288ffc89f3
Issuance Entry:  7f8ab15b40aaece738d6afa1053ba070a0daa28052be28e777a2ee6258bec569
Token ID:  SPARTA
Type:      FAT-0
Symbol:    "SPT"
Precision: 10
Supply:    3000000000000
Circulating Supply:      0
Burned:                  0
Number of Transactions:  0
Issuance Timestamp: 2019-10-24 17:00:00 -0800 AKDT
```

### Distribute tokens

Now that the token has been initialized, you need to initially distribute some
tokens to Factom addresses using a `coinbase` transaction. You don't need to
distribute the entire token supply in one coinbase transaction, you can submit
multiple coinbase transactions over time to distribute your token supply.

Coinbase transactions are signed with the same SK1 key that the Issuance entry
uses. The following command distributes the entire supply of the newly created
token to two addresses.

```
$ fat-cli transact fat0 \
    --chainid 56bba15293b1f24849a7c3205a30db5981022776bea711d217437a642e9e080c \
    --sk1 <REDACTED> \
    --output FA2kEkNgQ5RMNx5Y14HRQa4X8czeZqg74AJykR8f3jx4Cbk26gcM:1500000000000 \
    --output FA2mnS2QfXNQjdq6jJKxUxDnwPXzLpxivYDrYMtLgmRDbrxZztY5:1500000000000 \
    --ecadr EC3emTZegtoGuPz3MRA4uC8SU6Up52abqQUEKqW44TppGGAc4Vrq

FAT-0 Transaction Entry Created
Entry Hash:   dd2afeeb99232893caa52264cc80f094f78ebfed47766fd77ea99e8987e54a0f
Chain ID:     56bba15293b1f24849a7c3205a30db5981022776bea711d217437a642e9e080c
Factom Tx ID: 693060a967a1ceee0353e35ebf66dd30f0ec4fd76a65eac228a5f2a10aeb448b
```

See `fat-cli help transact` for more details.

### Send tokens

```
$ fat-cli transact fat0 \
    --chainid 56bba15293b1f24849a7c3205a30db5981022776bea711d217437a642e9e080c \
    --input FA2kEkNgQ5RMNx5Y14HRQa4X8czeZqg74AJykR8f3jx4Cbk26gcM:5
    --output FA2mnS2QfXNQjdq6jJKxUxDnwPXzLpxivYDrYMtLgmRDbrxZztY5:5 \
    --ecadr EC3emTZegtoGuPz3MRA4uC8SU6Up52abqQUEKqW44TppGGAc4Vrq

FAT-0 Transaction Entry Created
Chain ID:     56bba15293b1f24849a7c3205a30db5981022776bea711d217437a642e9e080c
Entry Hash:   5e170b84c076b71363a13075b40765617378db4960a7efc851658c97291010e4
Factom Tx ID: 404f7a19c515f42d057cc0f8981651750069e4c612cb5a3963ef5a5a18c94844
```

A normal transaction can have multiple inputs and multiple outputs. The private
key for any public address used as an `--input` will be fetched from
factom-walletd.

The sum of the inputs must equal the sum of the outputs.


## Potential Errors

#### No entry credits/ not configured with entry credits.

This error is caused with fat-cli doesn't know how a transaction should be
funded. To solve this, pass `-ecpub` with a funded EC address to all fat-cli
commands that submit a transaction.

#### Unable to retrieve account balance
If you want to find out how many tokens your wallet has, you would run the
following command:
```
fat-cli get balance --chainid <CHAIN ID> <FA ADDRESS>
```
and you might run into the following error (most likely if you are on a private
chain) - `jsonrpc2.Error{Code:-32800, Message:"Token Not Found", Data:"token
may be invalid, or not yet issued or tracked"}`

This happens when fatd starts scanning at a block height *much* higher than
where your private chain is most likely at. To solve this, you want to restart
fatd with the flag `-startscanheight 0`. fatd will scan through all the blocks
until your private chain's max block height and look for all valid FAT
transactions. Now you will be able to run the command with no error and see
your token wallet balance. A good indicator of this being the case is your
`fatd.db` folder will be empty. It *should* have a folder with the same name as
your <CHAIN ID>.

*By default, fatd will create its databases in your home folder: `~/.fatd/`*

*Do not use `-startscanheight 0` the next time you run fatd. You could, but
you'd just be wasting time. The next time it runs it'll run from the last
blockheight so there's no need to rescan.*
