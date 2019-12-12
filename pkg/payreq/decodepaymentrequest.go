package payreq

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/lightningnetwork/lnd/zpay32"
)

// PayReq is a type representing a swap payment request.
type PayReq struct {
	Invoice     string
	Destination string
	CreatedAt   time.Time
	Expiry      time.Duration
	Amount      uint64
	Description string
	PaymentHash []byte
}

// DecodeInvoice will return a PayReq given a Lightning network Bolt11 invoice.
func DecodeInvoice(bolt11 string) (PayReq, error) {
	c, err := GetCurrencyFromInvoice(bolt11)
	if err != nil {
		return PayReq{}, err
	}

	return decodeInvoiceWithCurrency(c, bolt11)
}

// decodeInvoiceWithCurrency decodes a Lightning network Bolt11 invoice to a PayReq
// using a provided cryptocurrency.
func decodeInvoiceWithCurrency(c string, bolt11 string) (PayReq, error) {
	inv, err := zpay32.Decode(bolt11, &chaincfg.Params{Bech32HRPSegwit: c})
	if err != nil {
		return PayReq{}, fmt.Errorf("Problem decoding invoice")
	}

	if time.Since(inv.Timestamp.Add(inv.Expiry())) >= 0 {
		return PayReq{}, fmt.Errorf("Invoice has already expired")
	}

	var sats uint64
	if inv.MilliSat != nil {
		sats = uint64(*inv.MilliSat) / 1000
	}

	var desc string
	if inv.Description != nil {
		desc = *inv.Description
	}

	return PayReq{
		Invoice:     bolt11,
		Destination: hex.EncodeToString(inv.Destination.SerializeCompressed()),
		CreatedAt:   inv.Timestamp,
		Expiry:      inv.Expiry(),
		Amount:      sats,
		Description: desc,
		PaymentHash: inv.PaymentHash[:],
	}, nil
}

// GetCurrencyFromInvoice returns the Bech32 HRP of a Lightning network Bolt11
// invoice without validating the checksum of the invoice.
func GetCurrencyFromInvoice(bolt11 string) (string, error) {
	// Check that the invoice field is not blank.
	if strings.TrimSpace(bolt11) == "" {
		return "", fmt.Errorf("Lightning invoice is required")
	}

	// The Bech32 human-readable part for the currency is everything after the
	// first 'ln' until the first '1'.
	one := strings.IndexByte(bolt11, '1')
	if one < 3 || one+7 > len(bolt11) {
		return "", fmt.Errorf("Invalid index of 1")
	}
	hrp := bolt11[2:one]

	// Treat anything inside the HRP up to a digit as the currency prefix.
	amntIdx := strings.IndexFunc(hrp+"0", func(c rune) bool {
		return unicode.IsDigit(c)
	})

	return hrp[:amntIdx], nil
}
