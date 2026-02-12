package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/dto"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/infrastructure/mongodb"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const testCollection = "test_ping"

// MongoTestHandler handles MongoDB test/debug endpoints.
type MongoTestHandler struct {
	mongo *mongodb.MongoDB
}

// NewMongoTestHandler creates a new MongoTestHandler.
func NewMongoTestHandler(m *mongodb.MongoDB) *MongoTestHandler {
	return &MongoTestHandler{mongo: m}
}

// Ping tests MongoDB connectivity.
// GET /api/v1/test/mongo/ping
func (h *MongoTestHandler) Ping(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err := h.mongo.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), fmt.Sprintf("MongoDB ping failed: %v", err))
	}

	res := dto.NewMongoPingResponse("connected", latency)
	return response.OK(adapter.NewFiberResponder(c), res)
}

// InsertTest inserts a test document into MongoDB.
// POST /api/v1/test/mongo/insert
func (h *MongoTestHandler) InsertTest(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := bson.M{
		"message":    "test from calendar-reg-main-api",
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}

	result, err := h.mongo.Database().Collection(testCollection).InsertOne(ctx, doc)
	if err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), fmt.Sprintf("Insert failed: %v", err))
	}

	return response.Created(adapter.NewFiberResponder(c), bson.M{
		"inserted_id": result.InsertedID,
		"document":    doc,
	})
}

// FindAll retrieves all test documents from MongoDB.
// GET /api/v1/test/mongo/find
func (h *MongoTestHandler) FindAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := h.mongo.Database().Collection(testCollection).Find(ctx, bson.M{})
	if err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), fmt.Sprintf("Find failed: %v", err))
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), fmt.Sprintf("Cursor decode failed: %v", err))
	}

	return response.OK(adapter.NewFiberResponder(c), bson.M{
		"count":     len(results),
		"documents": results,
	})
}

// DeleteAll deletes all test documents from MongoDB.
// @Summary Delete all test documents
// @Description Delete all documents from the test collection
// @Tags mongo
// @Accept json
// @Produce json
// @Success 200 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /test/mongo/delete [delete]
func (h *MongoTestHandler) DeleteAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.mongo.Database().Collection(testCollection).DeleteMany(ctx, bson.M{})
	if err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), fmt.Sprintf("Delete failed: %v", err))
	}

	return response.OK(adapter.NewFiberResponder(c), bson.M{
		"deleted_count": result.DeletedCount,
	})
}

// FullTest runs a full cycle: ping → insert → find → delete.
// @Summary Run full MongoDB test
// @Description Run a full test cycle: ping, insert, find, and delete
// @Tags mongo
// @Accept json
// @Produce json
// @Success 200 {object} dto.MongoTestResult
// @Failure 500 {object} interface{}
// @Router /test/mongo/full [get]
func (h *MongoTestHandler) FullTest(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res := dto.MongoTestResult{}

	// 1. Ping
	if err := h.mongo.Ping(ctx); err != nil {
		res.Ping = fmt.Sprintf("FAIL: %v", err)
		return response.InternalError(adapter.NewFiberResponder(c), fmt.Sprintf("Ping failed: %v", err))
	}
	res.Ping = "OK"

	// 2. Insert
	doc := bson.M{"message": "full-test", "created_at": time.Now().UTC().Format(time.RFC3339)}
	insertResult, err := h.mongo.Database().Collection(testCollection).InsertOne(ctx, doc)
	if err != nil {
		res.Insert = fmt.Sprintf("FAIL: %v", err)
		return response.OK(adapter.NewFiberResponder(c), res)
	}
	res.Insert = "OK"

	// 3. Find
	var found bson.M
	err = h.mongo.Database().Collection(testCollection).FindOne(ctx, bson.M{"_id": insertResult.InsertedID}).Decode(&found)
	if err != nil {
		res.Find = fmt.Sprintf("FAIL: %v", err)
		return response.OK(adapter.NewFiberResponder(c), res)
	}
	res.Find = "OK"

	// 4. Delete
	_, err = h.mongo.Database().Collection(testCollection).DeleteOne(ctx, bson.M{"_id": insertResult.InsertedID})
	if err != nil {
		res.Delete = fmt.Sprintf("FAIL: %v", err)
		return response.OK(adapter.NewFiberResponder(c), res)
	}
	res.Delete = "OK"

	return response.OK(adapter.NewFiberResponder(c), res)
}
