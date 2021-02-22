import devnet from './../tokens/devnet.json';
import mainnetBeta from './../tokens/mainnet-beta.json';
import testnet from './../tokens/testnet.json';

import * as cross from 'cross-fetch';

export enum ENV {
  MainnetBeta = 'mainnet-beta',
  Testnet = 'testnet',
  Devnet = 'devnet',
}

export const TOKENS = {
  [ENV.MainnetBeta]: mainnetBeta,
  [ENV.Testnet]: testnet,
  [ENV.Devnet]: devnet,
};

export interface KnownToken {
  tokenSymbol: string;
  tokenName: string;
  icon?: string;
  mintAddress: string;
}

export type KnownTokenMap = Map<string, KnownToken>;

export class SolanaTokenListResolutionStrategy {
  resolve = async (network: string) => {
    throw new Error(`Not Implemented Yet. ${network}`,);
  }
}

export class StaticTokenListResolutionStrategy {
  resolve = async (network: string) => {
    return TOKENS[network as ENV] as KnownToken[];
  }
}

const queryJsonFiles = async (network: string, files: string[]) => {
  const responses = await Promise.all(files.map(async repo => {
    const response = await cross.fetch(`${repo}/${network}.json`);
    const json = await response.json() as KnownToken[];

    return json;
  }));

  return responses.reduce((acc, arr) => acc.concat(arr), []);
}

export class GitHubTokenListResolutionStrategy {
  repositories = [
    'https://github.com/solana-labs/token-list/tree/main/src/tokens',
    'https://github.com/project-serum/serum-ts/tree/master/packages/tokens/src'
  ];

  resolve = async (network: string) => {
    return queryJsonFiles(network, this.repositories);
  }
}

export class CDNTokenListResolutionStrategy {
  repositories = [
    'https://cdn.jsdelivr.net/gh/solana-labs/token-list@tree/main/src/tokens',
    'https://cdn.jsdelivr.net/gh/project-serum/serum-ts@tree/master/packages/tokens/src'
  ];

  resolve = async (network: string) => {
    return queryJsonFiles(network, this.repositories);
  }
}

export enum Strategy {
  GitHub = 'GitHub',
  Static = 'Static',
  Solana = 'Solana',
  CDN = 'CDN',
}

export class TokenListProvider {
  static strategies = {
    [Strategy.GitHub]: new GitHubTokenListResolutionStrategy(),
    [Strategy.Static]: new StaticTokenListResolutionStrategy(),
    [Strategy.Solana]: new SolanaTokenListResolutionStrategy(),
    [Strategy.CDN]: new CDNTokenListResolutionStrategy(),
  }


  resolve = async (network: string, strategy: Strategy = Strategy.CDN) => {
    return TokenListProvider.strategies[strategy].resolve(network);
  };
}
