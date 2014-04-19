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
	session, err := mgo.Dial("localhost:27017")
	if(err != nil) {
		log.Printf("Error connecting to MongoDB: %s", err)
		return "", err
	}
	defer session.Close()
	err = session.DB("mailhog").C("messages").Insert(msg)
	if err != nil {
		log.Printf("Error inserting message: %s", err)
		return "", err
	}
	return msg.Id, nil
}

func Load(id string) (*data.Message, error) {
	session, err := mgo.Dial("localhost:27017")
	if(err != nil) {
		log.Printf("Error connecting to MongoDB: %s", err)
		return nil, err
	}
	defer session.Close()
	result := &data.Message{}
	err = session.DB("mailhog").C("messages").Find(bson.M{"id": id}).One(&result)
	if err != nil {
		log.Printf("Error loading message: %s", err)
		return nil, err
	}
	return result, nil;
}
