package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anthonybliss1/ufc-api/scrape/utils"
)

func main() {
	fmt.Println("[Starting Scraping...]")
	fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")

	client := http.Client{}

	// start a timer to track the scraping process speed
	start := time.Now()

	if err := utils.IterateFighters(&client); err != nil {
		log.Panic(err)
	}

	// measure time elapsed from the 'start' timestamp
	elapsed := time.Since(start)

	fmt.Println("\n[Scraping Completed!]")
	fmt.Printf("[Time: %.2fs]\n", elapsed.Seconds())
}
