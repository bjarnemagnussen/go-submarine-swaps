package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/bjarnemagnussen/go-submarine-swaps/pkg/payreq"
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
		invoice := r.PostForm.Get("invoice")

		// Initialize a map to hold any validation errors.
		errors := make(map[string]string)

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

		// Dump the value content out in a plain-text HTTP response
		w.Write([]byte(fmt.Sprintf("invoice:\n%v\n\n", inv)))

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
