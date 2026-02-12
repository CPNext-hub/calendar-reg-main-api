package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoDB wraps the official mongo client and a default database reference.
type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

// Connect creates a new MongoDB connection and pings to verify.
func Connect(ctx context.Context, host, dbName, user, password string) (*MongoDB, error) {
	uri := buildURI(host, dbName, user, password)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("mongodb connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongodb ping: %w", err)
	}

	log.Printf("Connected to MongoDB (%s/%s)", host, dbName)
	return &MongoDB{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

// Disconnect gracefully closes the connection.
func (m *MongoDB) Disconnect(ctx context.Context) error {
	log.Println("Disconnecting from MongoDB...")
	return m.client.Disconnect(ctx)
}

// Database returns the default *mongo.Database.
func (m *MongoDB) Database() *mongo.Database {
	return m.db
}

// Client returns the underlying *mongo.Client.
func (m *MongoDB) Client() *mongo.Client {
	return m.client
}

// Ping checks if the database is reachable.
func (m *MongoDB) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

// buildURI constructs a MongoDB connection string.
func buildURI(host, dbName, user, password string) string {
	if user != "" && password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=admin", user, password, host, dbName)
	}
	return fmt.Sprintf("mongodb://%s/%s", host, dbName)
}
