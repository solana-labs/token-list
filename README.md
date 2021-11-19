---
# Please note: This repository is being rebuilt to accept the new volume of token additions and modifications. PR merges will be delayed.
---

# @solana/spl-token-registry

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

## Adding new token

Submit PR with changes to JSON file `src/tokens/solana.tokenlist.json`

Please follow the Uniswap Token List specification found here: https://github.com/Uniswap/token-lists

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
