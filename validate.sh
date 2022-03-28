#!/usr/bin/env bash
cue vet src/tokens/solana.tokenlist.json automerge/schema.cue -d '#Tokenlist'
new TokenListProvider().resolve().then((tokens) => {
  const tokenList = tokens.filterByClusterSlug('mainnet-beta').getList();
  console.log(tokenList);
});
