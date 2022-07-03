---
# 🚨🚨🚨This repository is EOL 🚨🚨🚨
## Read below for instructions on new token metadata flow
---

As of June 20th, this repository will be archived and will receive no more updates. The repository will be set to read-only and the npm package will still exist at `@solana/spl-token-registry`.

## Adding a New Token

You can use one of two tools at the time of writing:

1. [Strata Protocol Token Launchpad](https://app.strataprotocol.com/launchpad/manual/new)
2. [Token Creator Demo](https://token-creator-lac.vercel.app/)

All new token metadata will be added using Metaplex Fungible Token Metadata. The steps to add new Fungible Token Metadata are as follows:

1. Use `CreateMetadataV2` instruction from Metaplex token metadata to create new metadata for token.
2. Make sure you use the correct format for the token metadata.
3. You must have mint authority in order to create or update the metadata

The token metadata for Metaplex Metadata Schema is in the following format:

```json
{
  "name": "TOKEN_NAME", 
  "symbol": "TOKEN_SYMBOL",
  "uri": "TOKEN_URI",
  "sellerFeeBasisPoints": 0,
  "creators": null,
  "collection": null,
  "uses": null
}
```

The `TOKEN_URI` must point to a file with the following format:

```json
{
  "name": "TOKEN_NAME",
  "symbol": "TOKEN_SYMBOL",
  "description": "TOKEN_DESC",
  "image": "TOKEN_IMAGE_URL"
}
```

Where `TOKEN_IMAGE_URL` is the image url.

An example of the `TOKEN_URI`: https://token-creator-lac.vercel.app/token_metadata.json

Which resolves to:

```json
{
  "name": "A test token",
  "symbol": "TEST",
  "description": "Fully for testing purposes only",
  "image": "https://token-creator-lac.vercel.app/token_image.png"
}
```

## Updating Token Metadata

To update token metadata you must use `createUpdateMetadataAccountV2Instruction` in `@metaplex-foundation/js` in order to update an existing token's metadata.

While updating, you provide the same details as when creating.

## Tools for Adding/Updating/Migrating

Update/migrate token metadata using [Strata Protocol update token tool](https://app.strataprotocol.com/edit-metadata).

A tutorial for adding/updating metadata can be found at the [Token-Creator demo](https://github.com/jacobcreech/Token-Creator).


## Reading Legacy Token-list

`@solana/spl-token-registry`

[![npm](https://img.shields.io/npm/v/@solana/spl-token-registry)](https://unpkg.com/@solana/spl-token-registry@latest/) [![GitHub license](https://img.shields.io/badge/license-APACHE-blue.svg)](https://github.com/solana-labs/token-list/blob/b3fa86b3fdd9c817139e38641d46c5a892542a52/LICENSE)

Solana Token Registry is a package that allows application to query for list of tokens.
The JSON schema for the tokens includes: chainId, address, name, decimals, symbol, logoURI (optional), tags (optional), and custom extensions metadata.

### Installation

```bash
npm install @solana/spl-token-registry
```

```bash
yarn add @solana/spl-token-registry
```

### Examples

#### Query available tokens

```typescript
new TokenListProvider().resolve().then((tokens) => {
  const tokenList = tokens.filterByClusterSlug('mainnet-beta').getList();
  console.log(tokenList);
});
```

#### Render icon for token in React

```typescript jsx
import React, { useEffect, useState } from 'react';
import { TokenListProvider, TokenInfo } from '@solana/spl-token-registry';


export const Icon = (props: { mint: string }) => {
  const [tokenMap, setTokenMap] = useState<Map<string, TokenInfo>>(new Map());

  useEffect(() => {
    new TokenListProvider().resolve().then(tokens => {
      const tokenList = tokens.filterByChainId(ENV.MainnetBeta).getList();

      setTokenMap(tokenList.reduce((map, item) => {
        map.set(item.address, item);
        return map;
      },new Map()));
    });
  }, [setTokenMap]);

  const token = tokenMap.get(props.mint);
  if (!token || !token.logoURI) return null;

  return (<img src={token.logoURI} />);

```

# Disclaimer

All claims, content, designs, algorithms, estimates, roadmaps,
specifications, and performance measurements described in this project
are done with the Solana Foundation's ("SF") good faith efforts. It is up to
the reader to check and validate their accuracy and truthfulness.
Furthermore nothing in this project constitutes a solicitation for
investment.

Any content produced by SF or developer resources that SF provides, are
for educational and inspiration purposes only. SF does not encourage,
induce or sanction the deployment, integration or use of any such
applications (including the code comprising the Solana blockchain
protocol) in violation of applicable laws or regulations and hereby
prohibits any such deployment, integration or use. This includes use of
any such applications by the reader (a) in violation of export control
or sanctions laws of the United States or any other applicable
jurisdiction, (b) if the reader is located in or ordinarily resident in
a country or territory subject to comprehensive sanctions administered
by the U.S. Office of Foreign Assets Control (OFAC), or (c) if the
reader is or is working on behalf of a Specially Designated National
(SDN) or a person subject to similar blocking or denied party
prohibitions.

The reader should be aware that U.S. export control and sanctions laws
prohibit U.S. persons (and other persons that are subject to such laws)
from transacting with persons in certain countries and territories or
that are on the SDN list. As a project based primarily on open-source
software, it is possible that such sanctioned persons may nevertheless
bypass prohibitions, obtain the code comprising the Solana blockchain
protocol (or other project code or applications) and deploy, integrate,
or otherwise use it. Accordingly, there is a risk to individuals that
other persons using the Solana blockchain protocol may be sanctioned
persons and that transactions with such persons would be a violation of
U.S. export controls and sanctions law. This risk applies to
individuals, organizations, and other ecosystem participants that
deploy, integrate, or use the Solana blockchain protocol code directly
(e.g., as a node operator), and individuals that transact on the Solana
blockchain through light clients, third party interfaces, and/or wallet
software.
