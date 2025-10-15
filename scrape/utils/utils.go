package utils

import (
	"fmt"
	"log"
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

		// load body of response in goquery doc
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}

		// first find the table rows
		rows := doc.Find(".b-statistics__table tbody tr")

		// iterate through each row
		rows.Each(func(i int, tr *goquery.Selection) {
			if i == 0 {
				return
			}

			td := tr.ChildrenFiltered("td") // td represents each cell (or column) in the row
			fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
			fmt.Printf("Name: %s %s\n", strings.TrimSpace(td.Eq(0).Text()), strings.TrimSpace(td.Eq(1).Text()))

			// find the link to the fighter profile page
			link, _ := td.Eq(0).Find("a").Attr("href")
			fmt.Printf("Link to Profile: %v\n\n", link)

			// navigate to the profile page and collect all data on the fighter
			err := CollectFighterData(link, client)
			if err != nil {
				fmt.Printf("failed to collect data from fighter profile page: %v", err)
				return
			}
		})
	}

	return nil
}

func CollectFighterData(link string, client *http.Client) error {
	// build request
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return fmt.Errorf("failed to construct fighter profile request: %v", err)
	}

	// add the necessary request headers just to simulate the browser, avoiding potential issues
	req.Header.Add("referer", Referer)
	req.Header.Add("host", Host)
	req.Header.Add("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request fighter profile page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("request not accepted, Status Code: %d | %v", resp.StatusCode, err)
	}

	// load body of response in goquery doc
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	page := doc.Find(".l-page__container") // page container that holds all the data i need

	statsBox := page.Find(".b-fight-details").First() // contains physical stats, career stats, and fights
	// quick check to make sure we found something
	if statsBox.Length() == 0 {
		log.Fatal("No <fight-details> found")
	}

	pStats := statsBox.Find("div .b-list__box-list").First() // contains physical stats
	// quick check to make sure we found something
	if pStats.Length() == 0 {
		log.Fatal("No <ul> found")
	}

	nickname := strings.TrimSpace(page.Find("p.b-content__Nickname").Text())

	fmt.Printf("Nickname: %s\n", nickname)

	pStats.Each(func(i int, ul *goquery.Selection) {
		li := ul.ChildrenFiltered("li")

		height := strings.TrimSpace(li.Eq(0).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Height: %s\n", height)

		weight := strings.TrimSpace(li.Eq(1).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Weight: %s\n", weight)

		reach := strings.TrimSpace(li.Eq(2).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Reach: %s\n", reach)

		stance := strings.TrimSpace(li.Eq(3).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Stance: %s\n", stance)

		dob := strings.TrimSpace(li.Eq(4).Clone().Find("i").Remove().End().Text())
		fmt.Printf("DOB: %s\n", dob)
	})

	cStatsBoxLeft := statsBox.Find("div .b-list__info-box-left").First() // contains left side of career stats box
	cStatsLeft := cStatsBoxLeft.Find("ul.b-list__box-list").First()      //narrow down to the element containing the list elements

	cStatsLeft.Each(func(i int, ul *goquery.Selection) {
		li := ul.ChildrenFiltered("li")

		slpm := strings.TrimSpace(li.Eq(0).Clone().Find("i").Remove().End().Text())
		fmt.Printf("SLpM: %s\n", slpm)

		strAcc := strings.TrimSpace(li.Eq(1).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Str. Acc.: %s\n", strAcc)

		sapm := strings.TrimSpace(li.Eq(2).Clone().Find("i").Remove().End().Text())
		fmt.Printf("SApM: %s\n", sapm)

		strDef := strings.TrimSpace(li.Eq(3).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Str. Def.: %s\n", strDef)
	})

	cStatsBoxRight := statsBox.Find("div .b-list__info-box-right").First() // contains right side of career stats box
	cStatsRight := cStatsBoxRight.Find("ul.b-list__box-list").First()      // narrow down to the element containing the list elements

	cStatsRight.Each(func(i int, ul *goquery.Selection) {
		li := ul.ChildrenFiltered("li")

		tdAvg := strings.TrimSpace(li.Eq(1).Clone().Find("i").Remove().End().Text())
		fmt.Printf("TD Avg.: %s\n", tdAvg)

		tdAcc := strings.TrimSpace(li.Eq(2).Clone().Find("i").Remove().End().Text())
		fmt.Printf("TD Acc.: %s\n", tdAcc)

		tdDef := strings.TrimSpace(li.Eq(3).Clone().Find("i").Remove().End().Text())
		fmt.Printf("TD Def.: %s\n", tdDef)

		subAvg := strings.TrimSpace(li.Eq(4).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Sub. Avg.: %s\n\n", subAvg)
	})

	return nil
}
