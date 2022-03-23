package parser

import "encoding/json"

type Token struct {
	ChainId    int               `json:"chainId"`
	Address    string            `json:"address"`
	Symbol     string            `json:"symbol"`
	Name       string            `json:"name"`
	Decimals   int               `json:"decimals"`
	LogoURI    string            `json:"logoURI"`
	Tags       []string          `json:"tags,omitempty"`
	Extensions map[string]string `json:"extensions,omitempty"`
}

type TokenList struct {
	Name      string          `json:"name"`
	LogoURI   string          `json:"logoURI"`
	Keywords  []string        `json:"keywords"`
	Tags      json.RawMessage `json:"tags"`
	Timestamp string          `json:"timestamp"`
	Tokens    []Token         `json:"tokens"`
	Version   struct {
		Major int `json:"major"`
		Minor int `json:"minor"`
		Patch int `json:"patch"`
	} `json:"version"`
}

