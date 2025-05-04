package DB

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Client *mongo.Client
}
//mongodb+srv://vansh106:vansh106@cluster0.ouilz.mongodb.net/

func NewDatabase(connectionString string) (*Database, error) {
	clientOptions := options.Client().ApplyURI(connectionString)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")

	return &Database{
		Client: client,
	}, nil
}

func (db *Database) InsertKeyValue(collectionName, key string, value interface{}) error {
	collection := db.Client.Database("bb").Collection(collectionName)
	_, err := collection.InsertOne(context.Background(), bson.M{key: value})
	if err != nil {
		fmt.Println("Error inserting key-value pair:", err)
		return err
	}
	fmt.Println("inserted key-value!")
	return nil
}

func (db *Database) GetValueByKey(collectionName, key string) (interface{}, error) {
	collection := db.Client.Database("bb").Collection(collectionName)
	var result map[string]interface{}
	filter := bson.D{{key, bson.D{{"$exists", true}}}}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		log.Println("Error retrieving value by key:", err)
		return nil, err
	}
	fmt.Println(result[key])
	return result[key], nil
}

func (db *Database) KeyExists(collectionName, key string)(bool) {
	collection := db.Client.Database("bb").Collection(collectionName)
	var result map[string]interface{}
	// filter := map[string]interface{}{
	// 	key: bson.M{"$exists": true},
	// }
	err := collection.FindOne(context.Background(), map[string]interface{}{key: map[string]interface{}{"$exists": true}}).Decode(&result)
	if err != nil {
		fmt.Println("[FIRST LOGIN]Error retrieving value by key:", err)
		return false
	}
	return true
}

func (db *Database) UpdateValueByKey(collectionName, key string, value interface{}) error {
	collection := db.Client.Database("bb").Collection(collectionName)
	// filter := bson.M{key: key}
	// update := bson.M{"$push": bson.M{key: value}}

	filter := bson.D{{key, bson.D{{"$exists", true}}}}
    update := bson.D{{"$push", bson.D{{key, value}}}}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Println("Error updating value by key:", err)
		return err
	}
	fmt.Println("updated key-value!")
	return nil
}

