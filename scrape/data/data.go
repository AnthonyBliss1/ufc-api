package data

import "time"

type Fighter struct {
	ID          string      `bson:"_id" json:"id"`                    // unique id given to the fighter
	Name        string      `bson:"name" json:"name"`                 // full name of the fighter
	Height      string      `bson:"height" json:"height"`             // height of the fighter (i.e 5'11)
	WeightLB    string      `bson:"weight_lb" json:"weight_lb"`       // weight of the fighter in pounds (i.e 155 lbs)
	ReachIN     string      `bson:"reach_in" json:"reach_in"`         // reach distance of the fighter in cm
	Stance      string      `bson:"stance" json:"stance"`             // stance style of the fighter
	DOB         time.Time   `bson:"dob" json:"dob"`                   // date of the birth
	CareerStats CareerStats `bson:"career_stats" json:"career_stats"` // see 'CareerStats' type below
}

type CareerStats struct {
	SLpM   float32 `bson:"slpm" json:"slpm"`       // significant strikes landed per minute
	StrAcc float32 `bson:"str_acc" json:"str_acc"` // significant striking accuracy
	SApM   float32 `bson:"sapm" json:"sapm"`       // significant strikes absorbed per minute
	StrDef float32 `bson:"str_def" json:"str_def"` // significant strike defence (the % of opponents strikes that did not land)
	TdAvg  float32 `bson:"td_avg" json:"td_avg"`   // average takedowns landed per 15 minutes
	TdAcc  float32 `bson:"td_acc" json:"td_acc"`   // takedown accuracy
	TdDef  float32 `bson:"td_def" json:"td_def"`   // takedown defense (the % of opponents TD attemtps that did not land)
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
	EndTime      time.Time    `bson:"end_time" json:"end_time"`           // ending time of the last round of the fight
	Referee      string       `bson:"referee" json:"referee"`             // referee for the given fight
	Participants []FightStats `bson:"participants" json:"participants"`   // slice of fight statistics (for both fighters) for the given fight
}

type FightStats struct {
	FighterID   string    `bson:"fighter_id" json:"fighter_id"`     // id of a specific fighter in the fight
	FighterName string    `bson:"fighter_name" json:"fighter_name"` // name of a specific fighter in the fight
	Outcome     string    `bson:"outcome" json:"outcome"`           // 'W', 'L', or 'D' for a specific fighter
	KD          int       `bson:"kd" json:"kd"`                     // number of knockdowns in the fight for a specific fighter
	SigStrL     int       `bson:"sig_str_l" json:"sig_str_l"`       // number of significant strikes landed in the fight for a specific fighter
	SigStrA     int       `bson:"sig_str_a" json:"sig_str_a"`       // number of significant strikes attempted in the fight for a specific fighter
	SigStrPerc  string    `bson:"sig_str_perc" json:"sig_str_perc"` // the percentage of significant strikes in the fight for a specific fighter
	TotalStrL   int       `bson:"total_str_l" json:"total_str_l"`   // number of total strikes landed in the fight for a specific fighter
	TotalStrA   int       `bson:"total_str_a" json:"total_str_a"`   // number of total strikes attempted in the fight for a specific fighter
	TdL         int       `bson:"td_l" json:"td_l"`                 // number of takedowns landed in the fight for a specific fighter
	TdA         int       `bson:"td_a" json:"td_a"`                 // number of takedowns attempted in the fight for a specific fighter
	Sub         int       `bson:"sub" json:"sub"`                   // number of sub attempts in the fight for a specific fighter
	Rev         int       `bson:"rev" json:"rev"`                   // number of reversals in the fight for a specific fighter
	Ctrl        time.Time `bson:"ctrl" json:"ctrl"`                 // total amount of control time for a specific fighter
	HeadL       int       `bson:"head_l" json:"head_l"`             // number of head strikes landed in the fight for a specific figher
	HeadA       int       `bson:"head_a" json:"head_a"`             // number of head strikes attempted in the fight for a specific figher
	BodyL       int       `bson:"body_l" json:"body_l"`             // number of body strikes landed in the fight for a specific fighter
	BodyA       int       `bson:"body_a" json:"body_a"`             // number of body strikes attempted in the fight for a specific fighter
	LegL        int       `bson:"leg_l" json:"leg_l"`               // number of leg strikes landed in the fight for a specific fighter
	LegA        int       `bson:"leg_a" json:"leg_a"`               // number of leg strikes attempted in the fight for a specific fighter
	DistanceL   int       `bson:"distance_l" json:"distance_l"`     // number of distance strikes landed in the fight for a specific fighter
	DistanceA   int       `bson:"distance_a" json:"distance_a"`     // number of distance strikes attempted in the fight for a specific fighter
	ClinchL     int       `bson:"clinch_l" json:"clinch_l"`         // number of clinch strikes landed in the fight for a specific fighter
	ClinchA     int       `bson:"clinch_a" json:"clinch_a"`         // number of clinch strikes attempted in the fight for a specific fighter
}

// this will feed a /Fighters endpoint
type Fighters struct {
	Item []Fighter `json:"fighters"`
}

// this will feed a /Fights endpoint
type Fights struct {
	Item []Fight `json:"fights"`
}

// this will feed an /Events endpoint
type Events struct {
	Item []Events `json:"events"`
}
