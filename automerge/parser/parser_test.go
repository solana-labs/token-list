package parser

import (
	"fmt"
	"testing"
)

var testData = []string{`
    },
    {
      "chainId": 101,
      "address": "8tGqYibsn9ZYv7513DLyNHUn5pBGKektsg4gKrMPfQrF",
      "symbol": "GRMW",
      "name": "GremWorld",
      "decimals": 9,
      "logoURI": "https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/8tGqYibsn9ZYv7513DLyNHUn5pBGKektsg4gKrMPfQrF/logo.jpg",
      "tags": [
        "Metaverse"
      ]
`,
	`
   {
     "symbol": "USDT-USDC",
     "name": "Saber USDT-USDC LP",
     "logoURI": "https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/2poo1w1DL6yd2WNTCnNTzDqkC6MBXq7axo77P16yrBuf/icon.png",
     "decimals": 6,
     "address": "2poo1w1DL6yd2WNTCnNTzDqkC6MBXq7axo77P16yrBuf",
     "chainId": 101,
     "tags": [
       "saber-stableswap-lp"
     ],
     "extensions": {
       "website": "https://app.saber.so/#/pools/usdc_usdt"
     }
   },
   {
     "symbol": "PAI-USDC",
     "name": "Saber PAI-USDC LP",
     "logoURI": "https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/PaiYwHYxr4SsEWox9YmyBNJmxVG7GdauirbBcYGB7cJ/icon.png",
     "decimals": 6,
     "address": "PaiYwHYxr4SsEWox9YmyBNJmxVG7GdauirbBcYGB7cJ",
     "chainId": 101,
     "tags": [
       "saber-stableswap-lp"
     ],
     "extensions": {
       "website": "https://app.saber.so/#/pools/usdc_pai"
     }
   },
   {
     "symbol": "BTC-renBTC",
     "name": "Saber BTC-renBTC LP",
     "logoURI": "https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/SLPbsNrLHv8xG4cTc4R5Ci8kB9wUPs6yn6f7cKosoxs/icon.png",
     "decimals": 8,
     "address": "SLPbsNrLHv8xG4cTc4R5Ci8kB9wUPs6yn6f7cKosoxs",
     "chainId": 101,
     "tags": [
       "saber-stableswap-lp"
     ],
     "extensions": {
       "website": "https://app.saber.so/#/pools/btc"
     }
   },
   {
     "symbol": "pBTC-renBTC",
     "name": "Saber pBTC-renBTC LP",
     "logoURI": "https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/pBTCmyG7FaZx4uk3Q2pT5jHKWmWDn84npdc7gZXpQ1x/icon.png",
     "decimals": 8,
     "address": "pBTCmyG7FaZx4uk3Q2pT5jHKWmWDn84npdc7gZXpQ1x",
     "chainId": 101,
     "tags": [
       "saber-stableswap-lp"
     ],
     "extensions": {
       "website": "https://app.saber.so/#/pools/pbtc"
     }
   },
   {
     "symbol": "CASH-USDC",
     "name": "Saber CASH-USDC LP",
     "logoURI": "https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/CLPKiHjoU5HwpPK5L6MBXHKqFsuzPr47dM1w4An3Lnvv/icon.png",
     "decimals": 6,
     "address": "CLPKiHjoU5HwpPK5L6MBXHKqFsuzPr47dM1w4An3Lnvv",
     "chainId": 101,
     "tags": [
       "saber-stableswap-lp"
     ],
     "extensions": {
       "website": "https://app.saber.so/#/pools/cash"
     }
   },
`,
	`
	{
		"chainId": 101,
		"address": "43UsEVeUuzHhM3vtB7a9c5Hy2mC27S24Exj24HsAqCYc",
		"symbol": "WILL",
		"name": "Will",
		"decimals": 9,
		"logoURI": "https://raw.githubusercontent.com/CyberGothica/WILL/main/logo.png",
		"tags": [
			"game-token",
			"game-currency"
		],
		"extensions": {
			"twitter": "https://twitter.com/Cyber_Gothica",
			"discord": "https://discord.com/channels/885149106341830666"
		}
	},
`,
	`    {
      "chainId": 101,
      "address": "476ZdKh1xue32zNzFWvnyaDEncrBEdq99sDiZXSGyyJu",
      "symbol": "TOF",
      "name": "Toffee Token",
      "decimals": 0,
      "logoURI": "https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/476ZdKh1xue32zNzFWvnyaDEncrBEdq99sDiZXSGyyJu/logo.png",
      "tags": [
        "utility-token"
      ],
      "extensions": {}
    },
`,
`}
{
  "chainId": 101,
  "address": "6t72LbKAVPD1Kq3b5vTMsfPsPUewLTxtBe57fngFUg7",
  "symbol": "ONGR",
  "name" : "OnigiriCoin",
  "decimals": 8,
  "logoURI": "https://github.com/arudeboy/arudeboy/blob/main/OnigiriCoin.JPG",
  "tags":[],
  "extensions":{
    "twitter": "https://twitter.com/OnigiriCoin"
  }
}
{
	"chainId": 101,
	"address": "9qXxEVGagc9ccd6b135Z8ZLr4VAWUd7T5KcmMyjYKBdB",
	"symbol": "HTK",
	"name": "HotoketCoin",
	"decimals": 8,
	"logoURI": "https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/9qXxEVGagc9ccd6b135Z8ZLr4VAWUd7T5KcmMyjYKBdB/logo.png",
	"tags":[],
	"extensions":{}
}`,
}

func TestNormalizeWhatever(t *testing.T) {
	for i, test := range testData {
		t.Run(fmt.Sprintf("Fixture-%d", i), func(t *testing.T) {
			if s, err := NormalizeWhatever(test); err != nil {
				t.Errorf("%s", err)
			} else {
				t.Log(s)
			}
		})
	}
}
