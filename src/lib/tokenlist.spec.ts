import test from 'ava';

import { CLUSTER_SLUGS, ENV, Strategy, TokenListProvider } from './tokenlist';

test('Token list is filterable by a tag', async (t) => {
  const list = (await new TokenListProvider().resolve(Strategy.Static))
    .filterByChainId(ENV.MainnetBeta)
    .filterByTag('nft')
    .getList();

  t.false(list.some((item) => item.symbol === 'SOL'));
});

test('Token list can exclude by a tag', async (t) => {
  const list = (await new TokenListProvider().resolve(Strategy.Static))
    .filterByChainId(ENV.MainnetBeta)
    .excludeByTag('nft')
    .getList();

  t.false(list.some((item) => item.tags === ['nft']));
});

test('Token list can exclude by a chain', async (t) => {
  const list = (await new TokenListProvider().resolve(Strategy.Static))
    .excludeByChainId(ENV.MainnetBeta)
    .getList();

  t.false(list.some((item) => item.chainId === ENV.MainnetBeta));
});

test('Token list returns new object upon filter', async (t) => {
  const list = await new TokenListProvider().resolve(Strategy.Static);
  const filtered = list.filterByChainId(ENV.MainnetBeta);
  t.true(list !== filtered);
  t.true(list.getList().length !== filtered.getList().length);
});

test('Token list throws error when calling filterByClusterSlug with slug that does not exist', async (t) => {
  const list = await new TokenListProvider().resolve(Strategy.Static);
  const error = await t.throwsAsync(
    async () => list.filterByClusterSlug('whoop'),
    { instanceOf: Error }
  );
  t.is(
    error.message,
    `Unknown slug: whoop, please use one of ${Object.keys(CLUSTER_SLUGS)}`
  );
});
