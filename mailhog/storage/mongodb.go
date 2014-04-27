package storage

import (
	"log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
    "github.com/ian-kent/MailHog/mailhog/data"
    "github.com/ian-kent/MailHog/mailhog/config"
)

type MongoDB struct {
	Session *mgo.Session
	Config *config.Config
	Collection *mgo.Collection
}

func CreateMongoDB(c *config.Config) *MongoDB {
	log.Printf("Connecting to MongoDB: %s\n", c.MongoUri)
	session, err := mgo.Dial(c.MongoUri)
	if(err != nil) {
		log.Printf("Error connecting to MongoDB: %s", err)
		return nil
	}
	return &MongoDB{
		Session: session,
		Config: c,
		Collection: session.DB(c.MongoDb).C(c.MongoColl),
	}
}

func (mongo *MongoDB) Store(m *data.Message) (string, error) {	
	err := mongo.Collection.Insert(m)
	if err != nil {
		log.Printf("Error inserting message: %s", err)
		return "", err
	}
	return m.Id, nil
}

func (mongo *MongoDB) List(start int, limit int) (*data.Messages, error) {
	messages := &data.Messages{}
	err := mongo.Collection.Find(bson.M{}).Skip(start).Limit(limit).Select(bson.M{
		"id": 1,
		"_id": 1,
		"from": 1,
		"to": 1,
		"content.headers": 1,
		"content.size": 1,
		"created": 1,
	}).All(messages)
	if err != nil {
		log.Printf("Error loading messages: %s", err)
		return nil, err
	}
	return messages, nil;
}

func (mongo *MongoDB) DeleteOne(id string) error {
	_, err := mongo.Collection.RemoveAll(bson.M{"id": id})
	return err
}

func (mongo *MongoDB) DeleteAll() error {
	_, err := mongo.Collection.RemoveAll(bson.M{})
	return err
}

func (mongo *MongoDB) Load(id string) (*data.Message, error) {
	result := &data.Message{}
	err := mongo.Collection.Find(bson.M{"id": id}).One(&result)
	if err != nil {
		log.Printf("Error loading message: %s", err)
		return nil, err
	}
	return result, nil;
}
