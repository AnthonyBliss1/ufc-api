package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/anthonybliss1/ufc-api/scrape/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// should dynamically collect this. not sure if this will affect status of request
const UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0"

// may not be REQUIRED as headers for request but good to have to avoid potential issues
const Referer = "http://ufcstats.com/statistics/fighters"
const Host = "ufcstats.com"

var (
	fighterMap       = make(data.FighterMap)
	fightMap         = make(data.FightMap)
	eventMap         = make(data.EventMap)
	upcomingEventMap = make(data.UpcomingEventMap)
	upcomingFightMap = make(data.UpcomingFightMap)
)

// query param 'char=' letters to iterate through
var Letters = []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i",
	"j", "k", "l", "m", "n", "o", "p", "q", "r",
	"s", "t", "u", "v", "w", "x", "y", "z",
}

func CreateProxyClient() (*http.Client, error) {
	// store env variables to build proxy url
	oxyName := os.Getenv("OXYLABS_USERNAME")
	if oxyName == "" {
		return nil, fmt.Errorf("oxy username not set")
	}

	oxyPass := os.Getenv("OXYLABS_PASSWORD")
	if oxyPass == "" {
		return nil, fmt.Errorf("oxy pass not set")
	}

	oxyProxyHost := os.Getenv("OXYLABS_PROXY_HOST")
	if oxyProxyHost == "" {
		return nil, fmt.Errorf("oxy host not set")
	}

	oxyProxyPort := os.Getenv("OXYLABS_PROXY_PORT")
	if oxyProxyPort == "" {
		return nil, fmt.Errorf("oxy port not set")
	}

	// build the proxy_url using the env variables
	proxy_string := fmt.Sprintf("http://user-%s:%s@%s:%s", oxyName, oxyPass, oxyProxyHost, oxyProxyPort)
	proxy_url, err := url.Parse(proxy_string)
	if err != nil {
		return nil, fmt.Errorf("parsing proxy url: %v", err)
	}

	dialer := &net.Dialer{
		Timeout:   3 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// create transport using the proxy_url
	transport := &http.Transport{
		Proxy:               http.ProxyURL(proxy_url),
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: 3 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   true,
		//DisableKeepAlives:   true,
	}

	// wrap the proxy transport in the client
	client := &http.Client{Transport: transport}

	return client, nil
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

			// create the fighter struct to store the data
			fighter := data.Fighter{ID: fighterID, Name: fighterName}

			fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
			fmt.Printf("Fighter Name: %s | Fighter Link: %s | FighterID: %s\n", fighterName, link, fighterID)

			// navigate to the profile page and collect all data on the fighter
			err = CollectFighterData(&fighter, link, client)
			if err != nil {
				fmt.Printf("failed to collect data from fighter profile page: %v", err)
				return
			}

			// store the collected struct in a FighterMap type variable
			fighterMap[fighter.ID] = &fighter
		})
	}

	return nil
}

func CollectFighterData(fighter *data.Fighter, fighterProfileLink string, client *http.Client) error {
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

	currentRecord := strings.TrimSpace(page.Find(".b-content__title-record").Text())
	fighter.CurrentRecord = strings.TrimSpace(strings.TrimPrefix(currentRecord, "Record:"))
	fmt.Printf("Current Record: %s\n", fighter.CurrentRecord)

	fighter.Nickname = strings.TrimSpace(page.Find("p.b-content__Nickname").Text())
	fmt.Printf("Nickname: %s\n", fighter.Nickname)

	// PHYSCIAL AND CAREER STATISTICS COLLECTION
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	pStats.Each(func(i int, ul *goquery.Selection) {
		li := ul.ChildrenFiltered("li")

		fighter.Height = strings.TrimSpace(li.Eq(0).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Height: %s\n", fighter.Height)

		fighter.WeightLB = strings.TrimSpace(li.Eq(1).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Weight: %s\n", fighter.WeightLB)

		fighter.ReachIN = strings.TrimSpace(li.Eq(2).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Reach: %s\n", fighter.ReachIN)

		fighter.Stance = strings.TrimSpace(li.Eq(3).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Stance: %s\n", fighter.Stance)

		dob := strings.TrimSpace(li.Eq(4).Clone().Find("i").Remove().End().Text())
		if dob == "--" || dob == "" {
			fighter.DOB = nil
		} else {
			parsedDOB, err := time.Parse("Jan 2, 2006", dob)
			if err != nil {
				log.Fatalf("failed to parse D.O.B: %v", err)
			}
			fighter.DOB = &parsedDOB
		}

		if fighter.DOB != nil {
			fmt.Printf("DOB: %s\n", fighter.DOB.Format("Jan 2, 2006"))
		} else {
			fmt.Println("DOB: nil")
		}
	})

	// LEFT SIDE OF CAREER STATS
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~

	cStatsBoxLeft := fighterStats.Find("div .b-list__info-box-left").First() // contains left side of career stats box
	cStatsLeft := cStatsBoxLeft.Find("ul.b-list__box-list").First()          //narrow down to the element containing the list elements

	cStatsLeft.Each(func(i int, ul *goquery.Selection) {
		li := ul.ChildrenFiltered("li")

		slpm, err := strconv.ParseFloat(strings.TrimSpace(li.Eq(0).Clone().Find("i").Remove().End().Text()), 32)
		if err != nil {
			log.Fatalf("failed to format float SLpM: %v", err)
		}

		fighter.CareerStats.SLpM = float32(slpm)
		fmt.Printf("SLpM: %.2f\n", fighter.CareerStats.SLpM)

		fighter.CareerStats.StrAcc = strings.TrimSpace(li.Eq(1).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Str. Acc.: %s\n", fighter.CareerStats.StrAcc)

		sapm, err := strconv.ParseFloat(strings.TrimSpace(li.Eq(2).Clone().Find("i").Remove().End().Text()), 32)
		if err != nil {
			log.Fatalf("failed to format float SApM: %v", err)
		}

		fighter.CareerStats.SApM = float32(sapm)
		fmt.Printf("SApM: %.2f\n", fighter.CareerStats.SApM)

		fighter.CareerStats.StrDef = strings.TrimSpace(li.Eq(3).Clone().Find("i").Remove().End().Text())
		fmt.Printf("Str. Def.: %s\n", fighter.CareerStats.StrDef)
	})

	// RIGHT SIDE OF CAREER STATS
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~

	cStatsBoxRight := fighterStats.Find("div .b-list__info-box-right").First() // contains right side of career stats box
	cStatsRight := cStatsBoxRight.Find("ul.b-list__box-list").First()          // narrow down to the element containing the list elements

	cStatsRight.Each(func(i int, ul *goquery.Selection) {
		li := ul.ChildrenFiltered("li")

		tdAvg, err := strconv.ParseFloat(strings.TrimSpace(li.Eq(1).Clone().Find("i").Remove().End().Text()), 32)
		if err != nil {
			log.Fatalf("failed to format float TdAvg: %v", err)
		}

		fighter.CareerStats.TdAvg = float32(tdAvg)
		fmt.Printf("TD Avg.: %.2f\n", fighter.CareerStats.TdAvg)

		fighter.CareerStats.TdAcc = strings.TrimSpace(li.Eq(2).Clone().Find("i").Remove().End().Text())
		fmt.Printf("TD Acc.: %s\n", fighter.CareerStats.TdAcc)

		fighter.CareerStats.TdDef = strings.TrimSpace(li.Eq(3).Clone().Find("i").Remove().End().Text())
		fmt.Printf("TD Def.: %s\n", fighter.CareerStats.TdDef)

		subAvg, err := strconv.ParseFloat(strings.TrimSpace(li.Eq(4).Clone().Find("i").Remove().End().Text()), 32)
		if err != nil {
			log.Fatalf("failed to format float SubAvg: %v", err)
		}

		fighter.CareerStats.SubAvg = float32(subAvg)
		fmt.Printf("Sub. Avg.: %.2f\n", fighter.CareerStats.SubAvg)
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
		if !e {
			log.Fatal("cannot find fight link")
		}
		// parse out link so i can grab the base path so i can save it as the FightID
		fLink, err := url.Parse(fightLink)
		if err != nil {
			log.Fatalf("failed to parse fight url: %v", err)
		}
		fightID := path.Base(fLink.Path)

		// create the fight struct to store the data
		fight := data.Fight{ID: fightID, Participants: make([]data.FightStats, 0, 2)}

		// for each fight in the fighters profile, find the fight link, capture the FightID and stats -> create fights struct
		fmt.Printf("Fight #%d | Fight Link: %s | FightID: %s\n\n", i, fightLink, fightID)

		if err = CollectFightData(&fight, fightLink, fighterProfileLink, client); err != nil {
			log.Fatalf("failed to collect fight data: %v", err)
		}

		// store the collected struct in a FightMap type variable
		fightMap[fight.ID] = &fight
	})

	return nil
}

func CollectFightData(fight *data.Fight, fightLink string, reqReferer string, client *http.Client) error {
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

	// adding this block to identify an 'Upcoming' fight which i want to skip (will have a function to gather Upcoming fights separately)
	fightDetailsSection := fightDetails.Find(".b-fight-details__section")
	fdSectionTitle := strings.TrimSpace(fightDetailsSection.Find("a").First().Text())
	if fdSectionTitle == "Matchup" {
		fmt.Print("[ UPCOMING FIGHT ]\n\n")
		return nil
	}

	// FIGHT DATA - HEADER TABLE
	// ~~~~~~~~~~~~~~

	event := data.Event{}

	eventLink, _ := fightEvent.Attr("href")

	if err := CollectEventDetails(&event, eventLink, fightLink, client); err != nil {
		log.Fatalf("failed to collect fighter event: %v", err)
	}

	// store the collected struct in an EventMap type variable
	eventMap[event.ID] = &event

	// add EventID to Fight struct
	fight.EventID = event.ID

	participants := fightDetails.Find(".b-fight-details__person")
	p1Header := participants.Eq(0)
	p2Header := participants.Eq(1)

	// collecting the FighterIDs
	p1Name := strings.TrimSpace(p1Header.Find("a").Text())
	p1Link, _ := p1Header.Find("a").Attr("href")
	p1Url, err := url.Parse(p1Link)
	if err != nil {
		log.Fatalf("failed to parse participant ID: %v", err)
	}
	p1ID := path.Base(p1Url.Path)

	p2Name := strings.TrimSpace(p2Header.Find("a").Text())
	p2Link, _ := p2Header.Find("a").Attr("href")
	p2Url, err := url.Parse(p2Link)
	if err != nil {
		log.Fatalf("failed to parse participant ID: %v", err)
	}
	p2ID := path.Base(p2Url.Path)

	p1Outcome := strings.TrimSpace(p1Header.Find("i").Text())
	p2Outcome := strings.TrimSpace(p2Header.Find("i").Text())

	// create FightStats struct for each fighter (will be []Participants in the Fight struct)
	p1 := data.FightStats{FighterID: p1ID, FighterName: p1Name, Outcome: p1Outcome}
	p2 := data.FightStats{FighterID: p2ID, FighterName: p2Name, Outcome: p2Outcome}

	fmt.Println("[ Fight Details ]")
	fmt.Printf("P1: %s | %s - %s \nP2: %s | %s - %s\n", p1.FighterName, p1.FighterID, p1.Outcome, p2.FighterName, p2.FighterID, p2.Outcome)

	fight.FightDetail = strings.TrimSpace(fightDetails.Find(".b-fight-details__fight-head").First().Text())
	fmt.Printf("Type: %s\n", fight.FightDetail)

	fightDetailsRow1 := fightDetails.Find(".b-fight-details__text").Eq(0)

	fight.Method = strings.TrimSpace(fightDetailsRow1.Find("i[style]").Text())
	fmt.Printf("Method: %s\n", strings.TrimSpace(fight.Method))

	round := fightDetailsRow1.Find(".b-fight-details__text-item").Eq(0).Text()
	roundFrmt := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(round), "Round:"))
	fight.Round, err = strconv.Atoi(roundFrmt)
	if err != nil {
		log.Fatalf("failed to parse int Round: %v", err)
	}
	fmt.Printf("Round: %d\n", fight.Round)

	endTime := fightDetailsRow1.Find(".b-fight-details__text-item").Eq(1).Text()
	endTimeFrmt := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(endTime), "Time:"))
	fight.EndTime = endTimeFrmt
	fmt.Printf("Time: %s\n", fight.EndTime)

	rTime := fightDetailsRow1.Find(".b-fight-details__text-item").Eq(2).Text()
	rTimeFrmt := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(rTime), "Time format:"))
	fight.TimeFormat = rTimeFrmt
	fmt.Printf("Time Format: %s\n", fight.TimeFormat)

	referee := fightDetailsRow1.Find(".b-fight-details__text-item").Eq(3).Text()
	refereeFrmt := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(referee), "Referee:"))
	fight.Referee = refereeFrmt
	fmt.Printf("Referee: %s\n", fight.Referee)

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
	fight.MethodDetail = details

	fmt.Printf("Details: %s\n\n", fight.MethodDetail)

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
				p1Int, _ := strconv.Atoi(p1Text)
				p1.KD = p1Int

				p2Int, _ := strconv.Atoi(p2Text)
				p2.KD = p2Int

				fmt.Printf("P1 KD: %d\n", p1.KD)
				fmt.Printf("P2 KD: %d\n\n", p2.KD)

			case 2:
				p1l, p1a, err := extracNums(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int Sig. Str. : %v", err)
				}
				p1.SigStrL = p1l
				p1.SigStrA = p1a
				fmt.Printf("P1 Sig. Str. Landed: %d\n", p1.SigStrL)
				fmt.Printf("P1 Sig. Str. Attempted: %d\n", p1.SigStrA)

				p2l, p2a, err := extracNums(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int Sig. Str. : %v", err)
				}
				p2.SigStrL = p2l
				p2.SigStrA = p2a
				fmt.Printf("P2 Sig. Str. Landed: %d\n", p2.SigStrL)
				fmt.Printf("P2 Sig. Str. Attempted: %d\n\n", p2.SigStrA)

			case 3:
				p1.SigStrPerc = p1Text
				fmt.Printf("P1 Sig. Str. Perc: %s\n", p1.SigStrPerc)

				p2.SigStrPerc = p2Text
				fmt.Printf("P2 Sig. Str. Perc: %s\n\n", p2.SigStrPerc)

			case 4:
				p1l, p1a, err := extracNums(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int Total Str.: %v", err)
				}
				p1.TotalStrL = p1l
				p1.TotalStrA = p1a
				fmt.Printf("P1 Total Str. Landed: %d\n", p1.TotalStrL)
				fmt.Printf("P1 Total Str. Attempted: %d\n", p1.TotalStrA)

				p2l, p2a, err := extracNums(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int Total Str.: %v", err)
				}
				p2.TotalStrL = p2l
				p2.TotalStrA = p2a
				fmt.Printf("P2 Total Str. Landed: %d\n", p2.TotalStrL)
				fmt.Printf("P2 Total Str. Attempted: %d\n\n", p2.TotalStrA)

			case 5:
				p1l, p1a, err := extracNums(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int TD: %v", err)
				}
				p1.TdL = p1l
				p1.TdA = p1a
				fmt.Printf("P1 TD Landed: %d\n", p1.TdL)
				fmt.Printf("P1 TD Attempted: %d\n", p1.TdA)

				p2l, p2a, err := extracNums(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int TD: %v", err)
				}
				p2.TdL = p2l
				p2.TdA = p2a
				fmt.Printf("P2 TD Landed: %d\n", p2.TdL)
				fmt.Printf("P2 TD Attempted: %d\n\n", p2.TdA)

			case 6:
				p1.TdPerc = p1Text
				fmt.Printf("P1 TD Perc: %s\n", p1.TdPerc)

				p2.TdPerc = p2Text
				fmt.Printf("P2 TD Perc: %s\n\n", p2.TdPerc)

			case 7:
				p1.Sub, err = strconv.Atoi(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int Sub. Att.: %v", err)
				}
				fmt.Printf("P1 Sub. Att.: %d\n", p1.Sub)

				p2.Sub, err = strconv.Atoi(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int Sub. Att.: %v", err)
				}
				fmt.Printf("P2 Sub. Att.: %d\n\n", p2.Sub)

			case 8:
				p1.Rev, err = strconv.Atoi(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int Rev.: %v", err)
				}
				fmt.Printf("P1 Rev.: %d\n", p1.Rev)

				p2.Rev, err = strconv.Atoi(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int Rev.: %v", err)
				}
				fmt.Printf("P2 Rev.: %d\n\n", p2.Rev)

			case 9:
				p1.Ctrl = p1Text
				fmt.Printf("P1 Ctrl: %s\n", p1.Ctrl)

				p2.Ctrl = p2Text
				fmt.Printf("P2 Ctrl: %s\n\n", p2.Ctrl)
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

				p1l, p1a, err := extracNums(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int Head: %v", err)
				}
				p1.HeadL = p1l
				p1.HeadA = p1a
				fmt.Printf("P1 Head Landed: %d\n", p1.HeadL)
				fmt.Printf("P1 Head Attempted: %d\n", p1.HeadA)

				p2l, p2a, err := extracNums(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int Head: %v", err)
				}
				p2.HeadL = p2l
				p2.HeadA = p2a
				fmt.Printf("P2 Head Landed: %d\n", p2.HeadL)
				fmt.Printf("P2 Head Attempted: %d\n\n", p2.HeadA)

			case 4:
				p1l, p1a, err := extracNums(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int Body: %v", err)
				}
				p1.BodyL = p1l
				p1.BodyA = p1a
				fmt.Printf("P1 Body Landed: %d\n", p1.BodyL)
				fmt.Printf("P1 Body Attempted: %d\n", p1.BodyA)

				p2l, p2a, err := extracNums(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int Body: %v", err)
				}
				p2.BodyL = p2l
				p2.BodyA = p2a
				fmt.Printf("P2 Body Landed: %d\n", p2.BodyL)
				fmt.Printf("P2 Body Attempted: %d\n\n", p2.BodyA)

			case 5:
				p1l, p1a, err := extracNums(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int Leg: %v", err)
				}
				p1.LegL = p1l
				p1.LegA = p1a
				fmt.Printf("P1 Leg Landed: %d\n", p1.LegL)
				fmt.Printf("P1 Leg Attempted: %d\n", p1.LegA)

				p2l, p2a, err := extracNums(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int Leg: %v", err)
				}
				p2.LegL = p2l
				p2.LegA = p2a
				fmt.Printf("P2 Leg Landed: %d\n", p2.LegL)
				fmt.Printf("P2 Leg Attempted: %d\n\n", p2.LegA)

			case 6:
				p1l, p1a, err := extracNums(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int Distance: %v", err)
				}
				p1.DistanceL = p1l
				p1.DistanceA = p1a
				fmt.Printf("P1 Distance Landed: %d\n", p1.DistanceL)
				fmt.Printf("P1 Distance Attempted: %d\n", p1.DistanceA)

				p2l, p2a, err := extracNums(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int Distance: %v", err)
				}
				p2.DistanceL = p2l
				p2.DistanceA = p2a
				fmt.Printf("P2 Distance Landed: %d\n", p2.DistanceL)
				fmt.Printf("P2 Distance Attempted: %d\n\n", p2.DistanceA)

			case 7:
				p1l, p1a, err := extracNums(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int Clinch: %v", err)
				}
				p1.ClinchL = p1l
				p1.ClinchA = p1a
				fmt.Printf("P1 Clinch Landed: %d\n", p1.ClinchL)
				fmt.Printf("P1 Clinch Attempted: %d\n", p1.ClinchA)

				p2l, p2a, err := extracNums(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int Clinch: %v", err)
				}
				p2.ClinchL = p2l
				p2.ClinchA = p2a
				fmt.Printf("P2 Clinch Landed: %d\n", p2.ClinchL)
				fmt.Printf("P2 Clinch Attempted: %d\n\n", p2.ClinchA)

			case 8:
				p1l, p1a, err := extracNums(p1Text)
				if err != nil {
					log.Fatalf("failed to parse int Ground: %v", err)
				}
				p1.GroundL = p1l
				p1.GroundA = p1a
				fmt.Printf("P1 Ground Landed: %d\n", p1.GroundL)
				fmt.Printf("P1 Ground Attempted: %d\n", p1.GroundA)

				p2l, p2a, err := extracNums(p2Text)
				if err != nil {
					log.Fatalf("failed to parse int Ground: %v", err)
				}
				p2.GroundL = p2l
				p2.GroundA = p2a
				fmt.Printf("P2 Ground Landed: %d\n", p2.GroundL)
				fmt.Printf("P2 Ground Attempted: %d\n\n", p2.GroundA)
			}
		})
	})

	fight.Participants = append(fight.Participants, p1)
	fight.Participants = append(fight.Participants, p2)

	return nil
}

// COMPLETED EVENT DETAILS
// ~~~~~~~~~~~~~~~~~~~~~

func CollectEventDetails(event *data.Event, eventLink string, reqReferer string, client *http.Client) error {
	l, err := url.Parse(eventLink)
	if err != nil {
		log.Fatalf("failed to parse fight url: %v", err)
	}
	event.ID = path.Base(l.Path)

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

	event.Name = strings.TrimSpace(page.Find(".b-content__title").First().Text())

	detailsList := page.Find(".b-fight-details div ul").First()

	listItems := detailsList.Find(".b-list__box-list-item")

	event.Date, err = time.Parse("January 2, 2006", strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(listItems.Eq(0).Text()), "Date:")))
	if err != nil {
		log.Fatalf("failed to parse event date: %v", err)
	}

	event.Location = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(listItems.Eq(1).Text()), "Location:"))

	fmt.Println("[ Event Details ]")
	fmt.Printf("Event Name: %s | Event Link: %s | EventID: %s\n", event.Name, eventLink, event.ID)
	fmt.Printf("Date: %s\n", event.Date.Format("January 2, 2006"))
	fmt.Printf("Location: %s\n\n", event.Location)

	return nil
}

// UPCOMING EVENT DATA
// ~~~~~~~~~~~~~~~~~~~~~

// TODO complete this function to collect data on all the upcoming fights and each matchup for the upcoming events
// this does not need to be super involved. I should only need to navigate two pages - 1. the upcoming events lists and
// 2. the fights listed for the specific event
// data needed will be the event information, fighters names and ids. from there i can just query the fighter data in the database using idss
func IterateUpcomingEvents(event *data.Event, client *http.Client) error {
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

	// find table containing upcoming fights
	events := page.Find("table.b-statistics__table-events tbody tr")
	if page.Length() == 0 {
		log.Fatal("failed to find completed events table")
	}

	// looping through every upcoming fight in the table
	events.Each(func(i int, tr *goquery.Selection) {
		//skip first column (is an empty row)
		if i == 0 {
			return
		}

		// each td child is a column, first column (Eq(0)) will contain the link the event data
		td := tr.ChildrenFiltered("td")

		// find the event link
		upcomingEventLink, _ := td.Eq(0).Find("a").Attr("href")

		eventURL, err := url.Parse(upcomingEventLink)
		if err != nil {
			log.Fatalf("failed to parse upcoming event url: %v", err)
		}
		eventID := path.Base(eventURL.Path)

		fmt.Printf("Event Link: %s | %s\n", upcomingEventLink, eventID)

		// create the upcoming event struct (make sure upcomingevent map is created)
		upcomingEvent := data.UpcomingEvent{ID: eventID}

		// then navigate to the event page which contains all fights, iterate the fights
		if err := CollectUpcomingEventData(&upcomingEvent, upcomingEventLink, eventUpcomingLink, client); err != nil {
			log.Fatalf("error on fights page of upcoming event: %v", err)
		}

		// add the upcoming event struct to the upcoming event map
		upcomingEventMap[upcomingEvent.ID] = &upcomingEvent
	})

	return nil
}

// navigate to the upcoming event page and iterate the list of fights
func CollectUpcomingEventData(upcomingEvent *data.UpcomingEvent, eventLink string, referer string, client *http.Client) error {
	request, err := http.NewRequest("GET", eventLink, nil)
	if err != nil {
		return fmt.Errorf("failed to build request to fights page of upcoming event: %v", err)
	}

	request.Header.Add("referer", referer)
	request.Header.Add("host", Host)
	request.Header.Add("User-Agent", UserAgent)

	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to submit request for fights page of upcoming event: %v", err)
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

	upcomingEvent.Name = strings.TrimSpace(page.Find(".b-content__title").First().Text())

	detailsList := page.Find(".b-fight-details div ul").First()

	listItems := detailsList.Find(".b-list__box-list-item")

	upcomingEvent.Date, err = time.Parse("January 2, 2006", strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(listItems.Eq(0).Text()), "Date:")))
	if err != nil {
		log.Fatalf("failed to parse upcomingEvent date: %v", err)
	}

	upcomingEvent.Location = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(listItems.Eq(1).Text()), "Location:"))

	fmt.Println("[ Upcoming Event Details ]")
	fmt.Printf("Event Name: %s | Event Link: %s | EventID: %s\n", upcomingEvent.Name, eventLink, upcomingEvent.ID)
	fmt.Printf("Date: %s\n", upcomingEvent.Date.Format("January 2, 2006"))
	fmt.Printf("Location: %s\n\n", upcomingEvent.Location)

	// find the table that contains all the fights for the upcoming event
	fights := page.Find(".b-fight-details__table tbody tr")
	if fights.Length() == 0 {
		log.Fatal("failed to find fights for upcoming events")
	}

	// loop through each fight in the table (rows)
	fights.Each(func(i int, tr *goquery.Selection) {
		// each child will be a column in the specific row, column 5 (Eq(4)) will contain the link to the matchup
		td := tr.ChildrenFiltered("td")

		participants := td.Eq(1).Find("p")

		// collect fighter's names and parse IDs
		p1Name := strings.TrimSpace(participants.Eq(0).Text())
		p1Link, e := participants.Eq(0).Find("a").Attr("href")
		if !e {
			log.Fatal("cannot find p1 link for ID")
		}
		u1, err := url.Parse(p1Link)
		if err != nil {
			log.Fatalf("failed to parse p1Link for ID: %v", err)
		}
		p1ID := path.Base(u1.Path)

		p2Name := strings.TrimSpace(participants.Eq(1).Text())
		p2Link, e := participants.Eq(1).Find("a").Attr("href")
		if !e {
			log.Fatal("cannot find p2 link for ID")
		}
		u2, err := url.Parse(p2Link)
		if err != nil {
			log.Fatalf("failed to parse p2Link for ID: %v", err)
		}
		p2ID := path.Base(u2.Path)

		// store collected fighter names and IDs into fighter structs
		p1 := data.Fighter{ID: p1ID, Name: p1Name}
		p2 := data.Fighter{ID: p2ID, Name: p2Name}

		// parse matchup link for fight ID
		upcomingFightLink, e := td.Eq(4).Find("a").Attr("data-link")
		if !e {
			log.Fatal("cannot find upcomingFightLink for ID")
		}
		ufl, err := url.Parse(upcomingFightLink)
		if err != nil {
			log.Fatalf("failed to parse upcoming fight link for ID: %v", err)
		}
		upcomingFightID := path.Base(ufl.Path)

		// here i should collect the matchup fighter's names and ids. then need to query the db to stitch the matchup together
		upcomingFight := data.UpcomingFight{ID: upcomingFightID, UpcomingEventID: upcomingEvent.ID, Participants: make([]data.Fighter, 0, 2)}

		// add fighters to []Participants of upcoming fight
		upcomingFight.Participants = append(upcomingFight.Participants, p1, p2)

		// add upcomingFight structs to the map
		upcomingFightMap[upcomingFight.ID] = &upcomingFight

		fmt.Printf("FightID: %s\nP1: %s | %s\nP2: %s | %s\n\n",
			upcomingFight.ID, upcomingFight.Participants[0].Name, upcomingFight.Participants[0].ID,
			upcomingFight.Participants[1].Name, upcomingFight.Participants[1].ID)
	})

	return nil
}

// ADDING NEW DATA
// ~~~~~~~~~~~~~~~~~~~~~

func RunUpdate(webClient *http.Client) error {
	const eventPage = "http://ufcstats.com/statistics/events/completed?page=all"

	var newEvents = make([]string, 0, 10)

	connString := os.Getenv("MONGO_URI")
	if connString == "" {
		log.Fatal("mongodb connection string empty")
	}

	ctx := context.Background()

	client, err := mongo.Connect(options.Client().ApplyURI(connString))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("could not connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("disconnect error: %v", err)
		}
	}()

	db := client.Database("ufc")
	events := db.Collection("events")

	ev, err := getLatestEvent(ctx, events)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Println("no events from db found")
			return nil
		}
		return err
	}

	recentEventID := ev.ID

	request, err := http.NewRequest("GET", eventPage, nil)
	if err != nil {
		return fmt.Errorf("failed to build request to events page: %v", err)
	}

	request.Header.Add("host", Host)
	request.Header.Add("User-Agent", UserAgent)

	resp, err := webClient.Do(request)
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

	allEvents := page.Find("table.b-statistics__table-events tbody tr")
	if page.Length() == 0 {
		log.Fatal("failed to find completed events table")
	}

	allEvents.EachWithBreak(func(i int, tr *goquery.Selection) bool {
		//skip first column (is an empty row)
		if i <= 1 {
			return true
		}

		td := tr.ChildrenFiltered("td")

		link, _ := td.Eq(0).Find("a").Attr("href")

		u, err := url.Parse(link)
		if err != nil {
			log.Panic("cannot parse url")
		}
		eventID := path.Base(u.Path)

		if eventID == recentEventID {
			fmt.Printf("[Found Event Match in DB!]\n[New Events: %d]\n", len(newEvents))
			return false
		}
		newEvents = append(newEvents, link)
		return true
	})

	if len(newEvents) > 0 {
		fmt.Print("\n[Collecting New Data...]\n\n")
		for _, link := range newEvents {
			request, err := http.NewRequest("GET", link, nil)
			if err != nil {
				return fmt.Errorf("failed to create newEvent request: %v", err)
			}

			request.Header.Add("referer", eventPage)
			request.Header.Add("host", Host)
			request.Header.Add("User-Agent", UserAgent)

			resp, err := webClient.Do(request)
			if err != nil {
				return fmt.Errorf("failed to make newevent request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				return fmt.Errorf("request not accepted, Status Code: %d | %v", resp.StatusCode, err)
			}

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %v", err)
			}

			page := doc.Find(".l-page__container")

			fightRows := page.Find(".b-fight-details__table tbody tr")

			if fightRows.Length() == 0 {
				log.Fatal("Cannot find fight rows")
			}

			fightRows.Each(func(i int, tr *goquery.Selection) {
				td := tr.ChildrenFiltered("td")

				// grab the names of each fighter in the fight and will need to update their records
				p1Name := strings.TrimSpace(td.Eq(1).Find("p").Eq(0).Text())
				p2Name := strings.TrimSpace(td.Eq(1).Find("p").Eq(1).Text())

				fighterLinks := td.Eq(1).Find("p")
				p1Link, _ := fighterLinks.Eq(0).Find("a").Attr("href")
				p2Link, _ := fighterLinks.Eq(1).Find("a").Attr("href")

				u1, err := url.Parse(p1Link)
				if err != nil {
					log.Panic("cannot parse url")
				}
				p1ID := path.Base(u1.Path)

				u2, err := url.Parse(p2Link)
				if err != nil {
					log.Panic("cannot parse url")
				}
				p2ID := path.Base(u2.Path)

				f1 := data.Fighter{ID: p1ID, Name: p1Name}
				f2 := data.Fighter{ID: p2ID, Name: p2Name}

				fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
				fmt.Printf("Fighter Name: %s | Fighter Link: %s | FighterID: %s\n", f1.Name, p1Link, f1.ID)
				if err := CollectFighterData(&f1, p1Link, webClient); err != nil {
					log.Fatalf("failed to collect p1 data: %v", err)
				}

				fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
				fmt.Printf("Fighter Name: %s | Fighter Link: %s | FighterID: %s\n", f2.Name, p2Link, f2.ID)
				if err := CollectFighterData(&f2, p2Link, webClient); err != nil {
					log.Fatalf("failed to collect p2 data: %v", err)
				}

				fighterMap[f1.ID] = &f1
				fighterMap[f2.ID] = &f2
			})
		}
	} else {
		fmt.Print("\n[No New Events Found]\n\n")
	}

	return nil
}

// query the events collection to return the most recent event
func getLatestEvent(ctx context.Context, coll *mongo.Collection) (*data.Event, error) {
	opts := options.FindOne().
		SetSort(bson.M{"date": -1}).
		SetProjection(bson.M{
			"_id":      1,
			"name":     1,
			"date":     1,
			"location": 1,
		})

	var ev data.Event
	if err := coll.FindOne(ctx, bson.M{}, opts).Decode(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}

func extracNums(s string) (i1, i2 int, err error) {
	ofIndex := strings.Index(s, "of")

	if ofIndex == -1 {
		return 0, 0, errors.New("failed to find 2 integers")
	}

	i1, err = strconv.Atoi(s[:ofIndex-1])
	if err != nil {
		return 0, 0, err
	}

	i2, err = strconv.Atoi(s[ofIndex+3:])
	if err != nil {
		return 0, 0, err
	}

	return i1, i2, nil
}

// BATCHING / POPULATING DATA IN DB
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func RunBatches() {
	connString := os.Getenv("MONGO_URI")
	if connString == "" {
		log.Fatal("mongodb connection string empty")
	}

	ctx := context.Background()

	client, err := mongo.Connect(options.Client().ApplyURI(connString))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("could not connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("disconnect error: %v", err)
		}
	}()

	db := client.Database("ufc")

	if err := data.BatchLoad(ctx, db.Collection("fighters"), fighterMap, 1000); err != nil {
		log.Fatalf("fighters load failed: %v", err)
	}
	if err := data.BatchLoad(ctx, db.Collection("events"), eventMap, 1000); err != nil {
		log.Fatalf("events load failed: %v", err)
	}
	if err := data.BatchLoad(ctx, db.Collection("fights"), fightMap, 1000); err != nil {
		log.Fatalf("fights load failed: %v", err)
	}
	if err := data.BatchLoad(ctx, db.Collection("upcomingEvents"), upcomingEventMap, 1000); err != nil {
		log.Fatalf("upcomingEvents load failed: %v", err)
	}

	// update fighter record in upcoming fights with 'Fighter' data before loading upcomingfights
	if err := EnrichUpcomingFightsFromDB(ctx, db, upcomingFightMap); err != nil {
		log.Fatalf("failed enriching upcoming fights: %v", err)
	}

	if err := data.BatchLoad(ctx, db.Collection("upcomingFights"), upcomingFightMap, 1000); err != nil {
		log.Fatalf("upcomingFights load failed: %v", err)
	}
}

// use this to add data to the tale_of_the_tape (Fighter data) to all []Fighter entries in UpcomingFight
func EnrichUpcomingFightsFromDB(ctx context.Context, db *mongo.Database, upcomingFightMap map[string]*data.UpcomingFight) error {
	// collect unique fighter IDs referenced across all upcoming fights
	idSet := make(map[string]struct{}, 256)
	for _, uf := range upcomingFightMap {
		for _, p := range uf.Participants {
			if p.ID != "" {
				idSet[p.ID] = struct{}{}
			}
		}
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil // nothing to do
	}

	// fetch fighters in one go; projection optional (grab everything)
	cur, err := db.Collection("fighters").Find(ctx, bson.M{
		"_id": bson.M{"$in": ids},
	})
	if err != nil {
		return fmt.Errorf("fighters find failed: %w", err)
	}
	defer cur.Close(ctx)

	fByID := make(map[string]data.Fighter, len(ids))
	for cur.Next(ctx) {
		var f data.Fighter
		if err := cur.Decode(&f); err != nil {
			return fmt.Errorf("decode fighter failed: %w", err)
		}
		fByID[f.ID] = f
	}
	if err := cur.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	// rewrite the participants
	for _, uf := range upcomingFightMap {
		full := make([]data.Fighter, 0, len(uf.Participants))
		for _, p := range uf.Participants {
			if f, ok := fByID[p.ID]; ok {
				full = append(full, f)
			} else {
				// not in DB yet (new signee, name change, etc.) — keep your minimal stub
				full = append(full, p)
			}
		}
		uf.Participants = full
	}

	return nil
}
