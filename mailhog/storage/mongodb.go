package storage

import (
	"log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
    "github.com/ian-kent/MailHog/mailhog/data"
    "github.com/ian-kent/MailHog/mailhog"
)

func Store(c *mailhog.Config, m *data.SMTPMessage) (string, error) {
	msg := data.ParseSMTPMessage(c, m)
	session, err := mgo.Dial(c.MongoUri)
	if(err != nil) {
		log.Printf("Error connecting to MongoDB: %s", err)
		return "", err
	}
	defer session.Close()
	err = session.DB(c.MongoDb).C(c.MongoColl).Insert(msg)
	if err != nil {
		log.Printf("Error inserting message: %s", err)
		return "", err
	}
	return msg.Id, nil
}

func List(c *mailhog.Config, start int, limit int) (*data.Messages, error) {
	session, err := mgo.Dial(c.MongoUri)
	if(err != nil) {
		log.Printf("Error connecting to MongoDB: %s", err)
		return nil, err
	}
	defer session.Close()
	messages := &data.Messages{}
	err = session.DB(c.MongoDb).C(c.MongoColl).Find(bson.M{}).Skip(start).Limit(limit).All(messages)
	if err != nil {
		log.Printf("Error loading messages: %s", err)
		return nil, err
	}
	return messages, nil;
}

func DeleteOne(c *mailhog.Config, id string) error {
	session, err := mgo.Dial(c.MongoUri)
	if(err != nil) {
		log.Printf("Error connecting to MongoDB: %s", err)
		return err
	}
	defer session.Close()
	_, err = session.DB(c.MongoDb).C(c.MongoColl).RemoveAll(bson.M{"id": id})
	return err
}

func DeleteAll(c *mailhog.Config) error {
	session, err := mgo.Dial(c.MongoUri)
	if(err != nil) {
		log.Printf("Error connecting to MongoDB: %s", err)
		return err
	}
	defer session.Close()
	_, err = session.DB(c.MongoDb).C(c.MongoColl).RemoveAll(bson.M{})
	return err
}

func Load(c *mailhog.Config, id string) (*data.Message, error) {
	session, err := mgo.Dial(c.MongoUri)
	if(err != nil) {
		log.Printf("Error connecting to MongoDB: %s", err)
		return nil, err
	}
	defer session.Close()
	result := &data.Message{}
	err = session.DB(c.MongoDb).C(c.MongoColl).Find(bson.M{"id": id}).One(&result)
	if err != nil {
		log.Printf("Error loading message: %s", err)
		return nil, err
	}
	return result, nil;
}
