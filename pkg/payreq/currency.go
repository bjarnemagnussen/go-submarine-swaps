package payreq

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/btcsuite/btcd/chaincfg"
)

// Currency is a type representing a cryptocurrency. It has a name and
// parameters defining the chain of the cryptocurrency.
type Currency struct {
	Name     string
	Chaincfg *chaincfg.Params
}

var (
	// liteCoinParams contains the custom parameters defined for the Litecoin chain.
	liteCoinParams = &chaincfg.Params{
		Name: "mainnet",

		// Human-readable part for Bech32 encoded segwit addresses, as defined in
		// BIP 173.
		Bech32HRPSegwit: "ltc", // always ltc for main net

		// Address encoding magics (version bytes)
		PubKeyHashAddrID:        0x30, // starts with L
		ScriptHashAddrID:        0x32, // starts with M
		PrivateKeyID:            0xB0, // starts with 6 (uncompressed) or T (compressed)
		WitnessPubKeyHashAddrID: 0x06, // starts with p2
		WitnessScriptHashAddrID: 0x0A, // starts with 7Xh
	}

	// liteCoinParamsTestnet contains the custom parameters defined for the Litecoin chain.
	liteCoinParamsTestnet = &chaincfg.Params{
		Name: "testnet3",

		// Human-readable part for Bech32 encoded segwit addresses, as defined in
		// BIP 173.
		Bech32HRPSegwit: "tltc", // always tltc for test net

		// Address encoding magics
		PubKeyHashAddrID:        0x6f, // starts with m or n
		ScriptHashAddrID:        0x3a, // starts with Q
		WitnessPubKeyHashAddrID: 0x52, // starts with QW
		WitnessScriptHashAddrID: 0x31, // starts with T7n
		PrivateKeyID:            0xef, // starts with 9 (uncompressed) or c (compressed)
	}

	// liteCoinParamsSimnet contains the custom parameters defined for the Litecoin chain.
	liteCoinParamsSimnet = &chaincfg.Params{
		Name: "simnet",

		// Human-readable part for Bech32 encoded segwit addresses, as defined in
		// BIP 173.
		Bech32HRPSegwit: "sltc", // always lsb for sim net

		// Address encoding magics
		PubKeyHashAddrID:        0x3f, // starts with S
		ScriptHashAddrID:        0x7b, // starts with s
		PrivateKeyID:            0x64, // starts with 4 (uncompressed) or F (compressed)
		WitnessPubKeyHashAddrID: 0x19, // starts with Gg
		WitnessScriptHashAddrID: 0x28, // starts with ?
	}

	// Btc represents the Bitcoin currency.
	Btc = Currency{
		Name:     "Bitcoin",
		Chaincfg: &chaincfg.MainNetParams,
	}

	// BtcTestnet represents the Bitcoin testnet currency.
	BtcTestnet = Currency{
		Name:     "Bitcoin Testnet",
		Chaincfg: &chaincfg.TestNet3Params,
	}

	// BtcSimnet represents the Bitcoin testnet currency.
	BtcSimnet = Currency{
		Name:     "Bitcoin Simnet",
		Chaincfg: &chaincfg.SimNetParams,
	}

	// Ltc represents the Litecoin currency.
	Ltc = Currency{
		Name:     "Litecoin",
		Chaincfg: liteCoinParams,
	}

	// LtcTestnet represents the Litecoin currency.
	LtcTestnet = Currency{
		Name:     "Litecoin Testnet",
		Chaincfg: liteCoinParamsTestnet,
	}

	// LtcSimnet represents the Litecoin currency.
	LtcSimnet = Currency{
		Name:     "Litecoin Simnet",
		Chaincfg: liteCoinParamsSimnet,
	}
)

// bech32ToCurrency is a map of Bech32 HRPs back to their Currency.
var bech32ToCurrency = map[string]Currency{
	Btc.Chaincfg.Bech32HRPSegwit:        Btc,
	BtcTestnet.Chaincfg.Bech32HRPSegwit: BtcTestnet,
	BtcSimnet.Chaincfg.Bech32HRPSegwit:  BtcSimnet,
	Ltc.Chaincfg.Bech32HRPSegwit:        Ltc,
	LtcTestnet.Chaincfg.Bech32HRPSegwit: LtcTestnet,
	LtcSimnet.Chaincfg.Bech32HRPSegwit:  LtcSimnet,
}

// RateMap contains exchange rates between Currencies.
var RateMap = map[Currency]map[Currency]float64{
	Btc: map[Currency]float64{
		Ltc: 10,
	},
	BtcTestnet: map[Currency]float64{
		LtcTestnet: 10,
	},
	BtcSimnet: map[Currency]float64{
		LtcSimnet: 10,
	},
	Ltc: map[Currency]float64{
		Btc: 0.1,
	},
	LtcTestnet: map[Currency]float64{
		BtcTestnet: 0.1,
	},
	LtcSimnet: map[Currency]float64{
		BtcSimnet: 0.1,
	},
}

// GetCurrency returns the Currency for a Bech32 HRP.
func GetCurrency(hrp string) (Currency, error) {
	// Check if Currency HRP is inside supported currencies.
	c, ok := bech32ToCurrency[hrp]
	if !ok {
		return Currency{}, fmt.Errorf("Invoice currency is not supported")
	}
	return c, nil
}

// GetCurrencyFromInvoice returns the Currency of a Lightning network Bolt11
// invoice without validating the checksum of the invoice.
func GetCurrencyFromInvoice(bolt11 string) (Currency, error) {
	if strings.TrimSpace(bolt11) == "" {
		return Currency{}, fmt.Errorf("Lightning invoice is required")
	}
	// The Bech32 human-readable part for the currency is everything after the
	// first 'ln' until the first '1'.
	one := strings.IndexByte(bolt11, '1')
	if one < 3 || one+7 > len(bolt11) {
		return Currency{}, fmt.Errorf("Invalid index of 1")
	}
	hrp := bolt11[2:one]

	// Treat anything inside the HRP up to a digit as the currency prefix.
	amntIdx := strings.IndexFunc(hrp+"0", func(c rune) bool {
		return unicode.IsDigit(c)
	})
	curHrp := hrp[:amntIdx]

	// Check if Currency HRP is inside supported currencies.
	return GetCurrency(curHrp)
}
