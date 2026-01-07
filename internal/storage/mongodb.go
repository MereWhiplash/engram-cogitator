package storage

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB implements Storage using MongoDB with Atlas Vector Search
type MongoDB struct {
	client    *mongo.Client
	db        *mongo.Database
	memories  *mongo.Collection
	idCounter int64
}

// memoryDoc is the MongoDB document structure
type memoryDoc struct {
	ID           int64     `bson:"_id"`
	Type         string    `bson:"type"`
	Area         string    `bson:"area"`
	Content      string    `bson:"content"`
	Rationale    string    `bson:"rationale,omitempty"`
	IsValid      bool      `bson:"is_valid"`
	SupersededBy *int64    `bson:"superseded_by,omitempty"`
	CreatedAt    time.Time `bson:"created_at"`
	Author       struct {
		Name  string `bson:"name"`
		Email string `bson:"email"`
	} `bson:"author"`
	Repo      string    `bson:"repo"`
	Embedding []float32 `bson:"embedding"`
}

// NewMongoDB creates a new MongoDB storage
func NewMongoDB(ctx context.Context, uri, database string) (*MongoDB, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	db := client.Database(database)
	memories := db.Collection("memories")

	m := &MongoDB{
		client:   client,
		db:       db,
		memories: memories,
	}

	if err := m.initIndexes(ctx); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	// Initialize ID counter from max existing ID
	if err := m.initIDCounter(ctx); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to init id counter: %w", err)
	}

	return m, nil
}

func (m *MongoDB) initIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "type", Value: 1}}},
		{Keys: bson.D{{Key: "area", Value: 1}}},
		{Keys: bson.D{{Key: "is_valid", Value: 1}}},
		{Keys: bson.D{{Key: "repo", Value: 1}}},
		{Keys: bson.D{{Key: "author.email", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}

	_, err := m.memories.Indexes().CreateMany(ctx, indexes)
	return err
}

func (m *MongoDB) initIDCounter(ctx context.Context) error {
	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	var doc memoryDoc
	err := m.memories.FindOne(ctx, bson.D{}, opts).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		m.idCounter = 0
		return nil
	}
	if err != nil {
		return err
	}
	m.idCounter = doc.ID
	return nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

func (m *MongoDB) Add(ctx context.Context, mem Memory, embedding []float32) (*Memory, error) {
	id := atomic.AddInt64(&m.idCounter, 1)
	now := time.Now()

	doc := memoryDoc{
		ID:        id,
		Type:      string(mem.Type),
		Area:      mem.Area,
		Content:   mem.Content,
		Rationale: mem.Rationale,
		IsValid:   true,
		CreatedAt: now,
		Repo:      mem.Repo,
		Embedding: embedding,
	}
	doc.Author.Name = mem.AuthorName
	doc.Author.Email = mem.AuthorEmail

	_, err := m.memories.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert memory: %w", err)
	}

	return &Memory{
		ID:          id,
		Type:        mem.Type,
		Area:        mem.Area,
		Content:     mem.Content,
		Rationale:   mem.Rationale,
		IsValid:     true,
		CreatedAt:   now,
		AuthorName:  mem.AuthorName,
		AuthorEmail: mem.AuthorEmail,
		Repo:        mem.Repo,
	}, nil
}

func (m *MongoDB) Search(ctx context.Context, embedding []float32, opts SearchOpts) ([]Memory, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}

	// Build filter
	filter := bson.D{{Key: "is_valid", Value: true}}
	if opts.Type != "" {
		filter = append(filter, bson.E{Key: "type", Value: string(opts.Type)})
	}
	if opts.Area != "" {
		filter = append(filter, bson.E{Key: "area", Value: opts.Area})
	}
	if opts.Repo != "" {
		filter = append(filter, bson.E{Key: "repo", Value: opts.Repo})
	}

	// Atlas Vector Search pipeline
	// Note: This requires an Atlas Vector Search index named "embedding_index"
	// For non-Atlas deployments, falls back to regular query (no vector search)
	pipeline := mongo.Pipeline{
		{{Key: "$vectorSearch", Value: bson.D{
			{Key: "index", Value: "embedding_index"},
			{Key: "path", Value: "embedding"},
			{Key: "queryVector", Value: embedding},
			{Key: "numCandidates", Value: limit * 10},
			{Key: "limit", Value: limit},
			{Key: "filter", Value: filter},
		}}},
	}

	cursor, err := m.memories.Aggregate(ctx, pipeline)
	if err != nil {
		// Fallback to regular query if vector search not available
		return m.listFallback(ctx, opts)
	}
	defer cursor.Close(ctx)

	return m.cursorToMemories(ctx, cursor)
}

func (m *MongoDB) listFallback(ctx context.Context, opts SearchOpts) ([]Memory, error) {
	listOpts := ListOpts{
		Limit: opts.Limit,
		Type:  opts.Type,
		Area:  opts.Area,
		Repo:  opts.Repo,
	}
	return m.List(ctx, listOpts)
}

func (m *MongoDB) List(ctx context.Context, opts ListOpts) ([]Memory, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	filter := bson.D{}
	if !opts.IncludeInvalid {
		filter = append(filter, bson.E{Key: "is_valid", Value: true})
	}
	if opts.Type != "" {
		filter = append(filter, bson.E{Key: "type", Value: string(opts.Type)})
	}
	if opts.Area != "" {
		filter = append(filter, bson.E{Key: "area", Value: opts.Area})
	}
	if opts.Repo != "" {
		filter = append(filter, bson.E{Key: "repo", Value: opts.Repo})
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := m.memories.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	return m.cursorToMemories(ctx, cursor)
}

func (m *MongoDB) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "is_valid", Value: false}}}}
	if supersededBy != nil {
		update = bson.D{{Key: "$set", Value: bson.D{
			{Key: "is_valid", Value: false},
			{Key: "superseded_by", Value: *supersededBy},
		}}}
	}

	result, err := m.memories.UpdateOne(ctx, bson.D{{Key: "_id", Value: id}}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("memory with id %d not found", id)
	}

	return nil
}

func (m *MongoDB) cursorToMemories(ctx context.Context, cursor *mongo.Cursor) ([]Memory, error) {
	var memories []Memory
	for cursor.Next(ctx) {
		var doc memoryDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		memories = append(memories, Memory{
			ID:           doc.ID,
			Type:         MemoryType(doc.Type),
			Area:         doc.Area,
			Content:      doc.Content,
			Rationale:    doc.Rationale,
			IsValid:      doc.IsValid,
			SupersededBy: doc.SupersededBy,
			CreatedAt:    doc.CreatedAt,
			AuthorName:   doc.Author.Name,
			AuthorEmail:  doc.Author.Email,
			Repo:         doc.Repo,
		})
	}

	return memories, cursor.Err()
}
