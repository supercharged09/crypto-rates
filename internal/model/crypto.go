package model

import "strings"

// CtyptoInfo - информация о криптовалюте для отображения
type CryptoInfo struct {
	ID     string //название в коингеко API:"bitcoin', "ethereum"
	Name   string //Человеческое название: "Bitcoin"
	Symbol string //Символ
	Emoji  string //Эмодзи
}

// SupportedCryptos — список поддерживаемых валют
var SupportedCryptos = []CryptoInfo{
	{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC", Emoji: "🟠"},
	{ID: "ethereum", Name: "Ethereum", Symbol: "ETH", Emoji: "🔷"},
	{ID: "tether", Name: "Tether", Symbol: "USDT", Emoji: "🟢"},
	{ID: "binancecoin", Name: "BNB", Symbol: "BNB", Emoji: "🟡"},
	{ID: "ripple", Name: "XRP", Symbol: "XRP", Emoji: "🔵"},
	{ID: "solana", Name: "Solana", Symbol: "SOL", Emoji: "🟣"},
}

// GetCryptoInfo возвращает информацию о валюте по ID или символу
func GetCryptoInfo(query string) *CryptoInfo {
	for i := range SupportedCryptos {
		c := &SupportedCryptos[i]
		if c.ID == query || strings.EqualFold(c.Symbol, query) || strings.EqualFold(c.Name, query) {
			return c
		}
	}
	return nil
}
