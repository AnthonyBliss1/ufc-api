package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
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
		page := fmt.Sprintf("http://ufcstats.com/statistics/fighters?char=%s&page=all", letter)

		// build request
		req, err := http.NewRequest("GET", page, nil)
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

			fighterName := fmt.Sprintf("%s %s", strings.TrimSpace(td.Eq(0).Text()), strings.TrimSpace(td.Eq(1).Text()))

			// find the link to the fighter profile page
			link, _ := td.Eq(0).Find("a").Attr("href")
			u, err := url.Parse(link)
			if err != nil {
				log.Panic("cannot parse url")
			}

			fighterID := path.Base(u.Path)

			fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
			fmt.Printf("Fighter Name: %s | Fighter Link: %s | FighterID: %s\n", fighterName, link, fighterID)

			// navigate to the profile page and collect all data on the fighter
			err = CollectFighterData(link, client)
			if err != nil {
				fmt.Printf("failed to collect data from fighter profile page: %v", err)
				return
			}
		})
	}

	return nil
}

func CollectFighterData(fighterProfileLink string, client *http.Client) error {
	// build request
	req, err := http.NewRequest("GET", fighterProfileLink, nil)
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

	fighterStats := page.Find(".b-fight-details").First() // contains physical stats, career stats, and fights
	// quick check to make sure we found something
	if fighterStats.Length() == 0 {
		log.Fatal("No <fight-details> found")
	}

	pStats := fighterStats.Find("div .b-list__box-list").First() // contains physical stats
	// quick check to make sure we found something
	if pStats.Length() == 0 {
		log.Fatal("No <ul> found")
	}

	nickname := strings.TrimSpace(page.Find("p.b-content__Nickname").Text())

	fmt.Printf("Nickname: %s\n", nickname)

	// PHYSCIAL AND CAREER STATISTICS COLLECTION
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

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

	// LEFT SIDE OF CAREER STATS
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~

	cStatsBoxLeft := fighterStats.Find("div .b-list__info-box-left").First() // contains left side of career stats box
	cStatsLeft := cStatsBoxLeft.Find("ul.b-list__box-list").First()          //narrow down to the element containing the list elements

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

	// RIGHT SIDE OF CAREER STATS
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~

	cStatsBoxRight := fighterStats.Find("div .b-list__info-box-right").First() // contains right side of career stats box
	cStatsRight := cStatsBoxRight.Find("ul.b-list__box-list").First()          // narrow down to the element containing the list elements

	cStatsRight.Each(func(i int, ul *goquery.Selection) {
		li := ul.ChildrenFiltered("li")

		tdAvg := strings.TrimSpace(li.Eq(1).Clone().Find("i").Remove().End().Text())
		fmt.Printf("TD Avg.: %s\n", tdAvg)

		tdAcc := strings.TrimSpace(li.Eq(2).Clone().Find("i").Remove().End().Text())
		fmt.Printf("TD Acc.: %s\n", tdAcc)

		tdDef := strings.TrimSpace(li.Eq(3).Clone().Find("i").Remove().End().Text())
		fmt.Printf("TD Def.: %s\n", tdDef)

		subAvg := strings.TrimSpace(li.Eq(4).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Sub. Avg.: %s\n", subAvg)
	})

	// FIGHT HISTORY
	// ~~~~~~~~~~~~~~

	fightRows := fighterStats.Find(".b-fight-details__table tbody tr")
	if fightRows.Length() == 0 {
		log.Fatal("cannot find fightRows")
	}

	// for debugging outputs
	var numFights int
	if fightRows.Length() == 0 {
		numFights = 0
	} else {
		numFights = fightRows.Length() - 1
	}

	// for debugging output
	fmt.Print("\n -----------------------\n")
	fmt.Printf("| Total Fights Found: %d |\n", numFights)
	fmt.Print(" -----------------------\n\n")

	fightRows.Each(func(i int, tr *goquery.Selection) {
		if i == 0 {
			return
		}

		// first need to capture the fight url for each of the fighter's fights
		td := tr.ChildrenFiltered("td")
		fightLink, e := td.Eq(0).Find("a").Attr("href")
		if e == false {
			log.Fatal("cannot find fight link")
		}
		// parse out link so i can grab the base path so i can save it as the FightID
		fLink, err := url.Parse(fightLink)
		if err != nil {
			log.Fatalf("failed to parse fight url: %v", err)
		}
		fightID := path.Base(fLink.Path)

		// for each fight in the fighters profile, find the fight link, capture the FightID and stats -> create fights struct
		fmt.Printf("Fight #%d | Fight Link: %s | FightID: %s\n\n", i, fightLink, fightID)

		if err = CollectFightData(fightLink, fighterProfileLink, client); err != nil {
			log.Fatalf("failed to collect fight data: %v", err)
		}

	})

	return nil
}

func CollectFightData(fightLink string, reqReferer string, client *http.Client) error {
	requestFight, err := http.NewRequest("GET", fightLink, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for fight: %v", err)
	}

	requestFight.Header.Add("referer", reqReferer)
	requestFight.Header.Add("host", Host)
	requestFight.Header.Add("User-Agent", UserAgent)

	resp, err := client.Do(requestFight)
	if err != nil {
		return fmt.Errorf("failed to submit request for fight: %v", err)
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

	page := doc.Find(".l-page__container")                   // page that contains all fight data
	fightEvent := page.Find("h2.b-content__title a").First() // element that contains the name and href of the event
	fightDetails := page.Find("div.b-fight-details").First()

	// FIGHT DATA - HEADER TABLE
	// ~~~~~~~~~~~~~~

	eventLink, _ := fightEvent.Attr("href")

	if err := CollectEventDetails(eventLink, fightLink, client); err != nil {
		log.Fatalf("failed to collect fighter event: %v", err)
	}

	participants := fightDetails.Find(".b-fight-details__person")
	p1Header := participants.Eq(0)
	p2Header := participants.Eq(1)

	p1Name := strings.TrimSpace(p1Header.Find("a").Text())
	p2Name := strings.TrimSpace(p2Header.Find("a").Text())

	p1Outcome := strings.TrimSpace(p1Header.Find("i").Text())
	p2Outcome := strings.TrimSpace(p2Header.Find("i").Text())

	fmt.Println("[ Fight Details ]")
	fmt.Printf("P1: %s - %s \nP2: %s - %s\n", p1Name, p1Outcome, p2Name, p2Outcome)

	boutType := strings.TrimSpace(fightDetails.Find(".b-fight-details__fight-head").First().Text())
	fmt.Printf("Type: %s\n", boutType)

	fightDetailsRow1 := fightDetails.Find(".b-fight-details__text").Eq(0)

	method := fightDetailsRow1.Find("i[style]").Text()
	fmt.Printf("Method: %s\n", strings.TrimSpace(method))

	round := fightDetailsRow1.Find(".b-fight-details__text-item").Eq(0).Text()
	fmt.Printf("Round: %s\n", strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(round), "Round:")))

	endTime := fightDetailsRow1.Find(".b-fight-details__text-item").Eq(1).Text()
	fmt.Printf("Time: %s\n", strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(endTime), "Time:")))

	format := fightDetailsRow1.Find(".b-fight-details__text-item").Eq(2).Text()
	fmt.Printf("Time Format: %s\n", strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(format), "Time format:")))

	referee := fightDetailsRow1.Find(".b-fight-details__text-item").Eq(3).Text()
	fmt.Printf("Referee: %s\n", strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(referee), "Referee:")))

	fightDetailsRow2 := fightDetails.Find(".b-fight-details__text").Eq(1).Find(".b-fight-details__text-item")

	var details string
	if fightDetailsRow2.Length() == 0 {
		// if the fight is a finish it will have the finishing method detail
		fightDetailsRow2 = fightDetails.Find(".b-fight-details__text").Eq(1)
		details = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(fightDetailsRow2.Text()), "Details:"))
	} else {
		// if the fight is not a finish it will have the judges scorecards
		details = strings.TrimSpace(fightDetailsRow2.Text())
	}

	// these whitespaces are going to drive me insane
	re := regexp.MustCompile(`\s+`)
	details = re.ReplaceAllString(details, " ")

	fmt.Printf("Details: %s\n\n", details)

	// TOTALS TABLE
	// ~~~~~~~~~~~~~

	totalsSection := fightDetails.Find(".b-fight-details__section").Eq(1)
	if totalsSection.Length() == 0 {
		fmt.Print("[No fight statistics found]\n\n")
	}

	// totalsTable will always be the first table on the page if present
	totalsTable := totalsSection.Find("table[style] tbody tr")
	if totalsSection.Length() > 0 && totalsTable.Length() == 0 {
		log.Fatal("failed to find totalsTable when totalsSection is found")
	}

	totalsTable.Each(func(i int, tr *goquery.Selection) {
		// there is only one row 'tr' so need to loop throught the 'td' children or columns
		td := tr.ChildrenFiltered("td")

		// loop through children
		td.Each(func(i int, td *goquery.Selection) {
			tableText := td.Find("p")

			// in each column, the first index of 'p' will be p1 and second will be p2
			p1Text := strings.TrimSpace(tableText.Eq(0).Text())
			p2Text := strings.TrimSpace(tableText.Eq(1).Text())

			switch i {
			case 1:
				fmt.Println("[ TOTALS ]")
				fmt.Printf("P1 KD: %s\n", p1Text)
				fmt.Printf("P2 KD: %s\n", p2Text)
			case 2:
				fmt.Printf("P1 Sig. Str.: %s\n", p1Text)
				fmt.Printf("P2 Sig. Str.: %s\n", p2Text)
			case 3:
				fmt.Printf("P1 Sig. Str. Perc: %s\n", p1Text)
				fmt.Printf("P2 Sig. Str. Perc: %s\n", p2Text)
			case 4:
				fmt.Printf("P1 Total Str.: %s\n", p1Text)
				fmt.Printf("P2 Total Str.: %s\n", p2Text)
			case 5:
				fmt.Printf("P1 TD: %s\n", p1Text)
				fmt.Printf("P2 TD: %s\n", p2Text)
			case 6:
				fmt.Printf("P1 TD Perc: %s\n", p1Text)
				fmt.Printf("P2 TD Perc: %s\n", p2Text)
			case 7:
				fmt.Printf("P1 Sub. Att.: %s\n", p1Text)
				fmt.Printf("P2 Sub. Att.: %s\n", p2Text)
			case 8:
				fmt.Printf("P1 Rev.: %s\n", p1Text)
				fmt.Printf("P2 Rev.: %s\n", p2Text)
			case 9:
				fmt.Printf("P1 Ctrl: %s\n", p1Text)
				fmt.Printf("P2 Ctrl: %s\n\n", p2Text)
			}
		})
	})

	// SIG STRIKES TABLE
	// ~~~~~~~~~~~~~~~~~~

	// since there are multiple tables on the page i need to filter by the one which contains the header 'Head' for head strikes
	sigStrikesTable := fightDetails.Find("table[style]").FilterFunction(func(i int, s *goquery.Selection) bool {
		theadText := strings.TrimSpace(s.Find("thead").Text())
		return strings.Contains(theadText, "Head")
	}).First()

	// the rows containing the data will be inside the sigStrikesTable that was filtered
	sigStrikesRows := sigStrikesTable.Find("tbody tr")

	sigStrikesRows.Each(func(i int, tr *goquery.Selection) {
		// there is only one row 'tr' so need to loop throught the 'td' children or columns
		td := tr.ChildrenFiltered("td")

		// loop through children
		td.Each(func(i int, td *goquery.Selection) {
			tableText := td.Find("p")

			// in each column, the first index of 'p' will be p1 and second will be p2
			p1Text := strings.TrimSpace(tableText.Eq(0).Text())
			p2Text := strings.TrimSpace(tableText.Eq(1).Text())

			switch i {
			// can start with the head strikes since i already have sig. strike and sig. strike %
			case 3:
				fmt.Println("[ Significant Strikes ]")
				fmt.Printf("P1 Head: %s\n", p1Text)
				fmt.Printf("P2 Head: %s\n", p2Text)
			case 4:
				fmt.Printf("P1 Body: %s\n", p1Text)
				fmt.Printf("P2 Body: %s\n", p2Text)
			case 5:
				fmt.Printf("P1 Leg: %s\n", p1Text)
				fmt.Printf("P2 Leg: %s\n", p2Text)
			case 6:
				fmt.Printf("P1 Distance: %s\n", p1Text)
				fmt.Printf("P2 Distance: %s\n", p2Text)
			case 7:
				fmt.Printf("P1 Clinch: %s\n", p1Text)
				fmt.Printf("P2 Clinch: %s\n", p2Text)
			case 8:
				fmt.Printf("P1 Ground: %s\n", p1Text)
				fmt.Printf("P2 Ground: %s\n\n", p2Text)
			}
		})
	})

	return nil
}

// COMPLETED EVENT DETAILS
// ~~~~~~~~~~~~~~~~~~~~~

func CollectEventDetails(eventLink string, reqReferer string, client *http.Client) error {
	l, err := url.Parse(eventLink)
	if err != nil {
		log.Fatalf("failed to parse fight url: %v", err)
	}
	eventID := path.Base(l.Path)

	request, err := http.NewRequest("GET", eventLink, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for event: %v", err)
	}

	request.Header.Add("referer", reqReferer)
	request.Header.Add("host", Host)
	request.Header.Add("User-Agent", UserAgent)

	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to submit request for the event: %v", err)
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

	page := doc.Find(".l-page__container")

	eventName := strings.TrimSpace(page.Find(".b-content__title").First().Text())

	detailsList := page.Find(".b-fight-details div ul").First()

	listItems := detailsList.Find(".b-list__box-list-item")

	eventDate := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(listItems.Eq(0).Text()), "Date:"))
	eventLocation := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(listItems.Eq(1).Text()), "Location:"))

	fmt.Println("[ Event Details ]")
	fmt.Printf("Event Name: %s | Event Link: %s | EventID: %s\n", eventName, eventLink, eventID)
	fmt.Printf("Date: %s\n", eventDate)
	fmt.Printf("Location: %s\n\n", eventLocation)

	return nil
}

// UPCOMING EVENT DATA
// ~~~~~~~~~~~~~~~~~~~~~

func CollectUpcomingEventData(client *http.Client) error {
	eventUpcomingLink := "http://ufcstats.com/statistics/events/upcoming?page=all"

	requestEvent, err := http.NewRequest("GET", eventUpcomingLink, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for event: %v", err)
	}

	requestEvent.Header.Add("referer", "http://ufcstats.com/statistics/events/upcoming")
	requestEvent.Header.Add("host", Host)
	requestEvent.Header.Add("User-Agent", UserAgent)

	resp, err := client.Do(requestEvent)
	if err != nil {
		return fmt.Errorf("failed to submit request for event: %v", err)
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

	page := doc.Find(".b-statistics__sub-inner")

	events := page.Find("table.b-statistics__table-events tbody tr")
	if page.Length() == 0 {
		log.Fatal("failed to find completed events table")
	}

	events.Each(func(i int, tr *goquery.Selection) {
		//skip first column (is an empty row)
		if i == 0 {
			return
		}

		td := tr.ChildrenFiltered("td")

		link, _ := td.Eq(0).Find("a").Attr("href")

		fmt.Printf("Event Link: %s\n", link)

	})

	return nil
}
