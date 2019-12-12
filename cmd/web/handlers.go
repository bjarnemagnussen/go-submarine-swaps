package main

import (
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/bjarnemagnussen/go-submarine-swaps/pkg/payreq"
	"github.com/btcsuite/btcd/btcec"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Check if the current request URL path exactly matches "/". If it doesn't, use
	// the http.NotFound() function to send a 404 response to the client.
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	// Initialize a slice containing the paths to the two files. Note that the
	// home.page.tmpl file must be the *first* file in the slice.
	files := []string{
		"./ui/html/home.page.tmpl",
		"./ui/html/base.layout.tmpl",
		"./ui/html/topnav.partial.tmpl",
		"./ui/html/footer.partial.tmpl",
	}

	// Use the template.ParseFiles() function to read the files and store the
	// templates in a template set. Notice that we can pass the slice of file paths
	// as a variadic parameter?
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// We then use the Execute() method on the template set to write the template
	// content as the response body. The last parameter to Execute() represents any
	// dynamic data that we want to pass in, which for now we'll leave as nil.
	err = ts.Execute(w, nil)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) refund(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w) // Use the notFound() helper.
		return
	}

	w.Write([]byte("Nothing to see on the refund page yet!"))
}

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
		// Initialize a slice containing the paths to the two files. Note that the
		// home.page.tmpl file must be the *first* file in the slice.
		files := []string{
			"./ui/html/create.page.tmpl",
			"./ui/html/base.layout.tmpl",
			"./ui/html/topnav.partial.tmpl",
			"./ui/html/footer.partial.tmpl",
			"./ui/html/form.partial.tmpl",
		}

		// Use the template.ParseFiles() function to read the files and store the
		// templates in a template set. Notice that we can pass the slice of file paths
		// as a variadic parameter?
		ts, err := template.ParseFiles(files...)
		if err != nil {
			app.serverError(w, err)
			return
		}

		// We then use the Execute() method on the template set to write the template
		// content as the response body. The last parameter to Execute() represents any
		// dynamic data that we want to pass in, which for now we'll leave as nil.
		err = ts.Execute(w, nil)
		if err != nil {
			app.serverError(w, err)
		}
	}
}
