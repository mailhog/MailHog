package storage

import (
	"github.com/mailhog/data"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

// MongoDB represents MongoDB backed storage backend
type MongoDB struct {
	Session    *mgo.Session
	Collection *mgo.Collection
}

// CreateMongoDB creates a MongoDB backed storage backend
func CreateMongoDB(uri, db, coll string, limit int) *MongoDB {
	log.Printf("Connecting to MongoDB: %s\n", uri)
	session, err := mgo.Dial(uri)
	if err != nil {
		log.Printf("Error connecting to MongoDB: %s", err)
		return nil
	}
	err = session.DB(db).C(coll).EnsureIndexKey("created")
	if err != nil {
		log.Printf("Failed creating index: %s", err)
		return nil
	}
	StorageLimit = limit
	return &MongoDB{
		Session:    session,
		Collection: session.DB(db).C(coll),
	}
}

// Store stores a message in MongoDB and returns its storage ID
func (mongo *MongoDB) Store(m *data.Message) (string, error) {
	c, _ := mongo.Collection.Count()
	// If number of documents bigger than the limit then remove the last document
	if c >= StorageLimit {
		log.Printf("Storage count (%d) bigger than limit (%d). Removing the exceeding documents.", c, StorageLimit)
		var result bson.M
		change := mgo.Change{
			Remove: true,
			ReturnNew: false,
		}
		for i := 0; i < (c - StorageLimit) + 1; i++ {
			_, err := mongo.Collection.Find(bson.M{}).Sort("created").Apply(change, &result)
			if err != nil {
				log.Printf("Error deleting messages: %s (continuing anyway)", err)
			}
		}
	}
	err := mongo.Collection.Insert(m)
	if err != nil {
					log.Printf("Error inserting message: %s", err)
					return "", err
	}
	return string(m.ID), nil
}


// Count returns the number of stored messages
func (mongo *MongoDB) Count() int {
	c, _ := mongo.Collection.Count()
	return c
}

// Search finds messages matching the query
func (mongo *MongoDB) Search(kind, query string, start, limit int) (*data.Messages, int, error) {
	messages := &data.Messages{}
	var count = 0
	var field = "raw.data"
	switch kind {
	case "to":
		field = "raw.to"
	case "from":
		field = "raw.from"
	}
	err := mongo.Collection.Find(bson.M{field: bson.RegEx{Pattern: query, Options: "i"}}).Skip(start).Limit(limit).Sort("-created").Select(bson.M{
		"id":              1,
		"_id":             1,
		"from":            1,
		"to":              1,
		"content.headers": 1,
		"content.size":    1,
		"created":         1,
		"raw":             1,
	}).All(messages)
	if err != nil {
		log.Printf("Error loading messages: %s", err)
		return nil, 0, err
	}
	count, _ = mongo.Collection.Find(bson.M{field: bson.RegEx{Pattern: query, Options: "i"}}).Count()

	return messages, count, nil
}

// List returns a list of messages by index
func (mongo *MongoDB) List(start int, limit int) (*data.Messages, error) {
	messages := &data.Messages{}
	err := mongo.Collection.Find(bson.M{}).Skip(start).Limit(limit).Sort("-created").Select(bson.M{
		"id":              1,
		"_id":             1,
		"from":            1,
		"to":              1,
		"content.headers": 1,
		"content.size":    1,
		"created":         1,
		"raw":             1,
	}).All(messages)
	if err != nil {
		log.Printf("Error loading messages: %s", err)
		return nil, err
	}
	return messages, nil
}

// DeleteOne deletes an individual message by storage ID
func (mongo *MongoDB) DeleteOne(id string) error {
	 _, err := mongo.Collection.RemoveAll(bson.M{"id": id})
	return err
}

// DeleteAll deletes all messages stored in MongoDB
func (mongo *MongoDB) DeleteAll() error {
	//_, err := mongo.Collection.RemoveAll(bson.M{})
	// Is faster to just drop the collection than delete all documents in it
	err := mongo.Collection.DropCollection()
	return err
}

// Load loads an individual message by storage ID
func (mongo *MongoDB) Load(id string) (*data.Message, error) {
	result := &data.Message{}
	err := mongo.Collection.Find(bson.M{"id": id}).One(&result)
	if err != nil {
		log.Printf("Error loading message: %s", err)
		return nil, err
	}
	return result, nil
}
