package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// should dynamically collect this. not sure if this will affect status of request
const UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0"

// may not be REQUIRED as headers for request but good to have to avoid potential issues
const Referer = "http://ufcstats.com/statistics/fighters"
const Host = "ufcstats.com"

// query param 'char=' letters to iterate through
var Letters = []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i",
	"j", "k", "l", "m", "n", "o", "p", "q", "r",
	"s", "t", "u", "v", "w", "x", "y", "z",
}

func IterateFighters(client *http.Client) error {
	for _, letter := range Letters {
		fmt.Printf("[Scraping fighters under letter '%s']\n", letter)
		url := fmt.Sprintf("http://ufcstats.com/statistics/fighters?char=%s&page=all", letter)

		// build request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to construct request alphabetical page: %s | %v", letter, err)
		}

		// add the necessary request headers just to simulate the browser, avoiding potential issues
		req.Header.Add("referer", Referer)
		req.Header.Add("host", Host)
		req.Header.Add("User-Agent", UserAgent)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to request alphabetical page: %s | %v", letter, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("request not accepted, Status Code: %d | %v", resp.StatusCode, err)
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}

		rows := doc.Find(".b-statistics__table tbody tr")

		rows.Each(func(i int, tr *goquery.Selection) {
			if i == 0 {
				return
			}

			td := tr.ChildrenFiltered("td")
			fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
			fmt.Printf("Name: %s %s\n", strings.TrimSpace(td.Eq(0).Text()), strings.TrimSpace(td.Eq(1).Text()))
			fmt.Printf("Nickname: %s\n", strings.TrimSpace(td.Eq(2).Text()))
			fmt.Printf("Height: %s\n", strings.TrimSpace(td.Eq(3).Text()))
			fmt.Printf("Weight: %s\n", strings.TrimSpace(td.Eq(4).Text()))
			fmt.Printf("Reach: %s\n", strings.TrimSpace(td.Eq(5).Text()))
			fmt.Printf("Stance: %s\n", strings.TrimSpace(td.Eq(6).Text()))
			fmt.Printf("Wins: %s\n", strings.TrimSpace(td.Eq(7).Text()))
			fmt.Printf("Loss: %s\n", strings.TrimSpace(td.Eq(8).Text()))
			fmt.Printf("Draw: %s\n\n", strings.TrimSpace(td.Eq(9).Text()))
		})
	}

	return nil
}
