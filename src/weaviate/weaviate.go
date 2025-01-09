package weaviate

import (
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

// SDK encapsulates all Weaviate operations
type SDK struct {
	client *weaviate.Client
}

// NewSDK creates a new instance of SDK
func NewSDK(client *weaviate.Client) *SDK {
	return &SDK{
		client: client,
	}
}

// CreateSchema creates a new class schema in Weaviate
func (w *SDK) CreateSchema(ctx context.Context, className string, properties []*models.Property, vectorizer string) error {
	// Check if class already exists
	exists, err := w.classExists(ctx, className)
	if err != nil {
		return fmt.Errorf("failed to check if class exists: %v", err)
	}
	if exists {
		return fmt.Errorf("class %s already exists", className)
	}

	class := &models.Class{
		Class:      className,
		Properties: properties,
		Vectorizer: vectorizer,
	}

	err = w.client.Schema().ClassCreator().WithClass(class).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Weaviate class: %v", err)
	}

	return nil
}

// classExists checks if a class exists in the schema
func (w *SDK) classExists(ctx context.Context, className string) (bool, error) {
	schema, err := w.client.Schema().Getter().Do(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get schema: %v", err)
	}

	for _, class := range schema.Classes {
		if class.Class == className {
			return true, nil
		}
	}

	return false, nil
}

// DeleteSchema deletes a class schema from Weaviate
func (w *SDK) DeleteSchema(ctx context.Context, className string) error {
	err := w.client.Schema().ClassDeleter().WithClassName(className).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete Weaviate class: %v", err)
	}

	return nil
}

// VectorObject represents a single object with its vector and properties
type VectorObject struct {
	Vector     []float32
	Properties map[string]interface{}
}

// AddVector adds a single vector object to a class
func (w *SDK) AddVector(ctx context.Context, className string, object VectorObject) error {
	_, err := w.client.Data().Creator().
		WithClassName(className).
		WithProperties(object.Properties).
		WithVector(object.Vector).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to add vector: %v", err)
	}

	return nil
}

// BatchAddVectors adds multiple vector objects to a class in a single operation
func (w *SDK) BatchAddVectors(ctx context.Context, className string, objects []VectorObject) error {
	// Convert VectorObjects to models.Object
	objs := make([]*models.Object, len(objects))
	for i, obj := range objects {
		objs[i] = &models.Object{
			Class:      className,
			Properties: obj.Properties,
			Vector:     obj.Vector,
		}
	}

	// Batch add objects
	batcher := w.client.Batch().ObjectsBatcher()
	resp, err := batcher.WithObjects(objs...).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to batch add vectors: %v", err)
	}
	if len(resp) == 0 {
		return fmt.Errorf("batch operation returned no results")
	}

	return nil
}

// QueryConfig represents configuration for vector similarity search
type QueryConfig struct {
	Fields    []string // Fields to return in the result
	Limit     int      // Maximum number of results
	Distance  float64  // Optional distance threshold
	Certainty float64  // Optional certainty threshold (1/distance)
}

const DefaultQueryLimit = 20

// QueryResult represents a single result from vector similarity search
type QueryResult struct {
	ID         string
	Score      float64 // Distance or certainty score
	Properties map[string]interface{}
}

// QueryVectors performs vector similarity search in a class
func (w *SDK) QueryVectors(ctx context.Context, className string, vector []float32, config QueryConfig) ([]QueryResult, error) {
	// Convert string fields to GraphQL fields
	fields := make([]graphql.Field, len(config.Fields))
	for i, field := range config.Fields {
		fields[i] = graphql.Field{Name: field}
	}
	// Add _additional field for metadata
	fields = append(fields, graphql.Field{Name: "_additional { id distance certainty }"})

	// Build near vector arguments
	nearVectorBuilder := w.client.GraphQL().NearVectorArgBuilder().
		WithVector(vector)

	if config.Distance > 0 {
		nearVectorBuilder.WithDistance(float32(config.Distance))
	}
	if config.Certainty > 0 {
		nearVectorBuilder.WithCertainty(float32(config.Certainty))
	}

	if config.Limit <= 0 {
		config.Limit = DefaultQueryLimit
	}

	// Execute query
	result, err := w.client.GraphQL().Get().
		WithClassName(className).
		WithFields(fields...).
		WithNearVector(nearVectorBuilder).
		WithLimit(config.Limit).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %v", err)
	}

	// Parse results
	var queryResults []QueryResult
	if data, ok := result.Data["Get"].(map[string]interface{}); ok {
		if objects, ok := data[className].([]interface{}); ok {
			for _, obj := range objects {
				if objMap, ok := obj.(map[string]interface{}); ok {
					additional := objMap["_additional"].(map[string]interface{})

					// Create properties map excluding _additional
					properties := make(map[string]interface{})
					for k, v := range objMap {
						if k != "_additional" {
							properties[k] = v
						}
					}

					queryResults = append(queryResults, QueryResult{
						ID:         additional["id"].(string),
						Score:      additional["distance"].(float64),
						Properties: properties,
					})
				}
			}
		}
	}

	return queryResults, nil
}

// DeleteVector deletes a vector object from a class by its auto-generated ID
func (w *SDK) DeleteVector(ctx context.Context, className string, id string) error {
	err := w.client.Data().Deleter().
		WithClassName(className).
		WithID(id).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete vector: %v", err)
	}

	return nil
}
