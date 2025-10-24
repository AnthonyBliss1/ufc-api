package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anthonybliss1/ufc-api/scrape/utils"
	"github.com/joho/godotenv"
)

//go:embed .env
var embeddedEnv string

func init() {
	godotenv.Load(".env")
	envMap, err := godotenv.Unmarshal(embeddedEnv)
	if err != nil {
		log.Fatalf("[Failed to unmarshal embeddedEnv file: %v]\n", err)
	}

	for key, value := range envMap {
		os.Setenv(key, value)
	}

	fmt.Print("[Environment Variables Set!]\n\n")
}

func main() {
	client, err := utils.CreateProxyClient()
	if err != nil {
		log.Fatalf("[Proxy Client Build Failed: %v]", err)
	}

	fmt.Println("[Starting Scraping...]")
	fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")

	// start a timer to track the scraping process speed
	start := time.Now()

	if err := utils.IterateFighters(client); err != nil {
		log.Panic(err)
	}

	// after all data is collected load batches into the mongodb
	fmt.Println("[Running Batches...]")
	fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
	utils.RunBatches()

	// if err := utils.CollectUpcomingEventData(&client); err != nil {
	// 	log.Panic(err)
	// }

	// measure time elapsed from the 'start' timestamp
	elapsed := time.Since(start)

	fmt.Println("\n[Process Completed!]")
	fmt.Printf("[Time: %.2fH]\n", elapsed.Hours())
}
