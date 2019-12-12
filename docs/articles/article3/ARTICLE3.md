# Creating A Broker For Submarine Swaps Part 3: Let's Build A Broker For Submarine Swaps

In Part 2: Decoding Lightning Invoices we developed a package called `payreq`, which allowed us to decode user submitted Lightning invoices.

In this part of the article series we will look into technical details on how to create Submarine Swaps and extend our `payreq` package adding functions to create the output scripts and deposit addresses.


# Disclaimer

**I have developed this project to the best of my knowledge. But I am no expert in web development and there may be mistakes and ways to optimize and better organize this code. It is only intended as an educational resource. Pull requests on the project's [repository](https://github.com/bjarnemagnussen/go-submarine-swaps) are very welcome!**

# Prerequisites

Throughout this article series we will mostly use Golang's standard library. Later we will also import packages to communicate with the Bitcoin and Lightning network and a database.

**Requirements:**

- Go $\geq$ v1.10

If you are familiar with Bitcoin and the Lightning Network and how they work on a technical level, then you should be able to easily follow along. Otherwise, since the basics won't be covered I suggest to brush up on topics such as Bitcoin's Script language before moving on. But it should hopefully not be too difficult to follow along, even if you’re fairly new.

**Suggestions to read up on:**

- [Hash functions](https://en.wikipedia.org/wiki/Hash_function).
- Generally about [public key cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography) and its [digital signatures](https://en.wikipedia.org/wiki/Digital_signature).
- Bitcoin [transactions](https://en.bitcoin.it/wiki/Transaction), especially its [Script](https://en.bitcoin.it/wiki/Script) language, see e.g. the Medium article [Behind the Scenes of a Bitcoin Transaction](https://medium.com/@raksha.p82/a-note-on-bitcoin-transaction-scripts-f04d298f1855) by Raksha M P or Chapter 7 of [Mastering Bitcoin](https://github.com/bitcoinbook/bitcoinbook) by Andreas Antonopolous.
- Some knowledge of the [Lightning Network](https://coincenter.org/entry/what-is-the-lightning-network).

**Optionally:**

- Go's [`html/template`](https://golang.org/pkg/html/template/) package, see [Golang article](https://golang.org/doc/articles/wiki/#tmp_6)
- Submarine Swaps, e.g. the Medium article [How do Submarine Swaps work?](https://medium.com/suredbits/how-do-submarine-swaps-work-907ed0d91498) by Torkel Rogstad.


# Project Repository

The code can be found at its project repository on [Github](https://github.com/bjarnemagnussen/go-submarine-swaps). Each part of the article series has its separate branch.

You can clone or download a starting point for the project here https://github.com/bjarnemagnussen/go-submarine-swaps:

```bash
git clone https://github.com/bjarnemagnussen/go-submarine-swaps.git
cd go-submarine-swaps
git fetch && git fetch --tags
git checkout part-3
```

# Submarine Swaps

There are at least two ways to send bitcoins from a regular on-chain address into an existing Lightning channel. Splicing could allow this by combining the open channel, close channel and on-chain Bitcoin outputs into a single channel. A splicing specification is under [current development](https://github.com/lightningnetwork/lightning-rfc/wiki/Lightning-Specification-1.1-Proposal-States#funding-and-splicing). We will therefore focus on another solution that is possible right now called Submarine Swaps. Submarine Swaps is based on the same principals as [_atomic swaps_](https://en.bitcoin.it/wiki/Atomic_swap). A Submarine Swap is a way to "buy" coins in a Lightning channel with on-chain bitcoins (and vice versa) using a middleman referred to as a broker.

This all is done trust-less, meaning that the broker cannot keep the received bitcoins without also paying coins into the Lightning channel. Furthermore, Submarine Swaps are not limited to receiving and paying of the same currency, but can be used across blockchains and cryptocurrencies.

What enables Submarine Swaps is the way Bitcoin allows expressing spending conditions with its _Script_ language.

## Uni- vs bi-directional Swaps

Although Submarine Swaps principally allows for bi-directional swaps, we will only build the platform going from on-chain to off-chain from the user's perspective. This means that the platform will receive a Lightning invoice to pay and generate a [pay-to-script-hash](https://en.bitcoin.it/wiki/Transaction#Pay-to-Script-Hash) (P2SH) deposit address for the user to pay coins to. Technically, swapping the opposite way is very similar, however the broker would have to lock up its coins before each swap, leading to possible denial-of-service attacks from the user which is unfeasible for the scope of this project.

## I :heart: Script

One of the powers of Bitcoin lies within its simple Script language to define spending conditions on outputs.

`<self-promotion>`This is what made me enthusiastic and fascinated about Bitcoin in the first place. I chose to write my CS master's thesis "[A Formal Programming Language For Bitcoin Transactions](https://github.com/bjarnemagnussen/Bitcoin-Masters-Thesis)" on this topic in the years 2015 to 2016.`</self-promotion>`

Spending bitcoins requires satisfying conditions expressed in Script by the previous sender. They are found inside the output scripts of a transaction. Typically, to prove e.g. ownership of bitcoins, the cryptographic signature of a public key that was defined in an output script must be provided with the spending transaction's input script.

{{< figure src="images/input_output_script.png" title="Image from https://medium.com/@raksha.p82/a-note-on-bitcoin-transaction-scripts-f04d298f1855" lightbox="true" >}}

But Script allows for far more complex spending conditions. To get more to know behind the workings of Bitcoin transactions, see [Prerequisites](#prerequisites) above.

## The Submarine Swap Script

Submarine Swaps are not limited to solely work with Script. Only very few and basic operators are needed in a language to make Submarine Swaps possible. Using the `btcutil` library we will however limit the platform to only work with Script based cryptocurrencies such as Bitcoin and Litecoin.

Let's deep dive 🦈 into it and glance at what a Submarine Swap script looks like:

```
OP_HASH160 <paymentHash> OP_EQUAL
OP_IF
  <brokerPubKey>
OP_ELSE
  <timeout> OP_CHECKLOCKTIMEVERIFY OP_DROP
  <userPubKey>
OP_ENDIF
OP_CHECKSIG
```

Due to Script's low-level nature, this may look a bit strange. In pseudocode we can write the script as:

```javascript
var controllingKey;

if (hash160(preimage) === paymentHash) {
  controllingKey = brokerPubKey;
} else if currentTime > refundTimeLock {
  controllingKey = userPubKey;
}

require validSignature(controllingKey)
```

There are different paths possible to spend from the Submarine Swap:

- **To complete the swap** the broker pays the Lightning invoice and receives the preimage to the payment hash value. Presenting it with a spending transaction triggers the `if` clause allowing to spend the bitcoins with a signature calculated from *the broker's key*.
- Alternatively, **the user can initiate a refund** if the broker does not pay the invoice (or claim the bitcoins!) before the _timeout_ by presenting any dummy value as preimage. This triggers the `else` clause, which verifies the timeout has indeed been exceeded and requires *the user's signature* to spend.

Those two paths guarantees the swap is trustless and atomic. Only if the broker pays the invoice will he receive the preimage that unlocks the bitcoins for him. If the broker never pays the invoice, then the user will be able to reclaim the coins from the deposit address.

## Hash Subtleties...

You may recall from [Part 2: Decoding Lightning Invoices]({{< ref "/post/lets-build-a-broker-for-submarine-swaps-part2/index.md#the-lightning-invoice-data-bolt11" >}}) that the payment hash value is a SHA-256 digest. The attentive reader will notice that we have instead used an `OP_HASH160` opcode inside the Submarine Swap script above. This calculates a HASH-160 digest from a provided preimage and compares it to the HASH-160 payment hash as defined in the script.

A> The HASH-160 function is a chained hash function consisting of `RIPEMD160(SHA256(preimage))`. In Bitcoin `OP_HASH160` is typically used inside scripts to save some space (20 vs 32 bytes for SHA-256).

We will use the HASH-160 because it is so simple to calculate from the Lightning payment hash by running it through RIPEMD-160 exactly once.

# Types Of Cryptocurrencies

Since Submarine Swaps can be used between different cryptocurrencies, we want the broker platform to handle that. We will however only focus on Bitcoin and Litecoin, as both are very similar and use the same Script language. For development purposes we will also add support for the testnet and simnet network.

We therefore first need a way to distinguish between currencies. In our `payreq` package we define a special currency type called `Currency`, which is a structure with a `Name` and `Chaincfg` field. This new type will be defined inside a new file `pkg/payreq/currency.go`.

**File: `pkg/payreq/currency.go`**
```golang
package payreq

import "github.com/btcsuite/btcd/chaincfg"

// Currency is a type representing a cryptocurrency. It has a name and
// parameters defining the chain of the cryptocurrency.
type Currency struct {
  Name     string
  Chaincfg *chaincfg.Params
}
```

The `Name` field will hold nothing more than the name of the specific cryptocurrency, e.g. _Bitcoin_, _Bitcoin Testnet_ and _Litecoin_. The `Chaincfg` field contains a pointer to the [`chaincfg.Params`](https://godoc.org/github.com/btcsuite/btcd/chaincfg#Params) structure we already met in [Part 2: Decoding Lightning Invoices]({{< ref "/post/lets-build-a-broker-for-submarine-swaps-part2/index.md" >}}). Recall that `chaicfg.Params` defines a cryptocurrency network by its parameters, specifying fields for e.g. the _network type_ (mainnet, testnet or simnet) and version bytes (_magics_) of addresses and keys. For Litecoin, we define and declare its network parameters with the variable `liteCoinParams`.

**File: `pkg/payreq/currency.go`**
```golang
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
  ...
)
```

We also define `liteCoinParamsTestnet` and `liteCoinParamsSimnet` with chaincode parameters corresponding to their appropriate values.

The `Bech32HRPSegwit` field is already known from the previous part. Here in our `Currency` structure however, we will use a few more fields. The `Name` field in `chaincfg.Params` is meant to be a human-readable network identifier and we will repurpose it to determine the network type of the implemented cryptocurrency. In particular this allows us to distinguish between general types of networks, e.g.:
- Mainnet,
- testnet, and
- simnet (only used during development).

This network separation independent of the cryptocurrency enables the application to easily determine which currencies are on comparable networks and hence swap-able with each other. For example, Bitcoin on mainnet can be swapped for Litcoin (or Bitcoin) on mainnet, however not for any coins on testnet, and vice-versa.

There exists many more parameters inside `chaincfg.Params`, which are not relevant for the broker platform and ignored here.

Note that for Bitcoin the `chaincfg.Params` already come [pre-defined](https://godoc.org/github.com/btcsuite/btcd/chaincfg#pkg-variables) with its network parameters. It therefore suffices to point to them using their variable names `chainfcg.MainNet`, `chainfcg.TestNet3Params` and `chainfcg.SimNet`.

For each flavour of the supported currencies we initiate the following variables of type `Currency` with their appropriate parameters:
- `Btc`: Bitcoin (Mainnet),
- `BtcTestnet`: Bitcoin Testnet,
- `BtcSimnet`: Bitcoin Simnet (only during development),
- `Ltc`: Litecoin (Mainnet),
- `LtcTestnet`: Litecoin (Testnet), and
- `LtcSimnet`: Litecoin (Simnet).

**File: `pkg/payreq/currency.go`**
```golang
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
```

As a helper tool we also declare a variable `bech32ToCurrency` to hold a mapping from a currency Bech32 HRP to its `Currency` representation (*`string`* $$\to$$ *`Currency`*). This allows quick lookup of currencies simply given their Bech32 HRPs.

**File: `pkg/payreq/currency.go`**
```golang
// bech32ToCurrency is a map of Bech32 HRPs back to their Currency.
var bech32ToCurrency = map[string]Currency{
  Btc.Chaincfg.Bech32HRPSegwit:        Btc,
  BtcTestnet.Chaincfg.Bech32HRPSegwit: BtcTestnet,
  BtcSimnet.Chaincfg.Bech32HRPSegwit:  BtcSimnet,
  Ltc.Chaincfg.Bech32HRPSegwit:        Ltc,
  LtcTestnet.Chaincfg.Bech32HRPSegwit: LtcTestnet,
  LtcSimnet.Chaincfg.Bech32HRPSegwit:  LtcSimnet,
}
```

Finally we will add a function `GetCurrency` that given the Bech32 HRP returns the `Currency` structure.

**File: `pkg/payreq/currency.go`**
```golang
// GetCurrency returns the Currency for a Bech32 HRP.
func GetCurrency(hrp string) (Currency, error) {
	// Check if Currency HRP is inside supported currencies.
	c, ok := bech32ToCurrency[hrp]
	if !ok {
		return Currency{}, fmt.Errorf("Invoice currency is not supported")
	}
	return c, nil
}
```

The currency Bech32 HRP is used to lookup the corresponding `Currency` type in `bech32ToCurrency` and return an error if it doesn't exist. We then return the whole currency structure to the caller.

## Improving The Existing Code

Prior we accepted any submitted Lightning invoice regardless of its currency by simply extracting its Bech32 HRP. But with our new `Currency` structure in place, let's update the `payreq` package to only allow currencies we have implemented support for.

We will first move the `GetCurrencyFromInvoice` function that is currently inside `pkg/payreq/decodepaymentrequest.go` it to `pkg/payreq/currency.go`, where it makes more sense to have it.

We will then make a small improvement by making use of our new `GetCurrency` function defined above. After extracting the currency prefix, we will return `GetCurrency` with the prefix.

**File: `pkg/payreq/currency.go`**
```golang
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
```

Let us continue by making a slight modification to the `PayReq` structure and `decodeInvoiceWithCurrency` function. We want to add a field `Currency` to the `PayReq`.

**File: `pkg/payreq/decodepaymentrequest.go`**

{{< highlight go "hl_lines=5" >}}
// PayReq is a type representing a swap payment request.
type PayReq struct {
  Invoice     string
  Destination string
  Currency    Currency // New currency field
  CreatedAt   time.Time
  Expiry      time.Duration
  Amount      uint64
  Description string
  PaymentHash []byte
}
{{< /highlight >}}

Inside `decodeInvoiceWithCurrency` we then pass the `chaincfg.Params` from the `Currency.Chaincfg` field directly to the `zpay32.Decode` function that decodes the invoice. We also add the currency to the `PayReq` we create.

**File: `pkg/payreq/decodepaymentrequest.go`**

{{< highlight go "hl_lines=4" >}}
// decodeInvoiceWithCurrency decodes a Lightning network Bolt11 invoice to a PayReq
// using a provided cryptocurrency.
func decodeInvoiceWithCurrency(c Currency, bolt11 string) (PayReq, error) {
  inv, err := zpay32.Decode(bolt11, c.Chaincfg) // Add the chaincfg parameter directly from the currency `c`
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
    Currency:    c, // Add the currency `c` to PayReq structure
    Destination: hex.EncodeToString(inv.Destination.SerializeCompressed()),
    CreatedAt:   inv.Timestamp,
    Expiry:      inv.Expiry(),
    Amount:      sats,
    Description: desc,
    PaymentHash: inv.PaymentHash[:],
  }, nil
}
{{< /highlight >}}

We will not be using the `Currency` field of the `PayReq` for now, but it will become useful in a later article when integrating the Lightning network.

After those changes we now only support currencies pre-defined in `bech32ToCurrency` and will return an error otherwise. This is necessary, as we can't possibly support _every_ cryptocurrency out there!

![cryptogods](images/cryptogods.jpg)

## Adding Deposit Currencies To The Frontend

To allow users to select their desired deposit currency, we now extend our form inside `ui/html/form.partial.tmpl` adding a drop-down menu. As options we naturally add the choice between Bitcoin (mainnet), Bitcoin Testnet, Bitcoin Simnet (for development) and the same for Litecoin, namely those currencies we defined earlier inside our `payreq` package.

The values used in the form are chosen to be the currencies' Bech32 HRP prefixes. This allows for easy lookup in our `bech32ToCurrency` map. In a later part of the article series we will see how to dynamically inject the deposit currencies to the form.

**File: `ui/html/form.partial.tmpl`**
```html
{{define "form"}}
  <form class="contact100-form validate-form" method="POST" action="/swap">
    ...
    <select class="selection-2 input100" name="deposit">
      <option>Choose Deposit Currency</option>
      <option value="bc">Bitcoin</option>
      <option value="tb">Bitcoin Testnet</option>
      <option value="sb">Bitcoin Simnet</option>
      <option value="ltc">Litecoin</option>
      <option value="tltc">Litecoin Testnet</option>
      <option value="sltc">Litecoin Simnet</option>
    </select>
    ...
    <input class="input100" type="text" name="invoice" placeholder="Enter your Lightning invoice">
    ...
  </form>
...
{{end}}
```

## Validating Currency Networks

We then update the `swap` handler to get the deposit and invoice currency from the submitted data, return errors if they are not defined and validate them against each other to make sure they are on similar networks. Highlighted code below contain the new additions.

**File: `cmd/web/handlers.go`**

{{< highlight go "hl_lines=17 23-40" >}}
func (app *application) swap(w http.ResponseWriter, r *http.Request) {
  // Use r.Method to check whether the request is using POST or not.
  if r.Method == "POST" {

    // First we call r.ParseForm() which adds any data in POST request bodies
    // to the r.PostForm map. This also works in the same way for PUT and PATCH
    // requests. If there are any errors, we use our app.ClientError helper to send
    // a 400 Bad Request response to the user.
    err := r.ParseForm()
    if err != nil {
      app.clientError(w, http.StatusBadRequest)
      return
    }

    // Use the r.PostForm.Get() method to retrieve the relevant data fields
    // from the r.PostForm map.
    dep := r.PostForm.Get("deposit")
    invoice := r.PostForm.Get("invoice")

    // Initialize a map to hold any validation errors.
    errors := make(map[string]string)

    // Checking deposit currency
    depCurrency, err := payreq.GetCurrency(dep)
    if err != nil {
      errors["deposit"] = "Deposit currency is not supported"
    }

    // Checking invoice currency
    invCurrency, err := payreq.GetCurrencyFromInvoice(invoice)
    if err != nil {
	     // Lightning currency not supported
      errors["invoice"] = err.Error()
    }

    // Check deposit and invoice currencies are on similar networks
    if errors["invoice"] == "" && depCurrency.Chaincfg.Name != invCurrency.Chaincfg.Name {
      errors["invoice"] = "Both deposit and invoice currencies must be on same network"
    }

    // Decode Lightning Bolt11 invoice
    inv, err := payreq.DecodeInvoice(invoice)
    if err != nil && errors["invoice"] == "" {
      errors["invoice"] = err.Error()
    }

    // If there are any errors, dump them in a plain text HTTP response and return
    // from the handler.
    if len(errors) > 0 {
      fmt.Fprint(w, errors)
      return
    }
    ...
  }
  ...
}
{{< / highlight >}}

## Converting Between Currencies

By supporting multiple currencies and the ability to exchange between those, we also need to be able to "convert" amounts between them using their exchange rates.

We will not support converting between currencies in any intelligent way. Instead we will use static exchange rates defined inside a map `RateMap` of the `payreq` package. It will be added to `pkg/payreq/currency.go`.

**File: `pkg/payreq/currencies.go`**

```golang
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
```

We will not make use of this function in this article, but converting between currencies will become useful in the following article when storing payment requests inside a database, from which it should be clear how many coins of the deposit currency must be deposited to pay the invoice.

# Deposits

We can now receive the client's desired deposit currency and the invoice to be paid. The next thing to do is to incorporate it inside a Submarine Swap script and generate the deposit address for it!

We will extend our `payreq` package with a file `pkg/payreq/addresses.go` that holds functions to generate the scripts and addresses. Let's take a look at the file's header first.

**File: `pkg/payreq/addresses.go`**
```golang
package payreq

import (
  "crypto/sha256"
  "encoding/hex"
  "errors"
  "fmt"
  "time"

  "github.com/btcsuite/btcd/btcec"
  "github.com/btcsuite/btcd/txscript"
  "github.com/btcsuite/btcutil"
  "github.com/btcsuite/btcutil/base58"
  "golang.org/x/crypto/ripemd160"
)
```

Those imported packages will be needed when generating the scripts and addresses. The `btcec` package implements public and private keys for Bitcoin (and Litecoin). The `base58` package is used to encode a legacy Bitcoin (and Litecoin) address, which uses the [Base58 encoding](https://en.wikipedia.org/wiki/Base58), and is a different encoding to Bech32 that we have already encountered. The `ripemd160` provides the equally named RIPEMD-160 hash function, which is used to calculate the HASH-160 payment hash value used in the script.

## Implementing Output Scripts

We extend the new go-file by implementing a function that returns the Submarine Swap script given the payment hash, the broker's and user's public key and a timeout value called `locktime`.

**File: `pkg/payreq/addresses.go`**
```golang
// createSubmarineScriptFromHash returns a Submarine Swap script given the
// payment hash and broker public key, and the user's recover key.
func createSubmarineScriptFromHash(paymentHash []byte, brokerKey, userKey *btcec.PublicKey, locktime int64) ([]byte, error) {
  if len(paymentHash) != 20 {
    return nil, fmt.Errorf("payment hash is of wrong length")
  }

  builder := txscript.NewScriptBuilder()

  builder.AddOp(txscript.OP_HASH160)
  builder.AddData(paymentHash)
  builder.AddOp(txscript.OP_EQUAL)
  builder.AddOp(txscript.OP_IF)
  builder.AddData(brokerKey.SerializeCompressed())
  builder.AddOp(txscript.OP_ELSE)
  builder.AddInt64(locktime)
  builder.AddOps([]byte{
    txscript.OP_CHECKLOCKTIMEVERIFY,
    txscript.OP_DROP
  })
  builder.AddData(userKey.SerializeCompressed())
  builder.AddOps([]byte{
    txscript.OP_ENDIF,
    txscript.OP_CHECKSIG
  })

  return builder.Script()
}
```

This may seem somewhat overwhelming at first sight, but it actually follows the exact same script we already discussed above under [Submarine Swaps Script](#the-submarine-swap-script).

We first check if the provided payment hash value is of exactly 20 bytes, which is the length of a HASH-160 value. We then initiate a new [`txscript.ScriptBuilder`](https://godoc.org/github.com/btcsuite/btcd/txscript#ScriptBuilder). The builder provides a facility for creating custom scripts with opcode values stored as constants inside `txscript`. Opcodes of the Submarine Swap script are added one-by-one using the builder's methods: `AddOp` for a single opcode, `AddData` for data and `AddOps` for subsequent opcode pushes.

The `locktime` value is directly pushed as an `int64` inside the script. It is up to the caller to define a valid locktime that follows the rules defined in [BIP-65](https://github.com/bitcoin/bips/blob/master/bip-0065.mediawiki).

A> **Note:** In a locktime a threshold value of 500000000 is used to distinguish between the meaning of the time. Values below the threshold are interpreted as the earliest _block height_ before spending is allowed, while values above it are interpreted as a _Unix timestamp_, where `500000000` is Tuesday, November 5th, 1985 at 00:53:20 UTC.

The public keys from the broker and user must be provided as a `btcec.PublicKey` type, which guarantees that whatever data they hold indeed is a valid public key on the Bitcoin (and Litecoin) [ECDSA curve](https://en.bitcoin.it/wiki/Elliptic_Curve_Digital_Signature_Algorithm). Then if `createSubmarineScriptFromHash` returns a script with no errors, it is a valid Submarine Swap script and spendable with the corresponding keys. To that end we serialize the keys for Script using their `SerializeCompressed` method.

Lastly we finalize the script and return it as a byte slice.

We will add another function called `CreateSubmarineSwapScript` that utilizes the helper function from above. It will be exported and generates the Submarine Swap script on a higher-level by taking as arguments a `PayReq`, a relative expiration as `time.Delta` and the broker's and user's public key.

**File: `pkg/payreq/addresses.go`**
```golang
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
```

The function internally uses the payment hash value from the `PayReq` and converts it to its HASH-160 digest. We then calculate the absolute `locktime` by adding the provided relative expiry timeout to the expiration datetime of the invoice. Those values are passed on to our `createSubmarineScriptFromHash` helper function returning the swap script for those parameters.

A> **Note:** The `locktime` is of type `time.Time` and passed on as Unix timestamp, meaning that it will safely commit to a date and not a block height inside the swap script.

## Generating Addresses

All the heavy work has now been done and what is left to do is simply generate the addresses from Submarine Swap scripts! Let's neglect Segwit for a moment and implement a function `CreateP2SHAddress` to generate a pay-to-script-hash (P2SH) address given any script.

A> A P2SH _output script_ commits to the hash value of the script that defines the conditions for spending, also referred to as the _redeem script_. The output script is `OP_HASH160 [20-byte-hash-value] OP_EQUAL` with the `[20-byte-hash-value]` being the HASH160 value of the redeem script.

**File: `pkg/payreq/addresses.go`**
```golang
// CreateP2SHAddress returns a P2SH address as string given the currency and script.
func CreateP2SHAddress(c Currency, script []byte) (string, error) {
  address := base58.CheckEncode(btcutil.Hash160(script), c.Chaincfg.ScriptHashAddrID)
  return address, nil
}
```

In those few lines of code a lot is actually going on! As arguments the function `CreateP2SHAddress` takes any redeem script (this will later be the Submarine Swap script) together with a Currency. As we only plan to implement Bitcoin and Litecoin we use the method `CheckEncode` from the `btcutil/base58` package for the address encoding.

The Currency structure contains a field that holds a script hash version byte and is used to generate an address for that currency. Recall that a Base58 encoded address is identified by its version byte as either a regular pay-to-public-key-hash (P2PKH) or P2SH. We use the version byte corresponding to P2SH inside the `Chaincfg.ScriptHashAddrID` field of the currency.

The payload for a P2SH address is simply the HASH-160 value of the redeem script. This hash function is implemented in the `btcutil` package as `Hash160`. We then use `base58.CheckEncode` on the payload and version byte to generate the address and return it.

Later we will be able to generate a deposit address simply by using our Submarine Swap script as the redeem script.

## Using Segwit Addresses

Hopefully you've got your head around the convoluted nature of P2SH. But we are in the end of 2019, and **we don't want to belong to those >40% _NOT_ [using Segwit!](https://p2sh.info/dashboard/db/segwit-usage?orgId=1)**

A> **Note:** We will only implement P2SH nested Segwit ([P2SH-P2WSH](https://github.com/bitcoin/bips/blob/master/bip-0141.mediawiki#P2WSH_nested_in_BIP16_P2SH)) and leave implementing _native_ Bech32 P2WSH addresses as an exercise and nice way for you to try out your skills :muscle:.

The steps to produce a P2SH-P2WSH are very similar to how we produced "pure" P2SH above. But instead of committing to the swap script as the redeem script, we commit to its witness script: `OP_0 [32-byte-hash-value]`. The `[32-byte-hash-value]` is the **SHA256** hash value of the swap script and referred to as the witness program while the `OP_0` defines the witness version.

A> **Note**: The witness version defines rules used to interpret the witness program. Currently only version 0 is defined.

Lastly, the P2SH output script commits to this witness script in the same way as before with `OP_HASH160 [20-byte-script-hash-value] OP_EQUAL`, where the `[20-byte-script-hash-value]` is the HASH160 value of the witness script.

It will become much clearer in code. Let's first create a helper function `createWitnessScript` that given a script produces its witness script. This function's logic follows naturally from the description above.

**File: `pkg/payreq/addresses.go`**
```golang
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
```

We simply calculate the witness program, which is the SHA256 hash value of the provided script, and commit to it as a version 0 witness program.

We can now ~~copy-and-paste `CreateP2SHAddress`~~ (never copy code!) create a wrapper function for `CreateP2SHAddress` that injects the witness script:

**File: `pkg/payreq/addresses.go`**
```golang
// CreateP2SHP2WSHAddress returns a P2SH nested P2WSH address as string given the currency and script.
func CreateP2SHP2WSHAddress(c Currency, script []byte) (string, error) {
  witnessScript, err := createWitnessScript(script)
  if err != nil {
    return "", err
  }

  return CreateP2SHAddress(c, witnessScript)
}
```

The `CreateP2SHP2WSHAddress` function takes the provided script and currency, creates its witness script and then passes it on to the legacy P2SH routine in `CreateP2SHAddress`.

We have now completed our `payreq` package and can make total use of it in the handlers!

![mission-accomplished](images/mission_accomplished.jpg)

## Creating Deposit Addresses

We will extend our favourite little `swap` handler to make use of the new address functions to generate deposit addresses. The changes made below are _after_ returning eventual errors from decoding the submitted invoice. Highlighted code below correspond to the additions made.

**File: `cmd/web/handlers.go`**

{{< highlight go "hl_lines=54-100 103-104" >}}
func (app *application) swap(w http.ResponseWriter, r *http.Request) {
  // Use r.Method to check whether the request is using POST or not.
  if r.Method == "POST" {
    // First we call r.ParseForm() which adds any data in POST request bodies
    // to the r.PostForm map. This also works in the same way for PUT and PATCH
    // requests. If there are any errors, we use our app.ClientError helper to send
    // a 400 Bad Request response to the user.
    err := r.ParseForm()
    if err != nil {
      app.clientError(w, http.StatusBadRequest)
      return
    }

    // Use the r.PostForm.Get() method to retrieve the relevant data fields
    // from the r.PostForm map.
    dep := r.PostForm.Get("deposit")
    invoice := r.PostForm.Get("invoice")

    // Initialize a map to hold any validation errors.
    errors := make(map[string]string)

    // Checking deposit currency
    depCurrency, err := payreq.GetCurrency(dep)
    if err != nil {
      errors["deposit"] = "Deposit currency is not supported"
    }

    // Checking invoice currency
    invCurrency, err := payreq.GetCurrencyFromInvoice(invoice)
    if err != nil {
      // Lightning currency not supported
      errors["invoice"] = err.Error()
    }

    // Check deposit and invoice currencies are on similar networks
    if errors["invoice"] == "" && depCurrency.Chaincfg.Name != invCurrency.Chaincfg.Name {
      errors["invoice"] = "Both deposit and invoice currencies must be on same network"
    }

    // Decode Lightning Bolt11 invoice
    inv, err := payreq.DecodeInvoice(invoice)
    if err != nil && errors["invoice"] == "" {
      errors["invoice"] = err.Error()
    }

    // If there are any errors, dump them in a plain text HTTP response and return
    // from the handler.
    if len(errors) > 0 {
      fmt.Fprint(w, errors)
      return
    }

    // Timedelta for expiration of Submarine Swap is three days.
    expireDelta := 3 * 24 * time.Hour

    // Create a the Submarine Swap script.

    // For development use pre-defined keys
    broker, err := hex.DecodeString("032C2CEC8D4D581F2589DB146339995000A1D399C6BFBAA4572AA84C3E11BE6939")
    if err != nil {
      app.serverError(w, fmt.Errorf("Invalid broker public key"))
      return
    }
    // Broker private key:
    // A7E3EA5D3B0A97EAE5CA1EB240902FADA0E0740F83F5DCCABD8E0BD60C3180F6

    brokerKey, err := btcec.ParsePubKey(broker, btcec.S256())
    if err != nil {
      app.serverError(w, err)
      return
    }

    // For development use pre-defined keys
    user, err := hex.DecodeString("039115E11C22A7699EFBB07CD2248E7158386D0EE0ECEC0188CBFBA21CE5C4BF42")
    if err != nil {
      app.serverError(w, fmt.Errorf("Invalid user public key"))
      return
    }
    // User private key:
    // 74DA34B76133040AB3F1EE6D40ED4BF338DEF1562DA8E6C3314BE19EB09AC2B1

    userKey, err := btcec.ParsePubKey(user, btcec.S256())
    if err != nil {
      app.serverError(w, err)
      return
    }

    swapScript, err := payreq.CreateSubmarineSwapScript(inv, expireDelta, brokerKey, userKey)
    if err != nil {
      app.serverError(w, fmt.Errorf("Cannot create Submarine Swap script: %s", err))
      return
    }

    // Create the P2SH nested P2WSH deposit address.
    address, err := payreq.CreateP2SHP2WSHAddress(depCurrency, swapScript)
    if err != nil {
      app.serverError(w, fmt.Errorf("Cannot create Submarine Swap deposit address: %s", err))
      return
    }

    // Dump the value content out in a plain-text HTTP response
    w.Write([]byte(fmt.Sprintf("invoice:\n%v\n\n", inv)))
    w.Write([]byte(fmt.Sprintf("deposit script:\n%v\n\n", swapScript)))
    w.Write([]byte(fmt.Sprintf("deposit address:\n%v\n\n", address)))

  } else if r.Method == "GET" { // Use r.Method to check whether the request is using GET or not.
    ...
  }
}
{{< / highlight >}}

Those are many lines of new code! But fortunately all of it has been covered already and now we just use it with the submitted data!

We globally define the expiration time delta `expireDelta` to be three days, meaning that the swap script allows for refunds three days after the invoice expires. There is no general rule as to what this value should be. Three days may seem excessive but is good enough for now under development.

For now we use **hard-coded** keys for both the broker platform and user. This is just a quick and dirty method to get all the functionality up and running. Later on we will let the user provide his own public key or could calculate one from a private key generated on the client's machine using Javascript. The broker key should also be provided by e.g. environment variables on the server.

The public keys are used as `btcec.PublicKey`, which we have already briefly discussed when generating the swap script. The main importance is that they provide the necessary structure to implement public keys and guarantees that whatever data they hold indeed is a valid public key on the Bitcoin (and Litecoin) [ECDSA curve](https://en.bitcoin.it/wiki/Elliptic_Curve_Digital_Signature_Algorithm).

We then simply call our `CreateSubmarineSwapScript` with those data as arguments to get the swap script and pass it on to the `CreateP2SHP2WSHAddress` function to create its deposit address.

Lastly we dump the invoice, script and address information as plain-text HTTP response.

# Things to improve

Instead of relying on the `string` type for addresses inside `pkg/payreq/addresses.go`, we may want to utilize the `Address` type from the `btcsuite/btcutil` package. This would guarantee correct encoding and decoding of addresses to make it more safe. This is left as an exercise to the reader making sure the codebase is understood before moving on the future articles :sunglasses:.
