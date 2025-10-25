package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anthonybliss1/ufc-api/scrape/data"
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

	fmt.Print("\n[Environment Variables Set!]\n\n")
}

func main() {
	var update = flag.Bool("update", false, "run update function only")
	var upcoming = flag.Bool("upcoming", false, "collect upcoming events and matchups")

	flag.Parse()

	client, err := utils.CreateProxyClient()
	if err != nil {
		log.Fatalf("[Proxy Client Build Failed: %v]", err)
	}

	// start a timer to track the scraping process speed
	start := time.Now()

	switch true {
	case *update:
		// only collect most recent data not in db
		fmt.Println("[Starting Update...]")
		fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")

		if err := utils.RunUpdate(client); err != nil {
			log.Panic(err)
		}
	case *upcoming:
		fmt.Println("[Starting Upcoming Collection...]")
		fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
		event := data.Event{}

		if err := utils.IterateUpcomingEvents(&event, client); err != nil {
			log.Panic(err)
		}
	default:
		// collect all data
		fmt.Println("[Starting Complete Refresh...]")
		fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")

		if err := utils.IterateFighters(client); err != nil {
			log.Panic(err)
		}
	}

	// after all data is collected load batches into the mongodb
	fmt.Println("[Running Batches...]")
	fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
	utils.RunBatches()

	// measure time elapsed from the 'start' timestamp
	elapsed := time.Since(start)

	fmt.Println("\n[Process Completed!]")
	fmt.Printf("[Time: %.2fH]\n", elapsed.Hours())
}
