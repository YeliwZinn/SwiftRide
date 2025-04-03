package db

import (
	"context"
	"fmt"
	//"log"
	"time"
	"uber-clone/config"
	//"uber-clone/models"

	//"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
     
func InitMongoDB() {
    uri := config.MustGetEnv("MONGODB_URI")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        panic(fmt.Sprintf("Failed to connect to MongoDB: %v", err))
    }

    Client = client
    fmt.Println("✅ Connected to MongoDB!")
}

func GetCollection(name string) *mongo.Collection {
    return Client.Database(config.GetEnv("DB_NAME", "uber_clone")).Collection(name)
}

