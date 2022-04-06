---
### Please note: This repository is being rebuilt to accept the new volume of token additions and modifications. PR merges will be delayed.
---

# Contents
* [Usage](#usage)
* [Adding new token](#adding-new-token)
* [Modifying existing token](#modifying-existing-token)
* [Common issues](#common-issues)
  * [Automerge failure: found removed line](#automerge-failure-found-removed-line)
  * [Failed to normalize: failed to parse JSON: json: unknown field](#failed-to-normalize-failed-to-parse-json-json-unknown-field)
  * [Duplicate token](#duplicate-token)
  * [Scanner/wallet hasn't updated yet](#scannerwallet-hasnt-updated-yet)
  * [error validating schema: chainId: conflicting values 103 and 0](#error-validating-schema-chainid-conflicting-values-103-and-0)
  * [warning about the last element in the list](#warning-about-the-last-element-in-the-list)
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
* please squash commits into a single commit for cleanliness

Changes will be automerged.  If automerge fails, you can click the 'Details' link for more information. 

Please follow the Uniswap Token List specification found here: https://github.com/Uniswap/token-lists


# Modifying existing token

Modifications currently must be manually reviewed.  For any modifications, please submit a PR, then raise an issue with a link to your PR (and leave the PR open) in order to request manual review.

* please check the 'Files changed' tab on your PR to ensure that your change is as expected
* please link the commit or PR where the token was originally added.  If the token was added by someone else, they will be asked to confirm that this change is authorized
* please squash commits into a single commit for cleanliness


# Common issues

## Automerge failure: found removed line
Any modifications must be manually merged; please submit an issue with a link to your PR (and leave the PR open).


## Failed to normalize: failed to parse JSON: json: unknown field
e.g. `failed to normalize: failed to parse JSON: json: unknown field "coingeckoId"`

If this error is encountered while modifying an existing entry, note that this error is misleading; 
it is the automerger's way of saying that adding `coingeckoId` to an existing entry is not allowed.

Any modifications must be manually merged; please submit an issue with a link to your PR (and leave the PR open).


## Duplicate token
"duplicate token: token address `...` is already used"

This occurs because the diff in your PR is re-adding a completely new block for a token that was already previously added. (You can verify this by looking at the 'Files changed' tab of your PR.)

This usually happens because your PR is intended to _update_ an existing token, but it still includes the commits that _added_ the original token (which were previously merged).  You can verify this by checking the 'commits' tab of the PR.  If you see the original commit in there, that's bad!  The PR should be relative to the current `HEAD` of `main`, i.e. your checkout should be [rebased](https://git-scm.com/book/en/v2/Git-Branching-Rebasing)

To fix this, you can either:

1. checkout the latest `HEAD` of `main` and then re-apply your change (simpler for git newbies but incurring a bit of duplicate work), or 
2. rebase your local checkout back to `origin/main` before opening a PR.  

For option (2), you can do this with:
```
git remote add pub-origin git@github.com:solana-labs/token-list.git
git fetch pub-origin main 
git rebase pub-origin/main
git push origin main -f
```

More generally, for modifications to existing tokens, be sure to checkout the `HEAD` of the `main` branch, locate the existing block in `solana.tokenlist.json`, and modify the appropriate fields.

Always check the 'Files changed' tab on your PR to see the impact of your change.


## Scanner/wallet hasn't updated yet
Solscan, solana explorer, and wallets all pull from this repo at different cadences.  Some update every few days.  

If your change has landed in the `HEAD` of `main` branch, it was successful, but it might take a few days for downstream users to reflect that change.

Please especially do not raise issues saying 'solscan has updated but phantom has not', that definitely means your change is in this repo!


## error validating schema: chainId: conflicting values 103 and 0
This automerge error arises if you touched a line outside of your token block.  Some text editors introduce a diff to the final line of the file.  You can see this by looking at the "Files changed" tab of your PR.

If using vim, you can probably address this by adding
```
set nofixendofline
```
to `~/.vimrc`

If you don't address this yourself, the PR will need to be manually merged; please submit an issue and link your PR.


## Warning about the last element in the list
Please do not add your token as the final element to the list (second-to-last is best).
This is because when the token is the final element, the closing brace won't be followed by a comma, which creates a specialcase which will create a merge conflict if the commit doesn't get automerged.  This prevents the maintainers from manually merging your change in the event that it needs to be automerged.

If the maintainers link you to this comment, it means you need to move your block in order for them to merge it.

Addressing this more seamlessly is an open item; bear with us for now.


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
