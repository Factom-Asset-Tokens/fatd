# FAT CLI Documentation

The fat-cli allows users to explore and interact with FAT chains.

fat-cli can be used to explore FAT chains to get balances, issuance, and
transaction data. It can also be used to send transactions on existing FAT
chains, and issue new FAT-0 or FAT-1 tokens.

**Chain ID Settings**

Most sub-commands need to be scoped to a specific FAT chain, identified by a
`--chainid`. Alternatively, this can be specified by using both the `--tokenid`
and `--identity`, which together determine the chain ID.

**API Settings**

fat-cli makes use of the fatd, factomd, and factom-walletd JSON-RPC 2.0 APIs
for various operations. Trust in these API endpoints is imperative to secure
operation.

The `--fatd` API is used to explore issuance, transactions, and balances for
existing FAT chains.

The `--factomd` API is used to submit entries directly to the Factom
blockchain, as well as for checking EC balances, chain existence, and identity
keys.

The `--walletd` API is used to access private keys for FA and EC addresses. To
avoid use of factom-walletd, use private Fs or Es keys directly on the CLI
instead.

If `--debug` is set, all fatd and factomd API calls will be printed to stdout.
API calls to factom-walletd are omitted to avoid leaking private key data.

**Offline Mode**

For increased security to protect private keys, it is possible to run fat-cli
such that it makes no network calls when generating Factom entries for FAT
transactions or token issuance.

Use `--curl` to skip submitting the entry directly to Factom, and instead print
out the curl commands for committing and revealing the entry. These curl
commands contain the encoded signed data and may be safely copied to, and run
from, a computer with access to factomd.

Use `--force` to skip all sanity checks that involve API calls out factomd or
fatd. As a result, this may result in generating a Factom Entry that is invalid
for Factom or FAT, but may still use up Entry Credits to submit.

Use private keys for `--ecadr` and --input directly to avoid any network calls
to factom-walletd.

**Entry Credits**
Making FAT transactions or issuing new FAT tokens requires creating entries on
the Factom blockchain. Creating Factom entries costs Entry Credits. Entry
Credits have a relatively fixed price of about $0.001 USD. Entry Credits can be
obtained by burning Factoids which can be done using the official factom-cli.
FAT transactions normally cost 1 EC. The full FAT Token Issuance process
normally costs 12 EC.

### CLI Completion
After installing fat-cli in some permanent location in your PATH. Use
--installcompletion to install CLI completion for Bash, Zsh, or Fish. This
simply adds a single line to your `~/.bash_profile` (or shell equivalent),
which can be removed with --uninstallcompletion. You must re-open your shell
before completion changes take effect.

No other programs or files need to be installed because fat-cli is also its own
completion program. If fat-cli is envoked by the completion system, it returns
completions for the currently typed arguments.

If the `--fatd` endpoint is available, Token Chain IDs can be completed based
on the chains that fatd is tracking.

If the `--walletd` endpoint is available, then all FA and EC addresses can be
completed based on the addresses saved by factom-walletd.

Since both of these completion flags require successful API calls, any required
API related flags must already be supplied before completion for Token Chain
IDs, FA or EC addresses can succeed. Otherwise, if the default settings are
incorrect, generating completion suggestions will fail silently. Note that
--timeout is ignored as a very short timeout is always used to avoid noticeable
blocking when generating completion suggestions.


# Flags

## General

### `--debug`

Print fatd and factomd API calls

### `--verbose`

Print verbose details about sanity check and other operations

### `--help`

Get help with using the CLI. Can follow any command or subcommand to get
detailed help.



## Network & Auth

### `--fatd`

Fatd URL

scheme://host:port for fatd (default `localhost:8078`)

### `--fatdpass`

Basic HTTP Auth Password for fatd

### `--fatduser`

Basic HTTP Auth User for fatd

### `--factomd`

Factomd URL

scheme://host:port for factomd (default `localhost:8088`)

### `--factomdpass`

Basic HTTP Auth Password for factomd

### `--factomduser`

Basic HTTP Auth User for factomd

### `--walletd`

Factom Walletd URL

scheme://host:port for factom-walletd (default `localhost:8089`)

### `--walletduser`

Basic HTTP Auth User for factom-walletd

### `--walletdpass`

Basic HTTP Auth Password for factom-walletd

### `--timeout`

Timeout for all API requests (i.e. `10s`, `1m`) (default `3s`)



## Tokens

### `--chainid`

The 32 Byte Factom Chain ID of the token to get data for. The token chain ID
can be calculated from `--identity` & `--tokenid`. Either `--chainid` OR
`--identity` & `--tokenid` should be supplied

### `--identity`

Token Issuer Identity Chain ID of a FAT token.

### `--tokenid`

The Token ID string of a FAT chain.



# Commands

## `get`

Retrieve data about FAT tokens or a specific FAT token

```
 fat-cli get [subcommand] [flags]
```

### Subcommands

#### `chains`

Get information about tokens and token chains. Print a list including the Chain
ID, Issuer identity chain ID, and token ID of each token currently tracked by
the fat daemon.

```
fat-cli get chains <chainid>
```

If the optional `<chainid>` argument is supplied the info for that specific
chain will be returned including statistics.


#### `balance`

Get the balance of a Factoid address for a token

```
fat-cli get balance --chainid <chainid> <FA Address>
```

#### `transactions`

Get transaction history and specific transactions belonging to a specific FAT
token.


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

Issuing a new FAT token chain is a two step process but only requires a single command.

First, the Token Chain must be created with the correct Name IDs on the Factom
Blockchain. So both --tokenid and --identity are required and use of --chainid
is not allowed for this step. If the Chain Creation Entry has already been
submitted then this step is skipped over.

Second, the Token Initialization Entry must be added to the Token Chain. The
Token Initialization Entry must be signed by the SK1 key corresponding to the
ID1 key declared in the --identity chain. Both --type and --supply are
required. The --supply must be positive or -1 for an unlimited supply of
tokens.

Note that publishing a Token Initialization Entry is an immutable operation.
The protocol does not permit altering the Token Initialization Entry in any
way.

Sanity Checks
        Prior to composing the Chain Creation or Token Initialization Entry, a
        number of calls to fatd and factomd are made to ensure that the token
        can be issued. These checks are skipped if --force is used.

        - Skip Chain Creation Entry if already submitted.
        - The token has not already been issued.
        - The --identity chain exists.
        - The --sk1 key corresponds to the --identity's id1 key.
        - The --ecadr has enough ECs to pay for all entries.

Identity Chain
        FAT token chains may only be issued by an entity controlling the
        sk1/id1 key established by the Identity Chain pointed to by the FAT
        token chain. An Identity Chain and the associated keys can be created
        using the factom-identity-cli.

        https://github.com/PaulBernier/factom-identity-cli

Entry Credits
        Creating entries on the Factom blockchain costs Entry Credits. The full
        Token Issuance process normally costs 12 ECs. You must specify a funded
        Entry Credit address with --ecadr, which may be either a private Es
        address, or a pubilc EC address that can be fetched from
        factom-walletd.

**Usage**

```
  fat-cli issue --ecadr <EC | Es> --sk1 <sk1-key>
        --identity <issuer-identity-chain-id> --tokenid <token-id>
        --type <"FAT-0" | "FAT-1"> --supply <supply> [--metadata <JSON>] [flags]
```

```
Flags:
      --curl                       Do not submit Factom entry; print curl commands
  -e, --ecadr <EC | Es>            EC or Es address to pay for entries
      --force                      Skip sanity checks for balances, chain status, and sk1 key
  -h, --help                       help for issue
  -m, --metadata JSON              JSON metadata to include in tx
      --sk1 sk1                    Secret Identity Key 1 to sign entry
      --supply int                 Max Token supply, use -1 for unlimited
      --symbol string              Optional abbreviated token symbol
      --type <"FAT-0" | "FAT-1">   Token standard to use
```

**Example Commands**

Initialize a FAT-0 token called "test" with a maximum supply of 100,000 units:

```
fat-cli issue --ecadr EC3cQ1QnsE5rKWR1B5mzVHdTkAReK5kJwaQn5meXzU9wANyk7Aej --sk1 sk1... --identity 888888a37cbf303c0bfc8d0cc7e77885c42000b757bd4d9e659de994477a0904 --tokenid test --type "FAT-0" --supply 100000
```

Initialize a FAT-1 token called "test-nft" with an unlimited supply:
```
fat-cli issue --ecadr EC3cQ1QnsE5rKWR1B5mzVHdTkAReK5kJwaQn5meXzU9wANyk7Aej --sk1 sk1... --identity 888888a37cbf303c0bfc8d0cc7e77885c42000b757bd4d9e659de994477a0904 --tokenid test-nft --type "FAT-1" --supply -1
```

## `transact`

Send or distribute FAT-0 or FAT-1 tokens.

Submitting a FAT transaction involves submitting a signed transaction entry to
the given FAT Token Chain

**Usage**

```
fat-cli transact <fat0 | fat1> --input <input> --output <output> [--metadata <metadata>] [--sk1 <sk1>] [--ecadr <EC | Es>] [--curl] [--force]
```

- `--input` - An input to the transaction. May be specified multiple times. For
  example a FAT-0 tx input could look like
`FA3SjebEevRe964p4tQ6eieEvzi7puv9JWF3S3Wgw2v3WGKueL3R:150`, a FAT-1 tx input
could look like
`FA3SjebEevRe964p4tQ6eieEvzi7puv9JWF3S3Wgw2v3WGKueL3R:[1,2,5-100]`. It is
allowed to use private Factoid keys instead of public ones if `--walletd` is
not specified or available.
- `--output` -  An output to the transaction. May be specified multiple times.
  Follows the same form as `--input` except no private Factoid keys are
permitted.
- `--metadata` - JSON compliant metadata to attach to the transaction
- `--sk1` - The SK1 Private identity key belonging to the issuer of the token.
  Required for coinbase transactions.
- `--ecadr` - EC or Es address to pay for the chain creation and token issuance
  entries, if `--factomd` is specified.
- `--curl` -  Do not submit the Factom entry; print curl commands instead!
- `--force` - Skip sanity checks for balances, chain status, and sk1 key

**Example Commands**

FAT-0 Coinbase Transaction

Use `--sk1` and `--output` for coinbase transactions; no `--input` is specified as the coinbase input address is always `FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC`, the public address that corresponds to a private key of all zeroes. This creates 100 new tokens and sends them to the the `FA2gCm...` address.

```
fat-cli transact fat0 --output FA2gCmih3PaSYRVMt1jLkdG4Xpo2koebUpQ6FpRRnqw5FfTSN2vW:100 --sk1 sk1... --ecadr EC3cQ1QnsE5rKWR1B5mzVHdTkAReK5kJwaQn5meXzU9wANyk7Aej
```

FAT-1 Transaction

This moves the token with an id of 10 from `FA2gCm...` to `FA3j68...`.

```
fat-cli transact fat1 --input FA2gCmih3PaSYRVMt1jLkdG4Xpo2koebUpQ6FpRRnqw5FfTSN2vW:[10] --output FA3j68XNwKwvHXV2TKndxPpyCK3KrWTDyyfxzi8LwuM5XRuEmhy6:[10] --ecadr EC3cQ1QnsE5rKWR1B5mzVHdTkAReK5kJwaQn5meXzU9wANyk7Aej
```