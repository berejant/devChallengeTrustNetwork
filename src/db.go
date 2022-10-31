package main

import (
	. "github.com/arangodb/go-driver"
	arangodbHttp "github.com/arangodb/go-driver/http"
	"log"
	"time"
)

const (
	PersonsCollections      string = "Persons"
	PersonsTrustCollections string = "PersonsTrust"
	trustNetworkGraph       string = "trustNetwork"
)

type DatabaseContainer struct {
	db                     Database
	personsCollection      Collection
	personsTrustCollection Collection
	trustNetworkGraph      Graph
}

type Person struct {
	Key    string   `json:"_key" binding:"required"`
	Topics []string `json:"topics"`
}

type PersonTrust struct {
	Key        string `json:"_key""`
	From       string `json:"_from"`
	To         string `json:"_to"`
	TrustLevel int    `json:"trustLevel"`
}

func createDatabaseContainer() DatabaseContainer {
	db := connectDatabase()

	return DatabaseContainer{
		db:                     db,
		personsCollection:      getCollection(db, PersonsCollections, createPersonsCollection),
		personsTrustCollection: getCollection(db, PersonsTrustCollections, createPersonsTrustCollection),
		trustNetworkGraph:      createTrustNetworkGraphIfNotExist(db),
	}
}

func connectDatabase() Database {
	conn, err := arangodbHttp.NewConnection(
		arangodbHttp.ConnectionConfig{
			Endpoints: []string{
				getEnv("DB_ENDPOINT", "http://localhost:8529"),
			},
		},
	)

	if err != nil {
		log.Fatalf("Failed to create HTTP connection: %v", err)
	}

	client, err := NewClient(
		ClientConfig{
			Connection: conn,
			Authentication: BasicAuthentication(
				getEnv("DB_USER", "root"),
				getEnv("DB_PASSWORD", ""),
			),
		},
	)

	if err != nil {
		log.Fatalf("Failed to create arango DB Connection connection: %v", err)
	}
	for {
		versionInfo, err := client.Version(nil)
		if err == nil {
			log.Printf("Database ready: %v \n", versionInfo)
			break
		}
		log.Printf("Database is not ready. Wait for database server be loaded...")
		time.Sleep(time.Second)
	}

	dbName := getEnv("DB_NAME", "trustNetwork")

	dbExists, err := client.DatabaseExists(nil, dbName)

	if !dbExists {
		log.Printf("Database %s is not exists. Probably first run. Do create.. \n", dbName)
		createDatabase(client, dbName)
	}

	db, err := client.Database(nil, dbName)
	if err != nil {
		log.Fatalf("Failed to open database: %s %v", dbName, err)
	}

	return db
}

func createDatabase(client Client, name string) {
	db, err := client.CreateDatabase(nil, name, nil)

	if err != nil {
		log.Fatalf("Failed to create database %s \n", err)
		return
	}

	createPersonsCollection(db)
	createPersonsTrustCollection(db)
}

func createPersonsCollection(db Database) Collection {
	opts := CreateCollectionOptions{
		Type: CollectionTypeDocument,
		Schema: &CollectionSchemaOptions{
			Level:   CollectionSchemaLevelStrict,
			Message: "The Person must contains list topics",
		},
	}

	var collection Collection
	var err error

	err = opts.Schema.LoadRule([]byte(`{
		"type": "object",
		"properties": {
			"topics": {
				"type": "array",
					"items": {
					"type": "string"
				}
			}
		},
		"required": ["topics"]
	}`))

	if err != nil {
		log.Fatalf("Failed to load schema rules for %s: %s \n", PersonsCollections, err)
	}

	collection, err = db.CreateCollection(nil, PersonsCollections, &opts)
	if err != nil {
		log.Fatalf("Failed to create Persons collection: %s \n", err)
	}

	return collection
}

func createPersonsTrustCollection(db Database) Collection {
	opts := CreateCollectionOptions{
		Type: CollectionTypeEdge,
		Schema: &CollectionSchemaOptions{
			Level:   CollectionSchemaLevelStrict,
			Message: "The Person Trust must contains trustLevel",
		},
	}

	var collection Collection
	var err error

	err = opts.Schema.LoadRule([]byte(`{
		"type": "object",
		"properties": {
			"trustLevel": {
				"type": "number",
				"maximum": 10,
				"minimum": 1
			}
		},
		"required": ["trustLevel"]
	}`))

	if err != nil {
		log.Fatalf("Failed to load schema rules for %s: %s \n", PersonsTrustCollections, err)
	}

	collection, err = db.CreateCollection(nil, PersonsTrustCollections, &opts)
	if err != nil {
		log.Fatalf("Failed to create Persons collection: %s \n", err)
	}

	return collection
}

func getCollection(db Database, name string, createFunc func(db Database) Collection) Collection {
	exist, err := db.CollectionExists(nil, name)
	if !exist {
		createFunc(db)
	}

	collection, err := db.Collection(nil, name)
	if err != nil {
		log.Fatalf("Failed to get collection %s. Try to reset DB. %s", name, err)
	}
	return collection
}

func createTrustNetworkGraphIfNotExist(db Database) Graph {
	var graph Graph
	exists, err := db.GraphExists(nil, trustNetworkGraph)
	if err != nil {
		log.Fatalf("Failed to check Graph exists %s. Try to reset DB. %s", trustNetworkGraph, err)
		return nil
	}
	if exists {
		graph, _ = db.Graph(nil, trustNetworkGraph)
		return graph
	}

	opts := CreateGraphOptions{
		EdgeDefinitions: []EdgeDefinition{
			{
				Collection: PersonsTrustCollections,
				From:       []string{PersonsCollections},
				To:         []string{PersonsCollections},
			},
		},
	}

	graph, err = db.CreateGraph(nil, trustNetworkGraph, &opts)
	if err != nil {
		log.Fatalf("Failed to create graph %s. Try to reset DB. %s", trustNetworkGraph, err)
	}

	return graph
}
