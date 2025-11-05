package db

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	// Fights
	_, _ = db.Collection("fights").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "event_id", Value: 1}, {Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "participants.fighter_id", Value: 1}, {Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "method", Value: 1}, {Key: "_id", Value: 1}}},
		// optional text index for q
		// {Keys: bson.D{{Key: "fight_detail", Value: "text"}, {Key: "method", Value: "text"}, {Key: "method_detail", Value: "text"}, {Key: "referee", Value: "text"}}},
	})

	// Fighters
	_, _ = db.Collection("fighters").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "stance", Value: 1}}},
		{Keys: bson.D{{Key: "career_stats.slpm", Value: 1}}},
	})

	// Events
	_, _ = db.Collection("events").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "date", Value: -1}, {Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "location", Value: 1}}},
	})

	// Upcoming
	_, _ = db.Collection("upcomingEvents").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "date", Value: -1}, {Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "name", Value: 1}}},
	})
	_, _ = db.Collection("upcomingFights").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "upcoming_event_id", Value: 1}, {Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "tale_of_the_tape._id", Value: 1}, {Key: "_id", Value: 1}}},  // if embedded fighters have _id
		{Keys: bson.D{{Key: "tale_of_the_tape.name", Value: 1}, {Key: "_id", Value: 1}}}, // helps name filters a bit
	})

	return nil
}
