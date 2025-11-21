# Overview
This project consists of two main parts:
* Webscraper to collect historical and future UFC data
* REST API service to integrate with the collected data

> [!NOTE]
> This assumes use of a No SQL database (preferably MongoDB)

## Scraping
The web scraping program can be run a few different ways:

```bash
./scrape --update
```
```bash
./scrape --upcoming
```
```bash
./scrape
```

### Update
The update flag will query the associated database collection for the most recent UFC Event. It will then collect any new data populated to `UfcStats.com`.
At a high level, the update process will find new events, collect data for each fight during the event, and update the associated fighter's record in the database.

### Upcoming
The upcoming flag will populate the `Upcoming Events` and `Upcoming Fights` collections in the database.
This process will navigate to the `Upcoming Events` page of the site and similarly iterate through each fight for the event.

### No Flags
Running `./scrape` with no flags will collect all available data (historical and upcoming) from UfcStats.com and store it in the database.

## REST API
### Features

* **Fighters** - search & filter athletes by name, stance, stats, and date-of-birth range
* **Fights** - query historical fight results with referee, method, and participant filters
* **Events** - past MMA event data searchable by event name, date range, and location
* **Upcoming Events & Upcoming Fights** - schedule access for future cards and matchups
* **30‑second response caching** for performance

### Endpoints

- **Fights**
  - `/fights` - List fights w/ filters
  - `/fights/search` - Search fights by keyword
  - `/fights/{id}` - Get single fight

- **Fighters**
  - `/fighters` - List fighters w/ filters
  - `/fighters/search` - Search fighters
  - `/fighters/{id}` - Get single fighter

- **Events**
  - `/events` - List past events w/ date filters
  - `/events/search` - Search events
  - `/events/{id}` - Get event details

- **Upcoming Events**
  - `/upcomingEvents` - List scheduled events
  - `/upcomingEvents/search` - Search upcoming events
  - `/upcomingEvents/{id}` - Get upcoming event

- **Upcoming Fights**
  - `/upcomingFights` - List upcoming fights
  - `/upcomingFights/{id}` - Get upcoming fight
