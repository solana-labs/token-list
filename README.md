---
# Please note: This repository is being rebuilt to accept the new volume of token additions and modifications. PR merges will be delayed.
---

# Contents
* [Usage](#usage)
* [Adding new token](#adding-new-token)
* [Modifying existing token](#modifying-existing-token)
* [Disclaimer](#disclaimer)


# Usage

@solana/spl-token-registry

[![npm](https://img.shields.io/npm/v/@solana/spl-token-registry)](https://unpkg.com/@solana/spl-token-registry@latest/) [![GitHub license](https://img.shields.io/badge/license-APACHE-blue.svg)](https://github.com/solana-labs/token-list/blob/b3fa86b3fdd9c817139e38641d46c5a892542a52/LICENSE)

Solana Token Registry is a package that allows application to query for list of tokens.
The JSON schema for the tokens includes: chainId, address, name, decimals, symbol, logoURI (optional), tags (optional), and custom extensions metadata.

## Installation

```bash
npm install @solana/spl-token-registry
```

```bash
yarn add @solana/spl-token-registry
```

## Examples

### Query available tokens

```typescript
new TokenListProvider().resolve().then((tokens) => {
  const tokenList = tokens.filterByClusterSlug('mainnet-beta').getList();
  console.log(tokenList);
});
```

### Render icon for token in React

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

# Adding new token

To add a new token, add another json block to the large `tokens` list in `src/tokens/solana.tokenlist.json` and submit a PR.

Tips:
* `logoURI` 
  * should point to a `png`, `jpg`, or `svg`.
  * the logo can be hosted in this repo in `assets/mainnet/TOKEN_ADDRESS/FILE` 
    * in that case, the image should be added to this repo and logoURI should point to `https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/TOKEN_ADDRESS/FILE`)
  * if your logo is hosted in any other repo or any other location, no need to add it here
* `tags`
  * please don't go crazy here, note that the valid tags are in the toplevel `tags` section and any other tags will likely have no effect
* `extensions: 
  * the `extensions` block can contain links to your twitter, discord, etc.  A list of allowed extensions is [here](automerge/schema.cue#L105).
  * `serumV3Usdc` and `serumV3Usdt` are the addresses of serum markets for your token (either paired with USDC or USDT)
  * `coingeckoId` is the string that appears as 'API id' on the corresponding coingecko page
* it's recommended to not add your token as the final element to the list (second-to-last is best).  This is because adding the token as the final element will create merge conflicts that are more difficult for maintainers to manually resolve.
* solscan, solana explorer, and wallets all pull from this repo at different cadences; some update every few days.  Please do not raise issues with us saying 'solscan has updated but phantom has not'.

Changes will be automerged.  If automerge fails, you can click the 'Details' link for more information. 

Please follow the Uniswap Token List specification found here: https://github.com/Uniswap/token-lists



# Modifying existing token

Modifications currently must be manually reviewed.  For any modifications, please submit a PR, then raise an issue with a link to your PR in order to request manual review.

Tips:
* Please be sure to modify the existing JSON block instead of adding a new one.  We see many commits that just add a new block instead of amending the existing one.
* Please check the 'Files changed' tab on your PR to ensure that your change is as expected.



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
