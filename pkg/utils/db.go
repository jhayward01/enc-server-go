package utils

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DB interface {
	StoreRecord(id, record string) (err error)
	RetrieveRecord(id string) (record string, err error)
	DeleteRecord(id string) (err error)
}

type dbImpl struct {
	mongoURI string
}

type Entry struct {
	Id     string
	Record string
}

func (db *dbImpl) getRecordCollection() (coll *mongo.Collection, err error) {

	log.Println("Connecting to data store...")
	
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

	// Create Mongo client.
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(db.mongoURI))
	if err != nil {
		return nil, err
	}
	log.Println("Data store connected...")

	// Verify connectivity.
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	log.Println("Data store ping verified...")

	// Retrieve record collection.
	coll = client.Database("enc-server-go").Collection("records")
	return coll, nil
}

func (db *dbImpl) StoreRecord(id, record string) (err error) {

	// Get reference to record collection.
	coll, err := db.getRecordCollection()
	if err != nil {
		return err
	}

	// Set query parameters.
	filter := bson.D{primitive.E{Key: "id", Value: id}}
	update := bson.D{primitive.E{Key: "$set",
		Value: bson.D{primitive.E{Key: "record", Value: record}}}}
	opts := options.Update().SetUpsert(true)

	// Update record record.
	result, err := coll.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	log.Printf("Number of record entries updated: %v\n", result.ModifiedCount)
	log.Printf("Number of record entries upserted: %v\n", result.UpsertedCount)

	return nil
}

func (db *dbImpl) RetrieveRecord(id string) (record string, err error) {

	// Get reference to record collection.
	coll, err := db.getRecordCollection()
	if err != nil {
		return "", err
	}

	// Set query parameters.
	filter := bson.D{primitive.E{Key: "id", Value: id}}

	// Query record entry.
	var entry Entry
	if err = coll.FindOne(context.TODO(), filter).Decode(&entry); err != nil {
		return "", err
	}

	// Extract record.
	record = entry.Record

	return record, nil
}

func (db *dbImpl) DeleteRecord(id string) (err error) {

	// Get reference to record collection.
	coll, err := db.getRecordCollection()
	if err != nil {
		return err
	}

	filter := bson.D{primitive.E{Key: "id", Value: id}}
	opts := options.Delete().SetHint(bson.D{{Key: "_id", Value: 1}})

	result, err := coll.DeleteMany(context.TODO(), filter, opts)
	if err != nil {
		return err
	}
	log.Printf("Number of record entries deleted: %d\n", result.DeletedCount)

	return nil
}

func MakeDB(configs map[string]string) (db DB, err error) {

	// Verify required configurations.
	if ok, missing := VerifyConfigs(configs, []string{"mongoURI"}); !ok {
		err = errors.New("MakeDB missing configuration " + missing)
		return nil, err
	}

	db = &dbImpl{
		mongoURI: configs["mongoURI"],
	}
	return db, nil
}
