package main

import (
	"log"
	"os"
)

var exitOnError = func(err error) {
	if err != nil {
		if os.Getenv("EXPERIA_V10_TEST_MODE") == "1" {
			log.Printf("exitOnError (test mode): %v", err)
			return
		}
		log.Fatal(err)
	}
}
