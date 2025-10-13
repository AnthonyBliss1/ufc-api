package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/anthonybliss1/ufc-api/scrape/utils"
)

func main() {
	fmt.Println("[Starting Scraping...]")
	fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")

	client := http.Client{}

	if err := utils.IterateFighters(&client); err != nil {
		log.Panic(err)
	}
}
