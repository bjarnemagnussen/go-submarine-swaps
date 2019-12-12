package payreq

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

// createSubmarineScriptFromHash returns a Submarine Swap script given the
// payment hash and broker public key, and the user's recover key.
func createSubmarineScriptFromHash(paymentHash []byte, brokerKey, userKey *btcec.PublicKey, locktime int64) ([]byte, error) {
	if len(paymentHash) != 20 {
		return nil, fmt.Errorf("payment hash is of wrong length")
	}

	builder := txscript.NewScriptBuilder()

	builder.AddOp(txscript.OP_HASH160).AddData(paymentHash).AddOp(txscript.OP_EQUAL)
	builder.AddOp(txscript.OP_IF)
	builder.AddData(brokerKey.SerializeCompressed())
	builder.AddOp(txscript.OP_ELSE)
	builder.AddInt64(locktime).AddOps([]byte{txscript.OP_CHECKLOCKTIMEVERIFY, txscript.OP_DROP})
	builder.AddData(userKey.SerializeCompressed())
	builder.AddOps([]byte{txscript.OP_ENDIF, txscript.OP_CHECKSIG})

	return builder.Script()
}

// createWitnessScript returns a version 0 Witness Script from a given Witness Program
func createWitnessScript(script []byte) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write(script)
	if err != nil {
		return nil, err
	}

	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_0).AddData(h.Sum(nil))

	return builder.Script()
}

// CreateP2SHAddress returns a P2SH address as string given the currency and script.
func CreateP2SHAddress(c Currency, script []byte) (string, error) {
	address := base58.CheckEncode(btcutil.Hash160(script), c.Chaincfg.ScriptHashAddrID)
	return address, nil
}

// CreateP2SHP2WSHAddress returns a P2SH nested P2WSH address as string given the currency and script.
func CreateP2SHP2WSHAddress(c Currency, script []byte) (string, error) {
	witnessScript, err := createWitnessScript(script)
	if err != nil {
		return "", err
	}

	return CreateP2SHAddress(c, witnessScript)
}

// CreateSubmarineSwapScript returns a Submarine Swap script from defined
// currency, invoice, expiration, broker and user key.
func CreateSubmarineSwapScript(inv PayReq, expires time.Duration, brokerKey, userKey *btcec.PublicKey) ([]byte, error) {
	payhash := inv.PaymentHash

	// convert SHA-256 payment hash to HASH160 value.
	hash160 := ripemd160.New()
	_, err := hash160.Write(payhash)
	if err != nil {
		return nil, err
	}

	// Deposit must expire after the invoice expiring.
	locktime := inv.CreatedAt.Add(inv.Expiry).Add(expires)

	// Construct the Submarine Swap WitnessProgram from hash, remote key, recover
	// key and locktime.
	script, err := createSubmarineScriptFromHash(hash160.Sum(nil), brokerKey, userKey, locktime.Unix())
	if err != nil {
		return nil, err
	}

	return script, nil
}
