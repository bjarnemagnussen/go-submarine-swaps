package payreq

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/lightningnetwork/lnd/zpay32"
)

// PayReq is a type representing a swap payment request.
type PayReq struct {
	Invoice     string
	Destination string
	Currency    Currency
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
func decodeInvoiceWithCurrency(c Currency, bolt11 string) (PayReq, error) {
	inv, err := zpay32.Decode(bolt11, c.Chaincfg)
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
		Currency:    c,
		Destination: hex.EncodeToString(inv.Destination.SerializeCompressed()),
		CreatedAt:   inv.Timestamp,
		Expiry:      inv.Expiry(),
		Amount:      sats,
		Description: desc,
		PaymentHash: inv.PaymentHash[:],
	}, nil
}

// ValidateSameNetwork takes a Bech32 HRP and invoice to check if they are on
// the same network.
// func ValidateSameNetwork(hrp string, invoice string) error {
// 	// Checking Bech32 HRP
// 	depositCurrency, ok := bech32ToCurrency[hrp]
// 	if !ok {
// 		// Deposit currency not supported
// 		return fmt.Errorf("Deposit currency not supported")
// 	}
//
// 	// Checking invoice currency
// 	invoiceCurrency, err := GetCurrencyFromInvoice(invoice)
// 	if err != nil {
// 		// Lightning currency not supported
// 		return err
// 	}
//
// 	// Illegal to convert e.g. from testnet to mainnet and vice versa
// 	if depositCurrency.Chaincfg.Name != invoiceCurrency.Chaincfg.Name {
// 		return fmt.Errorf("Using this currency to pay this invoice is not supported")
// 	}
//
// 	return nil
// }
