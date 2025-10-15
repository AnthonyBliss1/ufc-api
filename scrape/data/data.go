package data

type Fighter struct {
	Id     string // unique id given to the fighter
	Name   string // full name of the fighter
	Height string // height of the fighter (i.e 5'11)
	Weight string // weight of the fighter in pounds (i.e 155 lbs)
	Reach  string // reach distance of the fighter in cm
	Stance string // stance style of the fighter
	DOB    string // date of the birth

	CareerStats CareerStats // see 'CareerStats' type below

	Fights []Fight // slice of fights on the fighter's record
}

type CareerStats struct {
	SLpM   float32 // significant strikes landed per minute
	StrAcc float32 // significant striking accuracy
	SApM   float32 // significant strikes absorbed per minute
	StrDef float32 // significant strike defence (the % of opponents strikes that did not land)
	TdAvg  float32 // average takedowns landed per 15 minutes
	TdAcc  float32 // takedown accuracy
	TdDef  float32 // takedown defense (the % of opponents TD attemtps that did not land)
	SubAvg float32 // average submissions attempted per 15 minutes
}

type Fight struct {
	Id         string       // unique id given to the fight
	FightStats []FightStats // slice of fight statistics (for both fighters) for the given fight

	WeightClass   string // weight class of the given fight
	Method        string // winning method of the fight (not for a specific fighter)
	MethodDetails string // details of the winning method for the given fight
	Round         int    // ending round of the fight
	Time          string // ending time of the last round of the fight
	Referee       string // referee for the given fight

	Event Event // of the event type which only contains a handful of data (maybe... we'll see)
}

type FightStats struct {
	Fighter Fighter
	KD      int    // number of knockdowns in the fight for a specific fighter
	Str     int    // number of strikes in the fight for a specific fighter
	Td      int    // number of takedowns in the fight for a specific fighter
	Sub     int    // number of sub attempts in the fight for a specific fighter
	Outcome string // 'W' or 'L' for a specific fighter
}

type Event struct {
	Id       string  // unique id given to the event
	Name     string  // name of the event
	Date     string  // date of the event
	Location string  // location of the event
	Fights   []Fight // slice of fights in a specific event
}

// this will feed a /Fighters endpoint
type Fighters struct {
	Fighters []Fighter
}

// this will feed a /Fights endpoint
type Fights struct {
	Fights []Fight
}

// this will feed an /Events endpoint
type Events struct {
	Events []Events
}
