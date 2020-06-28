package storage

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/doctolib/MailHog/generated/queries"
	"github.com/doctolib/MailHog/pkg/data"
)

// PostgreSQL represents PostgreSQL backed storage backend
type PostgreSQL struct {
	Pool *pgxpool.Pool
}

// CreatePostgreSQL creates a PostgreSQL backed storage backend
func CreatePostgreSQL(uri string) *PostgreSQL {
	log.Printf("Connecting to PostgreSQL: %s\n", uri)
	pool, err := pgxpool.Connect(context.TODO(), uri)
	if err != nil {
		log.Printf("Error connecting to PostgreSQL: %s", err)
		// Do not fallback on in memory storage
		os.Exit(1)
		return nil
	}
	if schema, err := queries.Asset("queries/postgresql-schema.sql"); err != nil {
		panic(err)
	} else {
		log.Println("Creating or updating PostgreSQL schema.")
		pool.Exec(context.Background(), string(schema))
	}

	return &PostgreSQL{
		Pool: pool,
	}
}

// Store stores a message in PostgreSQL and returns its storage ID
func (pg *PostgreSQL) Store(m *data.Message) (string, error) {
	// log.Printf("Store message")
	conn, err := pg.Pool.Acquire(context.TODO())
	if err != nil {
		return "", err
	}
	defer conn.Release()
	_, error := conn.Exec(context.TODO(), "INSERT INTO messages(message) VALUES ($1)", m)
	if error != nil {
		log.Printf("Insert error %v", error)
		return "", err
	}
	return string(m.ID), nil
}

// Count returns the number of stored messages
func (pg *PostgreSQL) Count() int {
	// log.Printf("Count")
	conn, err := pg.Pool.Acquire(context.TODO())
	if err != nil {
		return -1
	}
	defer conn.Release()
	var n int
	error := conn.QueryRow(context.TODO(), "SELECT count(*) FROM messages").Scan(&n)
	if error != nil {
		log.Printf("Count error %v", error)
		return -1
	}
	return n
}

// Search finds messages matching the query
func (pg *PostgreSQL) Search(kind, query string, start, limit int) (*data.Messages, int, error) {
	// log.Printf("Searching messages %s %s %d %d", kind, query, start, limit)
	conn, err := pg.Pool.Acquire(context.TODO())
	if err != nil {
		return nil, 0, err
	}
	defer conn.Release()
	var field string
	switch kind {
	case SearchKindTo:
		field = "message->'Raw'->>'To'"
	case SearchKindFrom:
		field = "message->'Raw'->>'From'"
	case SearchKindContaining:
		field = "message->'Raw'->>'Data'"
	}
	sqlQuery := "SELECT message FROM messages WHERE to_tsvector('english', " + field + ") @@ to_tsquery('english', $1) ORDER BY message->'Created' DESC LIMIT $2 OFFSET $3"
	log.Printf("Query: %s", sqlQuery)
	rows, err := conn.Query(context.TODO(), sqlQuery, query, limit, start)
	if err != nil {
		log.Printf("Search error: %v", err)
		return nil, 0, err
	}
	defer rows.Close()
	messages := make([]data.Message, 0)
	for rows.Next() {
		var message data.Message
		err = rows.Scan(&message)
		if err != nil {
			log.Printf("Error %v", err)
			return nil, 0, err
		}
		messages = append(messages, message)
	}
	msgs := data.Messages(messages)
	var count int
	error := conn.QueryRow(context.TODO(), "SELECT count(*) FROM messages WHERE to_tsvector('english', "+field+") @@ to_tsquery('english', $1)", query).Scan(&count)
	if error != nil {
		log.Printf("Count error %v", error)
		return nil, 0, err
	}
	log.Printf("Query result: %d", count)
	return &msgs, count, nil
}

// List returns a list of messages by index
func (pg *PostgreSQL) List(start int, limit int) (*data.Messages, error) {
	// log.Printf("Listing messages %d %d", start, limit)
	conn, err := pg.Pool.Acquire(context.TODO())
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(context.TODO(), "SELECT message FROM messages ORDER BY message->'Created' DESC LIMIT $1 OFFSET $2", limit, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]data.Message, 0)
	for rows.Next() {
		var message data.Message
		err = rows.Scan(&message)
		if err != nil {
			log.Printf("Error %v", err)
			return nil, err
		}
		messages = append(messages, message)
	}
	msgs := data.Messages(messages)
	return &msgs, nil
}

// DeleteOne deletes an individual message by storage ID
func (pg *PostgreSQL) DeleteOne(id string) error {
	// log.Printf("Delete %v", id)
	conn, err := pg.Pool.Acquire(context.TODO())
	if err != nil {
		return err
	}
	defer conn.Release()
	_, error := conn.Exec(context.TODO(), "DELETE FROM messages WHERE message->>'ID' = $1", id)
	if error != nil {
		log.Printf("Delete error %v", error)
		return err
	}
	return nil
}

// DeleteAll deletes all messages stored in PostgreSQL
func (pg *PostgreSQL) DeleteAll() error {
	// log.Printf("Delete all")
	conn, err := pg.Pool.Acquire(context.TODO())
	if err != nil {
		return err
	}
	defer conn.Release()
	_, error := conn.Exec(context.TODO(), "DELETE FROM messages")
	if error != nil {
		log.Printf("Delete error %v", error)
		return err
	}
	return nil
}

// Load loads an individual message by storage ID
func (pg *PostgreSQL) Load(id string) (*data.Message, error) {
	// log.Printf("Get %v", id)
	conn, err := pg.Pool.Acquire(context.TODO())
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	var message data.Message
	error := conn.QueryRow(context.TODO(), "SELECT message FROM messages WHERE message->>'ID' = $1", id).Scan(&message)
	if error != nil {
		log.Printf("Get error %v", error)
		return nil, err
	}
	return &message, nil
}
