package main

import (
	"github.com/evgeniuz/shortener/shortener/handlers"
	"github.com/evgeniuz/shortener/shortener/store/bolt"
	"log"
)

func main() {
	store, err := bolt.NewStore("shortener.db", 7)
	if err != nil {
		log.Fatal(err)
	}

	shortener, err := handlers.NewShortener(store)
	if err != nil {
		log.Fatal(err)
	}

	err = shortener.Listen(8080)
	if err != nil {
		log.Fatal(err)
	}
}
