package tokenlist

// Solana-specific derivative of https://uniswap.org/tokenlist.schema.json,
// converted to a CUE schema from JSON Schema.
//
// The current solana.tokenlist.json does not validate against the
// Uniswap upstream schema! Deviations are marked INCOMPATIBLE.

import (
	"strings"
	"list"
	"struct"
)

#Base58Address: =~"^[1-9A-HJ-NP-Za-km-z]{43,44}$"

#EthAddress: =~"^0x[0-9a-fA-F]{40}$"

// Grandfathered non-compliant symbol names.
#SymbolWhitelist: ("GÜ" |
	"W technology" |
	"SHBL LP token" |
	"Unlimited Energy" |
	"Need for Speed" |
	"ADOR OPENS" |
	"CMS - Rare" |
	"Power User" |
	"VIP Member" |
	"Uni Christmas" |
	"Satoshi Closeup" |
	"Satoshi GB" |
	"Satoshi OG" |
	"Satoshi BTC" |
	"APESZN_HOODIE" |
	"APESZN_TEE_SHIRT" |
	"Satoshi Closeup" |
	"Satoshi BTC" |
	"Satoshi Nakamoto" |
	"Charles Hoskinson" |
	"Bitcoin Tram" |
	"SRM tee-shirt" |
	"USDT_ILT" |
	"NINJA NFT1" |
	"USDC/USDT[stable]" |
	"mSOL/SOL[stable]" |
	"Nordic Energy Token" |
	"USDC-wUSDC-wUSDT-wDAI" )

// Grandfathered non-compliant token names.
#NameWhitelist: (
	"Mike Krow's Official Best Friend Super Kawaii Kasu Token" |
	"B ❤ P" |
	"PHISHING SCAM TOKEN, PLEASE IGNORE" )

// INCOMPATIBLE: may contain -
// INCOMPATIBLE: max 20 characters (vs. 10)
#TagIdentifier: strings.MinRunes(1) & strings.MaxRunes(20) & =~"^[\\w-]+$"

#TagDefinition: {
	// The name of the tag
	// INCOMPATIBLE: may contain -
	name: =~"^[ \\w-]+$" & strings.MinRunes(1) & strings.MaxRunes(20)

	// A user-friendly description of the tag
	// INCOMPATIBLE: may contain -
	description: =~"^[ \\w\\.,:-]+$" & strings.MinRunes(1) & strings.MaxRunes(200)
}

#Version: {
	// The major version of the list. Must be incremented when tokens
	// are removed from the list or token addresses are changed.
	major: int & >=0

	// The minor version of the list. Must be incremented when tokens
	// are added to the list.
	minor: int & >=0

	// The patch version of the list. Must be incremented for any
	// changes to the list.
	patch: int & >=0
}

#URL: =~ #"^(ipfs|http[s]?)://(?:[a-zA-Z]|[0-9]|[$-_@.&+#~]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+$"#
#TelegramURL: =~ #"^https://t.me/([\w\+]{5,32}|joinchat/[\w-]{16})$"#

#Extensions: {
	website?: #URL
	twitter?: =~ #"^https://twitter.com/(\w){1,15}$"#
	telegram?: #TelegramURL
	telegramAnnouncements?: #TelegramURL
	serumV3Usdc?: #Base58Address
	serumV3Usdt?: #Base58Address
	coingeckoId?: =~ #"^[\w-]+$"#
	address?: #Base58Address | #EthAddress | "uusd" | "uluna"
	bridgeContract?: #URL
	assetContract?: #URL
	discord?: #URL
	medium?: #URL
	instagram?: #URL
	reddit?: #URL
	coinmarketcap?: #URL
	facebook?: #URL
	github?: #URL
	youtube?: #URL
	waterfallbot?: #URL
	dexWebsite?: #URL
	imageUrl?: #URL
	animationUrl?: #URL
	linkedin?: #URL
	description?: string & strings.MinRunes(1) & strings.MaxRunes(2000)
	blog?: #URL
	vault?: #URL
	whitepaper?: #URL
	twitch?: #URL
	solanium?: #URL
	vaultPubkey?: #Base58Address
}

#TokenInfo: {
	// The chain ID of the Solana network where this token is
	// deployed.
	chainId: 101 | 102 | 103

	// The checksummed address of the token on the specified chain ID
	// INCOMPATIBLE: base58
	address: #Base58Address

	// The number of decimals for the token balance
	decimals: int & >=0 & <=255

	// The name of the token
	name: strings.MinRunes(1) & strings.MaxRunes(50) & =~"^[ \\w.'+\\-%/À-ÖØ-öø-ÿ:&\\[\\]\\(\\)]+$" | #NameWhitelist

	// The symbol for the token; must be alphanumeric
	symbol: =~"^[a-zA-Z0-9+\\-%/$_.]+$" & strings.MinRunes(1) & strings.MaxRunes(20) | #SymbolWhitelist

	// A URI to the token logo asset; if not set, interface will
	// attempt to find a logo based on the token address; suggest SVG
	// or PNG of size 64x64
	logoURI?: #URL | ""

	// An array of tag identifiers associated with the token; tags are
	// defined at the list level
	tags?:       list.MaxItems(10) & [...#TagIdentifier]

	extensions?: #Extensions
}

#Tokenlist: {
	// The name of the token list
	name: strings.MinRunes(2) & strings.MaxRunes(20)

	// The timestamp of this list version; i.e. when this immutable
	// version of the list was created
	timestamp: string
	version:   #Version

	// The list of tokens included in the list
	tokens: list.MaxItems(10000) & [...#TokenInfo] & [_, ...]

	// Keywords associated with the contents of the list; may be used
	// in list discoverability.
	//
	// INCOMPATIBLE: keywords can contain -
	keywords?: list.UniqueItems() & list.MaxItems(20) & [...strings.MinRunes(1) & strings.MaxRunes(20) & =~"^[\\w -]+$"]

	// A mapping of tag identifiers to their name and description
	tags?: struct.MaxFields(20) & {
		[#TagIdentifier]: _
	} & {
		[string]: #TagDefinition
	}

	// A URI for the logo of the token list; prefer SVG or PNG of size
	// 256x256
	logoURI?: string
}

// Extra checks applied for new tokens only, not when processing the full file
#StrictTokenInfo: #TokenInfo & {
	// Require logoURI to be set
	logoURI: #URL
}
