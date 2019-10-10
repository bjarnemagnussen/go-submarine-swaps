package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	// Register the handler functions and corresponding URL patterns with the servemux.
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/swap", app.swap)
	mux.HandleFunc("/refund", app.refund)

	// Create a file server which serves files out of the "./ui/static" directory.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Use the mux.Handle() function to register the file server as the handler for
	// all URL paths that start with "/static/". For matching paths, we strip the
	// "/static" prefix before the request reaches the file server.
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	return mux
}
