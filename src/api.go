package main

import (
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type CreatePersonRequest struct {
	Key    string   `json:"id" binding:"required"`
	Topics []string `json:"topics"`
}

type CreateConnectionRequest map[string]int

type SendMessageRequest struct {
	Text          string   `json:"text"`
	Topics        []string `json:"topics"`
	FromPersonKey string   `json:"from_person_id"`
	MinTrustLevel int      `json:"min_trust_level" binding:"required,min=1,max=10"`
}

func createPeople(c *gin.Context) {
	var createPersonRequest CreatePersonRequest
	if err := c.ShouldBindJSON(&createPersonRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := databaseContainer.personsCollection
	person := Person(createPersonRequest)

	exists, err := collection.DocumentExists(nil, person.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if exists {
		_, err = collection.UpdateDocument(nil, person.Key, person)
	} else {
		_, err = collection.CreateDocument(nil, person)
	}

	if err == nil {
		c.Status(http.StatusCreated)
	} else if driver.IsPreconditionFailed(err) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func createTrustConnections(c *gin.Context) {
	var personTrust PersonTrust
	var personExists bool
	var firstErr error
	var err error

	var createConnectionRequest CreateConnectionRequest
	fromPersonKey := c.Param("id")

	if err = c.ShouldBindJSON(&createConnectionRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	personsCollection := databaseContainer.personsCollection

	personExists, err = personsCollection.DocumentExists(nil, fromPersonKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if personExists == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "From Person not exists " + fromPersonKey})
		return
	}

	for toPersonKey, trustLevel := range createConnectionRequest {
		if fromPersonKey == toPersonKey {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": fmt.Sprintf("From Person and To Person should be different %s - %s", fromPersonKey, toPersonKey),
			})
			return
		}

		personExists, err = personsCollection.DocumentExists(nil, toPersonKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if personExists == false {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "To Person not exists " + toPersonKey})
			return
		}

		if trustLevel < 1 || trustLevel > 10 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": fmt.Sprintf("Wrong Trust level %s %d", toPersonKey, trustLevel),
			})
			return
		}
	}

	trustCollection := databaseContainer.personsTrustCollection

	for toPersonKey, trustLevel := range createConnectionRequest {
		personTrust = PersonTrust{
			fromPersonKey + "-" + toPersonKey,
			PersonsCollections + "/" + fromPersonKey,
			PersonsCollections + "/" + toPersonKey,
			trustLevel,
		}

		personExists, err = databaseContainer.personsTrustCollection.DocumentExists(nil, personTrust.Key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if personExists {
			_, err = trustCollection.UpdateDocument(nil, personTrust.Key, personTrust)
		} else {
			_, err = trustCollection.CreateDocument(nil, personTrust)
		}

		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if firstErr == nil {
		c.Status(http.StatusCreated)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": firstErr.Error()})
	}
}

func sendMessages(c *gin.Context) {
	var sendMessageRequest SendMessageRequest
	if err := c.ShouldBindJSON(&sendMessageRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	personExists, err := databaseContainer.personsCollection.DocumentExists(nil, sendMessageRequest.FromPersonKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if personExists == false {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "From Person not exists " + sendMessageRequest.FromPersonKey})
		return
	}

	query := `FOR person, personTrust, path
		IN 1..99 OUTBOUND @fromPerson
		GRAPH @trustNetworkGraph
		PRUNE !IS_NULL(personTrust)
			AND ( personTrust.trustLevel < @minTrustLevel OR @messageTopics ANY NOT IN person.topics )
		OPTIONS {
			"order": "bfs",
			"uniqueVertices": "global"
		}
		FILTER @messageTopics ALL IN person.topics
		FILTER @minTrustLevel <= personTrust.trustLevel
		RETURN path.vertices[*]._key`

	bindVars := map[string]interface{}{
		"trustNetworkGraph": databaseContainer.trustNetworkGraph.Name(),
		"fromPerson":        databaseContainer.personsCollection.Name() + "/" + sendMessageRequest.FromPersonKey,
		"minTrustLevel":     sendMessageRequest.MinTrustLevel,
		"messageTopics":     sendMessageRequest.Topics,
	}

	cursor, err := databaseContainer.db.Query(nil, query, bindVars)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer cursor.Close()

	response := map[string][]string{}

	var personsInPath []string
	for ok := cursor.HasMore(); ok; ok = cursor.HasMore() {
		_, err = cursor.ReadDocument(nil, &personsInPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var previousPersonKey string
		for _, personKey := range personsInPath {
			if "" != previousPersonKey && !contains(response[previousPersonKey], personKey) {
				response[previousPersonKey] = append(response[previousPersonKey], personKey)
			}
			previousPersonKey = personKey
		}
	}

	c.JSON(http.StatusCreated, response)
}

func findPath(c *gin.Context) {
	var sendMessageRequest SendMessageRequest
	if err := c.ShouldBindJSON(&sendMessageRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	personExists, err := databaseContainer.personsCollection.DocumentExists(nil, sendMessageRequest.FromPersonKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if personExists == false {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "From Person not exists " + sendMessageRequest.FromPersonKey})
		return
	}

	query := `FOR person, personTrust, path
		IN 1..99 OUTBOUND @fromPerson
		GRAPH @trustNetworkGraph
		// !IS_NULL(personTrust) -> not rootNode
		PRUNE !IS_NULL(personTrust) AND (
		    personTrust.trustLevel < @minTrustLevel OR @messageTopics ALL IN person.topics
		)
		OPTIONS {
			"order": "bfs",
			"uniqueVertices": "path"
		}
		FILTER @minTrustLevel <= personTrust.trustLevel
		FILTER @messageTopics ALL IN path.vertices[-1].topics
		LIMIT 1
		RETURN SLICE(path.vertices, 1)[*]._key`

	bindVars := map[string]interface{}{
		"trustNetworkGraph": databaseContainer.trustNetworkGraph.Name(),
		"fromPerson":        databaseContainer.personsCollection.Name() + "/" + sendMessageRequest.FromPersonKey,
		"minTrustLevel":     sendMessageRequest.MinTrustLevel,
		"messageTopics":     sendMessageRequest.Topics,
	}

	cursor, err := databaseContainer.db.Query(nil, query, bindVars)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer cursor.Close()

	var response struct {
		From string   `json:"from"`
		Path []string `json:"path"`
	}

	response.From = sendMessageRequest.FromPersonKey
	response.Path = []string{}

	if cursor.HasMore() {
		_, err = cursor.ReadDocument(nil, &response.Path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, response)
}

func runApiServer() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.POST("/api/people", createPeople)

	r.POST("/api/people/:id/trust_connections", createTrustConnections)

	r.POST("/api/messages", sendMessages)
	r.POST("/api/path", findPath)

	r.GET("/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "health")
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}

func contains(list []string, search string) bool {
	for i, _ := range list {
		// reverse search
		if list[len(list)-1-i] == search {
			return true
		}
	}
	return false
}
