package dbclient

// CreateMessage inserts a new message
func (client *DBClient) CreateMessage(topicID, subscriptionID int, content string) (int64, error) {
	stmt, err := client.Db.Prepare("INSERT INTO messages (topic_id, subscription_id, content) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(topicID, subscriptionID, content)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetMessage retrieves a message by ID
func (client *DBClient) GetMessage(id int) (*Message, error) {
	row := client.Db.QueryRow("SELECT id, topic_id, subscription_id, content FROM messages WHERE id = ?", id)
	msg := &Message{}
	err := row.Scan(&msg.ID, &msg.TopicID, &msg.SubscriptionID, &msg.Content)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// UpdateMessage updates an existing message
func (client *DBClient) UpdateMessage(id int, content string) error {
	stmt, err := client.Db.Prepare("UPDATE messages SET content = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(content, id)
	return err
}

// DeleteMessage removes a message
func (client *DBClient) DeleteMessage(id int) error {
	stmt, err := client.Db.Prepare("DELETE FROM messages WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	return err
}
