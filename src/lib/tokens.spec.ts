import test from 'ava';
import { ENV, TOKENS } from './tokens';

test('Token env is array', (t) => {
  t.true(Array.isArray(TOKENS[ENV.MainnetBeta]));
  t.true(Array.isArray(TOKENS[ENV.Testnet]));
  t.true(Array.isArray(TOKENS[ENV.Devnet]));
});
