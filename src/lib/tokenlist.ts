import { fetch } from 'cross-fetch';

import tokenlist from './../tokens/solana.tokenlist.json';

export enum ENV {
  MainnetBeta = 101,
  Testnet = 102,
  Devnet = 103,
}

export interface TokenList {
  readonly name: string;
  readonly logoURI: string;
  readonly tags: { [tag: string]: TagDetails };
  readonly timestamp: string;
  readonly tokens: TokenInfo[];
}

export interface TagDetails {
  readonly name: string;
  readonly description: string;
}

export interface TokenExtensions {
  readonly website?: string;
  readonly bridgeContract?: string;
  readonly assetContract?: string;
  readonly address?: string;
  readonly explorer?: string;
  readonly twitter?: string;
  readonly github?: string;
  readonly medium?: string;
  readonly tgann?: string;
  readonly tggroup?: string;
  readonly discord?: string;
  readonly serumV3Usdt?: string;
  readonly serumV3Usdc?: string;
  readonly coingeckoId?: string;
  readonly imageUrl?: string;
  readonly description?: string;
}

export interface TokenInfo {
  readonly chainId: number;
  readonly address: string;
  readonly name: string;
  readonly decimals: number;
  readonly symbol: string;
  readonly logoURI?: string;
  readonly tags?: string[];
  readonly extensions?: TokenExtensions;
}

export type TokenInfoMap = Map<string, TokenInfo>;

export const CLUSTER_SLUGS: { [id: string]: ENV } = {
  'mainnet-beta': ENV.MainnetBeta,
  testnet: ENV.Testnet,
  devnet: ENV.Devnet,
};

export class GitHubTokenListResolutionStrategy {
  repositories = [
    'https://raw.githubusercontent.com/solana-labs/token-list/main/src/tokens/solana.tokenlist.json',
  ];

  resolve = () => {
    return queryJsonFiles(this.repositories);
  };
}

export class CDNTokenListResolutionStrategy {
  repositories = [
    'https://cdn.jsdelivr.net/gh/solana-labs/token-list@main/src/tokens/solana.tokenlist.json',
  ];

  resolve = () => {
    return queryJsonFiles(this.repositories);
  };
}

export class SolanaTokenListResolutionStrategy {
  repositories = ['https://token-list.solana.com/solana.tokenlist.json'];

  resolve = () => {
    return queryJsonFiles(this.repositories);
  };
}

const queryJsonFiles = async (files: string[]) => {
  const responses: TokenList[] = (await Promise.all(
    files.map(async (repo) => {
      try {
        const response = await fetch(repo);
        const json = (await response.json()) as TokenList;
        return json;
      } catch {
        console.info(
          `@solana/token-registry: falling back to static repository.`
        );
        return tokenlist;
      }
    })
  )) as TokenList[];

  return responses
    .map((tokenlist: TokenList) => tokenlist.tokens)
    .reduce((acc, arr) => (acc as TokenInfo[]).concat(arr), []);
};

export enum Strategy {
  GitHub = 'GitHub',
  Static = 'Static',
  Solana = 'Solana',
  CDN = 'CDN',
}

export class StaticTokenListResolutionStrategy {
  resolve = () => {
    return tokenlist.tokens;
  };
}

export class TokenListProvider {
  static strategies = {
    [Strategy.GitHub]: new GitHubTokenListResolutionStrategy(),
    [Strategy.Static]: new StaticTokenListResolutionStrategy(),
    [Strategy.Solana]: new SolanaTokenListResolutionStrategy(),
    [Strategy.CDN]: new CDNTokenListResolutionStrategy(),
  };

  resolve = async (
    strategy: Strategy = Strategy.CDN
  ): Promise<TokenListContainer> => {
    return new TokenListContainer(
      await TokenListProvider.strategies[strategy].resolve()
    );
  };
}

export class TokenListContainer {
  constructor(private tokenList: TokenInfo[]) {}

  filterByTag = (tag: string) => {
    return new TokenListContainer(
      this.tokenList.filter((item) => (item.tags || []).includes(tag))
    );
  };

  filterByChainId = (chainId: number | ENV) => {
    return new TokenListContainer(
      this.tokenList.filter((item) => item.chainId === chainId)
    );
  };

  excludeByChainId = (chainId: number | ENV) => {
    return new TokenListContainer(
      this.tokenList.filter((item) => item.chainId !== chainId)
    );
  };

  excludeByTag = (tag: string) => {
    return new TokenListContainer(
      this.tokenList.filter((item) => !(item.tags || []).includes(tag))
    );
  };

  filterByClusterSlug = (slug: string) => {
    if (slug in CLUSTER_SLUGS) {
      return this.filterByChainId(CLUSTER_SLUGS[slug]);
    }
    throw new Error(
      `Unknown slug: ${slug}, please use one of ${Object.keys(CLUSTER_SLUGS)}`
    );
  };

  getList = () => {
    return this.tokenList;
  };
}
