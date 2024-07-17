package user
import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientInstance *mongo.Client
var clientInstanceError error
var mongoOnce sync.Once

const (
	CONNECTIONSTRING = "mongodb+srv://adioolamilekan225:<LDHaDXIWS9xL1BK7>@photo.iapgayi.mongodb.net/?retryWrites=true&w=majority&appName=PHOTO"
	DB               = "PHOTO"
	USERS            = "users"
)

// GetMongoClient returns a MongoDB client instance
func GetMongoClient() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		clientOptions := options.Client().ApplyURI(CONNECTIONSTRING)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Printf("Failed to connect to MongoDB: %v", err)
			clientInstanceError = err
			return
		}

		err = client.Ping(ctx, nil)
		if err != nil {
			log.Printf("Failed to ping MongoDB: %v", err)
			clientInstanceError = err
			return
		}

		log.Println("Successfully connected to MongoDB")
		clientInstance = client
	})

	return clientInstance, clientInstanceError
}