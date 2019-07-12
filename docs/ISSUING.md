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
```bash
fat-cli -identity <ROOT CHAIN> -tokenid <STRING TOKEN ID> issue -ecpub <EC Public> -name <STRING TOKEN NAME> -sk1 <SK1> -supply <INTEGER MAX TOKEN SUPPLY> -symbol <STRING TOKEN SYMBOL> -type <"FAT-0" | "FAT-1>
```

The token symbol can be up to 4 letters.
Once this transaction completes - instantaneously on a private chain, about 10 minutes on public testnet, you need to run the same command again. 

The first time you run it is to create the chain for the new token, you should see an output like this:
```
Created Token Chain
Token Chain ID:  d769a40522998f10ed82e8cd96f875d3163742efa07d973d799a507246210cb7
First Entry Hash:  5b11d18d1d17ebdeffc2f96a50dc1e9e24c171df6f9e7aaca46186ff40fdad48
Factom TxID:  ac42be15931ed95b288e13696a0224c4d172fed737b84d6e6398475cea5c74f1
You must wait until the Token Chain is created before issuing the token.
This can take up to 10 minutes.
```

The second time is to create the issuance entry, this will actually issue the tokens. You should see an output like this:
```
Created Issuance Entry
Token Chain ID:  d769a40522998f10ed82e8cd96f875d3163742efa07d973d799a507246210cb7
Issuance Entry Hash:  44e6994c12315244d2920902273bbbfb000eed7a64aa55da30902c280131998f
Factom TxID:  f93fe0e6056f9ca0b5eabceb9af3880a93b6861ddb666a0b3cec21a9d54fff8e
```

You can find these in your factomd control panel (port 8090)

### Distribute tokens

Now that the token have been initialized, you actually need to distribute the tokens to Factom addresses. A `coinbase` transaction will be used to do so. You don't need to distribute the entire token supply in one coinbase transaction, you can submit multiple transactions over time to distribute your token supply.

The coinbase commmand looks like this:
```bash
fat-cli -chainid <token CHAIN ID> <transactFAT0 | transactFAT1> -sk1 <SK1> -coinbase <AMOUNT/IDS TO DISTRIBUTE> -output <FA1:AMOUNT/IDS> ... -ecpub <EC PUBLIC KEY>
```

There can be many outputs. You may want to supply tokens to multiple accounts in the coinbase command, just remember the sum of the output amounts should equal `AMOUNT TO DISTRIBUTE`. In the case of FAT-1 tokens the inputs and outputs must contain the same set of NF token IDs.

An entry credit address needs to be provided so that fatd can pay for the transaction. 

Using the FA Addresses you've created, run this:

```
fat-cli -chainid d769a40522998f10ed82e8cd96f875d3163742efa07d973d799a507246210cb7 transact -sk1 <SK1> -coinbase 1000 -output FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q:500 -output FA3TMQHrCrmLa4F9t442U3Ab3R9sM1gThYMDoygPEVtxrbHtFRtg:500
```

You may see an error like this: `jsonrpc2.Error{Code:-32806, Message:"No Entry Credits", Data:"not configured with entry credits"}`

If you do, simply add the `-ecpub` flag to the command and it should work as expected.

Great! Now that we have 2 FA addresses funded with your token, you're free to transact between other addresses.

### Send tokens

```bash
fat-cli -chainid <CHAIN ID> transact -input <FA ADDRESS:AMOUNT> -output <FA ADDRESS:AMOUNT> -ecpub <EC PUBLIC KEY>
```

An entry credit address needs to be provided so that fatd can pay for the transaction. 

The `input` and `output` flags are used to indicate how the transaction is funded and who the funds should be dispersed to.


## Potential Errors

* No entry credits/ not configured with entry credits.
This error is caused with fat-cli doesn't know how a transaction should be funded. There are two ways to solve this.<br>
1. Pass `-ecpub` with a funded EC address to all fat-cli commands that submit a transaction. <br>
OR <br>
2. Run/restart fat-cli with the flag `-ecpub <EC PUBLIC KEY>` <br>
This will ensure that fat-cli knows how to pay for the transactions.

* Unable to retrieve account balance 
If you want to find out how many tokens your wallet has, you would run the following command: <br>
`fat-cli -chainid <CHAIN ID> balance <FA ADDRESS>` and you might run into the following error (most likely if you are on a private chain) - <br> 
`jsonrpc2.Error{Code:-32800, Message:"Token Not Found", Data:"token may be invalid, or not yet issued or tracked"}`

This happens when fatd starts scanning at a block height *much* higher than where your private chain is most likely at. To solve this, you want to restart fatd with the flag `-startscanheight 0`. fatd will scan through all the blocks until your private chain's max block height and look for all valid FAT transactions. Now you will be able to run the command with no error and see your token wallet balance. A good indicator of this being the case is your `fatd.db` folder will be empty. It *should* have a folder with the same name as you <CHAIN ID>.

*fatd will create fatd.db in the folder from which you run fatd. It is advised to run fatd from the same location each time you do it until a better implementation of this is released.*

*Do not use `-startscanheight 0` the next time you run fatd. You could, but you'd just be wasting time. The next time it runs it'll run from the last blockheight so there's no need to rescan.*
