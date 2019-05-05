# FAT CLI Documentation

The fat-cli allows users to explore and interact with FAT chains.

fat-cli can be used to explore FAT chains to get balances, issuance, and
transaction data. It can also be used to send transactions on existing FAT
chains, and issue new FAT-0 or FAT-1 tokens.

**Chain ID Settings**

Most sub-commands need to be scoped to a specific FAT chain, identified by a `--chainid`. Alternatively, this can be specified by using both the `--tokenid` and `--identity`, which together determine the chain ID.

**API Settings**

fat-cli makes use of the fatd, factomd, and factom-walletd JSON-RPC 2.0 APIs for various operations. Trust in these API endpoints is imperative to secure operation.

The `--fatd` API is used to explore issuance, transactions, and balances for existing FAT chains.

The `--factomd` API is used to submit entries directly to the Factom
blockchain, as well as for checking EC balances, chain existence, and
identity keys.

The `--walletd` API is used to access private keys for FA and EC
addresses. To avoid use of factom-walletd, use private Fs or Es keys
directly on the CLI instead.

If `--debug` is set, all fatd and factomd API calls will be printed to
stdout. API calls to factom-walletd are omitted to avoid leaking
private key data.

**Offline Mode**

For increased security to protect private keys, it is possible to run fat-cli such that it makes no network calls when generating Factom entries for FAT transactions or token issuance.

Use `--curl` to skip submitting the entry directly to Factom, and instead print out the curl commands for committing and revealing the entry. These curl commands contain the encoded signed data and may be safely copied to, and run from, a computer with access to factomd.

Use `--force` to skip all sanity checks that involve API calls out factomd or fatd. As a result, this may result in generating a Factom Entry that is invalid for Factom or FAT, but may still use up Entry Credits to submit.

Use private keys for `--ecadr` and --input directly to avoid any network calls to factom-walletd.

**Entry Credits**
Making FAT transactions or issuing new FAT tokens requires creating entries on the Factom blockchain. Creating Factom entries costs Entry Credits. Entry Credits have a relatively fixed price of about $0.001 USD. Entry Credits can be obtained by burning Factoids which can be done using the official factom-cli.  FAT transactions normally cost 1 EC. The full FAT Token Issuance process normally costs 12 EC.

**CLI Completion**
After installing fat-cli in some permanent location in your PATH. Use --installcompletion to install CLI completion for Bash, Zsh, or Fish. This simply adds a single line to your `~/.bash_profile` (or shell equivalent), which can be removed with --uninstallcompletion. You must re-open your shell before completion changes take effect.

No other programs or files need to be installed because fat-cli is also its own completion program. If fat-cli is envoked by the completion system, it returns completions for the currently typed arguments.

If the `--fatd` endpoint is available, Token Chain IDs can be completed based on the chains that fatd is tracking.

If the `--walletd` endpoint is available, then all FA and EC addresses can be completed based on the addresses saved by factom-walletd.

Since both of these completion flags require successful API calls, any required API related flags must already be supplied before completion for Token Chain IDs, FA or EC addresses can succeed. Otherwise, if the default settings are incorrect, generating completion suggestions will fail silently. Note that --timeout is ignored as a very short timeout is always used to avoid noticeable blocking when generating completion suggestions.



# Flags

## **General**

### `--debug`

Print fatd and factomd API calls

### `--verbose`

Print verbose details about sanity check and other operations

### `--help`

Get help with using the CLI. Can follow any command or subcommand to get detailed help.



## **Network & Auth**

### `--fatd`

Fatd URI

scheme://host:port for fatd (default `localhost:8078`)

### `--fatdpass`

 Basic HTTP Auth Password for fatd

### `--fatduser`

Basic HTTP Auth User for fatd

### `--factomd`

Factomd URI

scheme://host:port for factomd (default `localhost:8088`)

### `--factomdpass`

Basic HTTP Auth Password for factomd

### `--factomduser`

Basic HTTP Auth User for factomd

### `--walletd`

Factom Walletd URI

scheme://host:port for factom-walletd (default `localhost:8089`)

### `--walletduser`

Basic HTTP Auth User for factom-walletd

### `--walletdpass`

Basic HTTP Auth Password for factom-walletd

### `--timeout`

 Timeout for all API requests (i.e. `10s`, `1m`) (default `3s`)



## **Tokens**

### `--chainid`

The 32 Byte Factom Chain ID of the token to get data for. The token chain ID can be calculated from `--identity` & `--tokenid`. Either `--chainid` OR `--identity` & `--tokenid` should be supplied

### `--identity`

 Token Issuer Identity Chain ID of a FAT token. 

### `--tokenid`

The Token ID string of a FAT chain.



# Commands

## `get`

Retrieve data about FAT tokens or a specific FAT token

**Usage**

```
 fat-cli get [subcommand] [flags]
```



### Subcommands

#### `chains`

Get information about tokens and token chains. Print a list including the Chain ID, Issuer identity chain ID, and token ID of each token currently tracked by the fat daemon.

**Usage**

```
fat-cli get chains <chainid>
```

 If the optional `<chainid>` argument is supplied the info for that specific chain will be returned including statistics.



#### `balance`

Get the balance of a Factoid address for a token

**Usage**

```
fat-cli get balance --chainid <chainid> <FA Address>
```



#### `transactions`

Get transaction history and specific transactions belonging to a specific FAT token.

**Usage**

```
fat-cli get transactions --chainid <chain-id> [--starttx <tx-hash>]
        [--page <page>] [--limit <limit>] [--order <"asc" | "desc">]
        [--address <FA> [--address <FA>]... [--to] [--from]]
        [--nftokenid <nf-token-id>]
```

- `--address` - Add to the set of addresses to lookup txs for
- `--from` - Request only txs FROM the given `--address` set
- `--to` - Request only txs TO the given --address set
- `--limit` - Limit of returned txs (default `10`)
- `--nftokenid` - Request only txs involving this NF Token ID
- `--order` -  Order of returned txs (`asc`|`desc`, default `asc`)
- `--page` - Page of returned txs (default `1`)
- `--starttx` - Entryhash of tx to start indexing from



## `issue`

Issue a new FAT-0 or FAT-1 token chain.

Issuing a new FAT token chain is a two step process. First the Token Chain must
be created on the Factom Blockchain. Both `--tokenid` and `--identity` are
required. Use of `--chainid` is not allowed for this step as the chain ID is being calculated and the chain established in these steps. Up to 10 minutes is required in between steps to allow for the block to finish and the chain creation to confirm.

**Step 1 Usage** (Chain Creation)

```
fat-cli issue chain --factomd <factomd> --ecadr <EC | Es> --identity <issuer-identity-chain-id> --tokenid <token-id> [--curl] [--force]
```

**Step 2 Usage** (Token Issuance)

```
fat-cli issue token --factomd <factomd> --ecadr <EC | Es> --chainid <token-chain-id> --sk1 <sk1-key> --type <FAT-0|FAT-1> [--supply <max-supply>] [--curl] [--force]
```

- `--ecadr` - EC or Es address to pay for the chain creation and token issuance entries
- `--sk1` - The SK1 Private identity key belonging to `--identity`
- `--type` - The token type to issue. Either `FAT-0` or `FAT-1`
- `--supply` - The maximum supply of the token can achieve ("outstanding shares")
- `--curl` -  Do not submit the Factom entry; print curl commands instead!
- `--force` - Skip sanity checks for balances, chain status, and sk1 key



## `transact`

Send or distribute FAT-0 or FAT-1 tokens.

Submitting a FAT transaction involves submitting a signed transaction entry to
the given FAT Token Chain

**Usage**

```
fat-cli transact <fat0 | fat1> --input <input> --output <output> [--metadata <metadata>] [--sk1 <sk1>] [--ecadr <EC | Es>] [--curl] [--force]
```

- `--input` - An input to the transaction. May be specified multiple times. For example a FAT-0 tx input could look like `FA3SjebEevRe964p4tQ6eieEvzi7puv9JWF3S3Wgw2v3WGKueL3R:150`, a FAT-1 tx input could look like `FA3SjebEevRe964p4tQ6eieEvzi7puv9JWF3S3Wgw2v3WGKueL3R:[1,2,5-100]`. It is allowed to use private Factoid keys instead of public ones if `--walletd` is not specified or available.
- `--output` -  An output to the transaction. May be specified multiple times. Follows the same form as `--input` except no private Factoid keys are permitted.
- `--metadata` - JSON compliant metadata to attach to the transaction
- `--sk1` - The SK1 Private identity key belonging to the issuer of the token. Required for coinbase transactions.
- `--ecadr` - EC or Es address to pay for the chain creation and token issuance entries, if `--factomd` is specified.
- `--curl` -  Do not submit the Factom entry; print curl commands instead!
- `--force` - Skip sanity checks for balances, chain status, and sk1 key