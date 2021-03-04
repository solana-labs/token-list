import test from 'ava';
import { ENV, Strategy, TokenListProvider } from './tokenlist';

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
