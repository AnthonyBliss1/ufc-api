package data

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// creating maps to load scraping data (avoiding dupes) then converted to slice structs for laoding into mongo db
type (
	FighterMap       map[string]*Fighter
	EventMap         map[string]*Event
	FightMap         map[string]*Fight
	UpcomingEventMap map[string]*UpcomingEvent
	UpcomingFightMap map[string]*UpcomingFight
)

// creating 'IDable' interface which is any type that implements these methods (used for BatchLoad())
type IDable interface {
	GetID() string
	SetID(string)
}

type Fighter struct {
	ID            string      `bson:"_id" json:"id"`                                // unique id given to the fighter
	Name          string      `bson:"name" json:"name"`                             // full name of the fighter
	CurrentRecord string      `bson:"current_record" json:"current_record"`         // the current record of the fighter as of date of collection
	Nickname      string      `bson:"nickname,omitempty" json:"nickname,omitempty"` // nickname of the fighter
	Height        string      `bson:"height" json:"height"`                         // height of the fighter (i.e 5'11)
	WeightLB      string      `bson:"weight_lb" json:"weight_lb"`                   // weight of the fighter in pounds (i.e 155 lbs)
	ReachIN       string      `bson:"reach_in" json:"reach_in"`                     // reach distance of the fighter in in
	Stance        string      `bson:"stance,omitempty" json:"stance,omitempty"`     // stance style of the fighter
	DOB           *time.Time  `bson:"dob,omitempty" json:"dob,omitempty"`           // date of the birth
	CareerStats   CareerStats `bson:"career_stats" json:"career_stats"`             // see 'CareerStats' type below
}

type CareerStats struct {
	SLpM   float32 `bson:"slpm" json:"slpm"`       // significant strikes landed per minute
	StrAcc string  `bson:"str_acc" json:"str_acc"` // significant striking accuracy
	SApM   float32 `bson:"sapm" json:"sapm"`       // significant strikes absorbed per minute
	StrDef string  `bson:"str_def" json:"str_def"` // significant strike defence (the % of opponents strikes that did not land)
	TdAvg  float32 `bson:"td_avg" json:"td_avg"`   // average takedowns landed per 15 minutes
	TdAcc  string  `bson:"td_acc" json:"td_acc"`   // takedown accuracy
	TdDef  string  `bson:"td_def" json:"td_def"`   // takedown defense (the % of opponents TD attemtps that did not land)
	SubAvg float32 `bson:"sub_avg" json:"sub_avg"` // average submissions attempted per 15 minutes
}

type Event struct {
	ID       string    `bson:"_id" json:"id"`            // unique id given to the event
	Name     string    `bson:"name" json:"name"`         // name of the event
	Date     time.Time `bson:"date" json:"date"`         // date of the event
	Location string    `bson:"location" json:"location"` // location of the event
}

type Fight struct {
	ID           string       `bson:"_id" json:"id"`                      // unique id given to the fight
	EventID      string       `bson:"event_id" json:"event_id"`           // id of the event
	FightDetail  string       `bson:"fight_detail" json:"fight_detail"`   // weight class of the given fight sometimes indicates if its a title fight
	Method       string       `bson:"method" json:"method"`               // winning method of the fight (not for a specific fighter)
	MethodDetail string       `bson:"method_detail" json:"method_detail"` // details of the winning method for the given fight
	Round        int          `bson:"round" json:"round"`                 // ending round of the fight
	EndTime      string       `bson:"end_time" json:"end_time"`           // ending time of the last round of the fight
	TimeFormat   string       `bson:"time_format" json:"time_format"`     // time format of the fight ie 5 rounds of 5 minutes
	Referee      string       `bson:"referee" json:"referee"`             // referee for the given fight
	Participants []FightStats `bson:"participants" json:"participants"`   // slice of fight statistics (for both fighters) for the given fight
}

type FightStats struct {
	FighterID   string `bson:"fighter_id" json:"fighter_id"`                   // id of a specific fighter in the fight
	FighterName string `bson:"fighter_name" json:"fighter_name"`               // name of a specific fighter in the fight
	Outcome     string `bson:"outcome" json:"outcome"`                         // 'W', 'L', or 'D' for a specific fighter
	KD          int    `bson:"kd" json:"kd"`                                   // number of knockdowns in the fight for a specific fighter
	SigStrL     int    `bson:"sig_str_landed" json:"sig_str_landed"`           // number of significant strikes landed in the fight for a specific fighter
	SigStrA     int    `bson:"sig_str_attempted" json:"sig_str_attempted"`     // number of significant strikes attempted in the fight for a specific fighter
	SigStrPerc  string `bson:"sig_str_perc" json:"sig_str_perc"`               // the percentage of significant strikes landed in the fight for a specific fighter
	TotalStrL   int    `bson:"total_str_landed" json:"total_str_landed"`       // number of total strikes landed in the fight for a specific fighter
	TotalStrA   int    `bson:"total_str_attempted" json:"total_str_attempted"` // number of total strikes attempted in the fight for a specific fighter
	TdL         int    `bson:"td_landed" json:"td_landed"`                     // number of takedowns landed in the fight for a specific fighter
	TdA         int    `bson:"td_attempted" json:"td_attempted"`               // number of takedowns attempted in the fight for a specific fighter
	TdPerc      string `bson:"td_perc" json:"td_perc"`                         // the percentage of takedowns landed in the fight for a specific fighter
	Sub         int    `bson:"sub" json:"sub"`                                 // number of sub attempts in the fight for a specific fighter
	Rev         int    `bson:"rev" json:"rev"`                                 // number of reversals in the fight for a specific fighter
	Ctrl        string `bson:"ctrl" json:"ctrl"`                               // total amount of control time for a specific fighter
	HeadL       int    `bson:"head_landed" json:"head_landed"`                 // number of head strikes landed in the fight for a specific figher
	HeadA       int    `bson:"head_attempted" json:"head_attempted"`           // number of head strikes attempted in the fight for a specific figher
	BodyL       int    `bson:"body_landed" json:"body_landed"`                 // number of body strikes landed in the fight for a specific fighter
	BodyA       int    `bson:"body_attempted" json:"body_attempted"`           // number of body strikes attempted in the fight for a specific fighter
	LegL        int    `bson:"leg_landed" json:"leg_landed"`                   // number of leg strikes landed in the fight for a specific fighter
	LegA        int    `bson:"leg_attempted" json:"leg_attempted"`             // number of leg strikes attempted in the fight for a specific fighter
	DistanceL   int    `bson:"distance_landed" json:"distance_landed"`         // number of distance strikes landed in the fight for a specific fighter
	DistanceA   int    `bson:"distance_attempted" json:"distance_attempted"`   // number of distance strikes attempted in the fight for a specific fighter
	ClinchL     int    `bson:"clinch_landed" json:"clinch_landed"`             // number of clinch strikes landed in the fight for a specific fighter
	ClinchA     int    `bson:"clinch_attempted" json:"clinch_attempted"`       // number of clinch strikes attempted in the fight for a specific fighter
	GroundL     int    `bson:"ground_landed" json:"ground_landed"`             // number of ground strikes attempted in the fight for a specific fighter
	GroundA     int    `bson:"ground_attempted" json:"ground_attempted"`       // number of ground strikes attempted in the fight for a specific fighter
}

type UpcomingEvent struct {
	ID       string    `bson:"_id" json:"id"`            // unique id given to the upcoming event
	Name     string    `bson:"name" json:"name"`         // name of the upcoming event
	Date     time.Time `bson:"date" json:"date"`         // date of the upcoming event
	Location string    `bson:"location" json:"location"` // location of the upcoming event
}

type UpcomingFight struct {
	ID              string    `bson:"_id" json:"id"`                              // unique id given to the matchups
	UpcomingEventID string    `bson:"upcoming_event_id" json:"upcoming_event_id"` // id of the upcoming event
	Participants    []Fighter `bson:"tale_of_the_tape" json:"tale_of_the_tape"`   // all fighter stats for both fighters
}

// this will feed a /Fighters endpoint
type Fighters struct {
	Items []Fighter `bson:"fighters" json:"fighters"`
}

// this will feed a /Fights endpoint
type Fights struct {
	Items []Fight `bson:"fights" json:"fights"`
}

// this will feed an /Events endpoint
type Events struct {
	Items []Event `bson:"events" json:"events"`
}

// this will feed /UpcomingEvents endpoint
type UpcomingEvents struct {
	Item []UpcomingEvent `bson:"upcoming_events" json:"upcoming_events"`
}

// this will feed /UpcomingFights endpoint
type UpcomingFights struct {
	Item []UpcomingFight `bson:"upcoming_fights" json:"upcoming_fights"`
}

// defining methods to make struct types 'IDable'
func (f *Fighter) GetID() string   { return f.ID }
func (f *Fighter) SetID(id string) { f.ID = id }

func (e *Event) GetID() string   { return e.ID }
func (e *Event) SetID(id string) { e.ID = id }

func (ft *Fight) GetID() string   { return ft.ID }
func (ft *Fight) SetID(id string) { ft.ID = id }

func (ue *UpcomingEvent) GetID() string   { return ue.ID }
func (ue *UpcomingEvent) SetID(id string) { ue.ID = id }

func (uf *UpcomingFight) GetID() string   { return uf.ID }
func (uf *UpcomingFight) SetID(id string) { uf.ID = id }

// upload data into mongodb collection for either FighterMap, EventMap, or FightMap with values of 'IDable'. default batch size is 1000
func BatchLoad[T IDable](ctx context.Context, coll *mongo.Collection, m map[string]T, batchSize int) error {
	if len(m) == 0 {
		fmt.Printf("[%s data is empty, now exiting...]\n", coll.Name())
		return nil
	}

	// if dealing with any of the upcoming collections, drop first to clear the data
	if coll.Name() == "upcomingEvents" || coll.Name() == "upcomingFights" {
		if err := coll.Drop(ctx); err != nil {
			log.Fatalf("cannot drop [%s]: %v", coll.Name(), err)
		} else {
			fmt.Printf("[%s dropped successfully...]\n", coll.Name())
		}
	}

	//setting default batch size
	if batchSize <= 0 {
		batchSize = 1000
	}

	batch := make([]mongo.WriteModel, 0, batchSize)

	// looping through each struct in the map
	for k, v := range m {
		v.SetID(k)

		// define batch
		batch = append(batch, mongo.NewReplaceOneModel().
			SetFilter(bson.M{"_id": v.GetID()}).
			SetReplacement(v).
			SetUpsert(true))

		// batch upload when batch exceeds defined batch size
		if len(batch) >= batchSize {
			if _, err := coll.BulkWrite(ctx, batch, options.BulkWrite().SetOrdered(false)); err != nil {
				return fmt.Errorf("bulk write failed: %v", err)
			}
			batch = batch[:0]
		}
	}

	// final batch upload if it does not exceed batch size
	if len(batch) > 0 {
		if _, err := coll.BulkWrite(ctx, batch, options.BulkWrite().SetOrdered(false)); err != nil {
			return fmt.Errorf("final bulk write failed: %v", err)
		}
	}

	fmt.Printf("✅ [%s data loaded successfully!]\n", coll.Name())

	return nil
}
