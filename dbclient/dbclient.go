package dbclient

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type DBClient struct {
	Db *sql.DB
}

// NewDBClient initializes a new SQLite client
func NewDBClient(dataSourceName string) (*DBClient, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Ensure the connection is valid
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Initialize tables if they don't exist
	err = createTables(db)
	if err != nil {
		return nil, err
	}
	return &DBClient{Db: db}, nil
}

// Close closes the database connection
func (client *DBClient) Close() error {
	return client.Db.Close()
}

// createTables creates the necessary tables if they don't exist.
func createTables(db *sql.DB) error {
	// Create topics table
	topicsTable := `
    CREATE TABLE IF NOT EXISTS topics (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL
    );`

	_, err := db.Exec(topicsTable)
	if err != nil {
		return err
	}

	// Create subscriptions table
	subscriptionsTable := `
    CREATE TABLE IF NOT EXISTS subscriptions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        topic_id INTEGER NOT NULL,
        endpoint TEXT NOT NULL,
        FOREIGN KEY (topic_id) REFERENCES topics(id)
    );`

	_, err = db.Exec(subscriptionsTable)
	if err != nil {
		return err
	}

	// Create messages table
	messagesTable := `
    CREATE TABLE IF NOT EXISTS messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        topic_id INTEGER NOT NULL,
        subscription_id INTEGER NOT NULL,
        content TEXT NOT NULL,
        FOREIGN KEY (topic_id) REFERENCES topics(id),
        FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
    );`

	_, err = db.Exec(messagesTable)
	if err != nil {
		return err
	}

	return nil
}
