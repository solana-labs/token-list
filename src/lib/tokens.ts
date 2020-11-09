import devnet from './../tokens/devnet.json';
import mainnetBeta from './../tokens/mainnet-beta.json';
import testnet from './../tokens/testnet.json';

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
