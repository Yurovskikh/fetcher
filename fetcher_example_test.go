package fetcher

import (
	"fmt"
	"log"
	"time"
)

func ExampleNewFetcher() {
	fetcher := NewFetcher("http://localhost:8080", 5*time.Second)

	get, err := fetcher.Get()
	if err != nil {
		log.Printf("failed to fetcher get: %s", err)
	}
	fmt.Println(get)

	list, err := fetcher.List()
	if err != nil {
		log.Printf("failed to fetcher list: %s", err)
	}
	fmt.Println(list)
}
