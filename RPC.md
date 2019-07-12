# FAT Daemon RPC Documentation

This document defines the Remote Procedure Call API for fatd. The RPC encompasses methods to read, transact, and issue tokens on FAT.

The API will follow [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification).

It's important to note that this RPC specification will cover responses for many token types(FAT-0, FAT-1, and so on). As a result, the example JSON responses in this spec are solely for a generic FAT-0 token. JSON out will vary depending on the token type returned.

## API Version

This standard covers RPC API version `v1`



# Token Methods

**All Token Methods require the following parameters at minimum:**

| Name       | Type   | Description          | Validation                                                   | Required |
| ---------- | ------ | -------------------- | ------------------------------------------------------------ | -------- |
| `chainid`  | string | The Token Chain ID   | Must resolve to a valid FAT token chain. Either both `tokenid` and `issuerid` should be specified, or only `chainid` should be specified. | N*       |
| `tokenid`  | string | Token ID             |                                                              | N*       |
| `issuerid` | string | Issuer Root Chain ID | Must be a valid issuer root chain ID. When combined with `tokenid` as per FATIP-100, must resolve to a valid token chain. | N*       |

\* = Either `tokenid` + `issuerid`, or just `chainid`. One of the two options must be selected.



### `get-issuance` :

Get the issuance entry for a token.

#### Parameters:

| Name | Type | Description | Validation | Required |
| ---- | ---- | ----------- | ---------- | -------- |
|      |      |             |            |          |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "chainid": "0cccd100a1801c0cf4aa2104b15dec94fe6f45d0f3347b016ed20d81059494df",
    "tokenid": "test",
    "issuerid": "888888ab72e748840d82c39213c969a11ca6cb026f1d3da39fd82b95b3c1fced",
    "entryhash": "fc0f57ea3a4dc5b8ffc1a9c051f4b6ae0cd7137f9110b98e3c3eb08f132a5e18",
    "timestamp": 1550612940,
    "issuance": {
      "type": "FAT-0",
      "supply": -1,
      "symbol": "T0"
    }
  },
  "id": 6806
}
```

<br/>

### `get-transaction` :

Get a valid FAT transaction for a token

#### Parameters:

| Name        | Type   | Description                                     | Validation                                                   | Required |
| ----------- | ------ | ----------------------------------------------- | ------------------------------------------------------------ | -------- |
| `entryhash` | string | The entry hash of the FAT transaction on Factom | Entry hash must exist as a valid transaction in the FAT database | Y        |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "entryhash": "68f3ca3a8c9f7a0cb32dc9717347cb179b63096e051a60ce8be9c292d29795af",
    "timestamp": 1550696040,
    "data": {
      "inputs": {
        "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC": 10
      },
      "outputs": {
        "FA3aECpw3gEZ7CMQvRNxEtKBGKAos3922oqYLcHQ9NqXHudC6YBM": 10
      }
    }
  },
  "id": 7850
}
```

<br/>

### `get-transactions` :

Get time ordered valid FAT transactions for a token, or token address, non-fungible token ID, or a combination.

- Transactions returned are ordered starting from newest(0th index) to oldest(last index)

#### Parameters:

| Name        | Type   | Description                                                  | Validation                                                   | Required |
| ----------- | ------ | ------------------------------------------------------------ | ------------------------------------------------------------ | -------- |
| `nftokenid` | string | The ID of the non-fungible token to get transactions for.    | The token resolved from `token-id` and`issuer-id` must be a non-fungible token type. | N        |
| `addresses` | array  | Return transactions that include these Factoid addresses in the inputs or outputs | Must all be valid Factoid addresses                          | N        |
| `tofrom`    | string | Return transactions that include this Factoid address in the inputs or outputs | Must be a valid Factoid address                              | N        |
| `entryhash` | string | The tx entryhash to take as the starting point for the page page (inclusive of tx `entryhash`) | Must be a valid FAT tx in the result set determined by the above parameters. | N        |
| `page`      | number | The starting index of the page, inclusive.                   | Integer >= 0. Defaults to 0                                  | N        |
| `limit`     | number | The page size of transactions returned.                      | Integer > 0. Defaults to 25                                  | N        |
| `order`     | string | The time order to return results in. Default `"asc"`         | Either `"asc"` or `"desc"`.                                  | N        |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": [
    {
      "entryhash": "4a6976db7d81fda5067eb2eae1ddf9d5d7de63edb5f8a02b390afeda66dee9bc",
      "timestamp": 1550695620,
      "data": {
        "inputs": {
          "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC": 10
        },
        "outputs": {
          "FA3aECpw3gEZ7CMQvRNxEtKBGKAos3922oqYLcHQ9NqXHudC6YBM": 10
        }
      }
    },
    {
      "entryhash": "894bc5f5fdf7e93b2e77902fe79e460081e9844960878f9256201f7aec3f4b1e",
      "timestamp": 1550695860,
      "data": {
        "inputs": {
          "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC": 10
        },
        "outputs": {
          "FA3aECpw3gEZ7CMQvRNxEtKBGKAos3922oqYLcHQ9NqXHudC6YBM": 10
        }
      }
    }
  ]
}
  
```

<br/>

### `get-balance` :

Get the balance of an address for a token

#### Parameters:

| Name      | Type   | Description                               | Validation                      | Required |
| --------- | ------ | ----------------------------------------- | ------------------------------- | -------- |
| `address` | string | The Factoid address to get the balance of | Must be a valid Factoid address | Y        |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": 891,
  "id": 2007
}
```

<br/>



### `get-nf-balance` :

Get the tokens belonging to an address on a non-fungible token

#### Parameters:

| Name      | Type   | Description                               | Validation                      | Required |
| --------- | ------ | ----------------------------------------- | ------------------------------- | -------- |
| `address` | string | The Factoid address to get the balance of | Must be a valid Factoid address | Y        |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": [
    {
      "min": 12,
      "max": 141
    },
    {
      "min": 143,
      "max": 162
    }
  ]
}
```

<br/>

### `get-stats` :

Get overall statistics for a token

#### Parameters:

| Name | Type | Description | Validation | Required |
| ---- | ---- | ----------- | ---------- | -------- |
|      |      |             |            |          |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "chainid": "1e5037be95e108c34220d724763444098528e88d08ec30bc15204c98525c3f7d",
    "tokenid": "test-nft",
    "issuerid": "888888a37cbf303c0bfc8d0cc7e77885c42000b757bd4d9e659de994477a0904",
    "Issuance": {
      "type": "FAT-1",
      "supply": -1
      },
    "circulating": 1,
    "burned": 0,
    "transactions": 1,
    "issuancets": 1557873300,
    "lasttxts": 1557880560
  },
  "id": 1
}
```



### `get-nf-token` :

Get a non fungible token by ID. The token belong to non fungible token class.

#### Parameters:

| Name        | Type   | Description                             | Validation                                           | Required |
| ----------- | ------ | --------------------------------------- | ---------------------------------------------------- | -------- |
| `nftokenid` | number | The unique ID of the non fungible token | Must be a valid issued non fungible token ID integer | Y        |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "id": 12,
    "owner": "FA3aECpw3gEZ7CMQvRNxEtKBGKAos3922oqYLcHQ9NqXHudC6YBM"
  },
  "id": 4739
}
```

### `get-nf-tokens` :

List all issued non fungible tokens in circulation

#### Parameters:

| Name    | Type   | Description                                          | Validation                  | Required |
| ------- | ------ | ---------------------------------------------------- | --------------------------- | -------- |
| `page`  | number | The starting index of the page, inclusive.           | Integer >= 0. Defaults to 0 | N        |
| `limit` | number | The page size of transactions returned.              | Integer > 0. Defaults to 25 | N        |
| `order` | string | The time order to return results in. Default `"asc"` | Either `"asc"` or `"desc"`. | N        |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": [
    {
      "id": 12,
      "owner": "FA3aECpw3gEZ7CMQvRNxEtKBGKAos3922oqYLcHQ9NqXHudC6YBM"
    },
    {
      "id": 13,
      "owner": "FA3aECpw3gEZ7CMQvRNxEtKBGKAos3922oqYLcHQ9NqXHudC6YBM"
    }
  ]
}
```



### `send-transaction`:

Send A FAT transaction to a token

#### Parameters:

| Name      | Type   | Description                                           | Validation                                                   | Required |
| --------- | ------ | ----------------------------------------------------- | ------------------------------------------------------------ | -------- |
| `extids`  | array  | The hex encoded extids of the signedtransaction entry | Must conform to all transaction validation criteria of the destination token's spec | Y        |
| `content` | string | The transactions hex encoded content                  | Must conform to all transaction validation criteria of the destination token's spec | Y        |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "chainid": "962a18328c83f370113ff212bae21aaf34e5252bc33d59c9db3df2a6bfda966f",
    "txid": "7222ca8f594f7476edc70d7cf7c89c4714239e25626ad578b67de51562288cf9",
    "entryhash": "06fe00477fa198bb221fd0e033a61bb09b2b981529260f516fc5e9bf81ab7a8f"
  },
  "id": 3680
}
```





# Daemon Methods

### `get-daemon-tokens`:

Get the list of FAT tokens the daemon is currently tracking

#### Parameters:

| Name | Type | Description | Validation | Required |
| ---- | ---- | ----------- | ---------- | -------- |
|      |      |             |            |          |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": [
    {
      "chainid": "0cccd100a1801c0cf4aa2104b15dec94fe6f45d0f3347b016ed20d81059494df",
      "tokenid": "test",
      "issuerid": "888888ab72e748840d82c39213c969a11ca6cb026f1d3da39fd82b95b3c1fced"
    },
    {
      "chainid": "962a18328c83f370113ff212bae21aaf34e5252bc33d59c9db3df2a6bfda966f",
      "tokenid": "testnf",
      "issuerid": "888888ab72e748840d82c39213c969a11ca6cb026f1d3da39fd82b95b3c1fced"
    }
  ],
  "id": 8158
}
```



### `get-daemon-properties`:

Get basic properties about the fat daemon

#### Parameters:

| Name | Type | Description | Validation | Required |
| ---- | ---- | ----------- | ---------- | -------- |
|      |      |             |            |          |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "fatdversion": "r162.3d7f272",
    "apiversion": "0"
  },
  "id": 6831
}
```



### `get-sync-status`:

Get the Factom block height sync status of the daemon

#### Parameters:

| Name | Type | Description | Validation | Required |
| ---- | ---- | ----------- | ---------- | -------- |
|      |      |             |            |          |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "syncheight": 70990,
    "factomheight": 70990
  },
  "id": 6482
}
```



### `get-balances`:

Get the numeric balance count for all tracked tokens of a public Factoid address. The returned object has keys representing token chain IDs and values represinting the balance of the address in FAT-0 or FAT-1 tokens.

#### Parameters:

| Name      | Type   | Description                | Validation                   | Required |
| --------- | ------ | -------------------------- | ---------------------------- | -------- |
| `address` | string | The public Factoid address | Valid Public Factoid address | Y        |

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
  "0cccd100a1801c0cf4aa2104b15dec94fe6f45d0f3347b016ed20d81059494df": 9007199254743259,
  "962a18328c83f370113ff212bae21aaf34e5252bc33d59c9db3df2a6bfda966f": 99694
},
  "id": 6482
}
```





## Error Codes

### `-32800` - Token Not Found

The given `token-id` & `issuer-id`, or `chain-id` did not result in a vaild known Issuance and Transaction chain. This is also used when `nf-token-id` does not resolve to a known non-fungible token.



### `-32801` - Invalid Token

The token found given `token-id` & `issuer-id`, or `chain-id` had an invalid issuance.



### `-32802` - Invalid Address

The given `fa-address` was not a valid Factoid address.



### `-32803` - Transaction Not Found

No transaction matching `tx-id` was found in the token's transaction database, or address transaction history



### `-32804` - Invalid Transaction

The submitted `tx`  FAT transaction object was invalid according to the validation rules of the standard



### `-32805` - Token Syncing

The token requested in the request is syncing. Please retry the request again later.



# Implementation



### HTTP Status code mapping for RPC errors (Optional)

| Error Code | HTTP Status Code |
| ---------- | ---------------- |
| -32800     | 404              |
| -32801     | 409              |
| -32802     | 400              |
| -32803     | 404              |
| -32804     | 400              |
| -32805     | 408              |



# Copyright

Copyright and related rights waived via
[CC0](https://creativecommons.org/publicdomain/zero/1.0/).
