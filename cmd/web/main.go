package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Define an application struct to hold the application-wide dependencies for the
// web application. For now we'll only include fields for the two custom loggers, but
// we'll add more to it as the build progresses.
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {
	// Define a new command-line flag with the name 'port' with a default value of "8080"
	port := flag.Int("port", 8080, "HTTP network address")
	flag.Parse()

	// Use log.New() to create a logger for writing information messages.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize a new instance of application containing the dependencies.
	app := application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	// Initialize a http.Server struct.
	srv := &http.Server{
		Addr:     fmt.Sprintf(":%d", *port),
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	// Write messages using the two loggers, instead of the standard logger.
	infoLog.Printf("Starting server on %d", *port)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
