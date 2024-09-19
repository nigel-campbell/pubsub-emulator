package dbclient

// CreateSubscription inserts a new subscription
func (client *DBClient) CreateSubscription(topicID int, endpoint string) (int64, error) {
	stmt, err := client.Db.Prepare("INSERT INTO subscriptions (topic_id, endpoint) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(topicID, endpoint)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetSubscription retrieves a subscription by ID
func (client *DBClient) GetSubscription(id int) (*Subscription, error) {
	row := client.Db.QueryRow("SELECT id, topic_id, endpoint FROM subscriptions WHERE id = ?", id)
	sub := &Subscription{}
	err := row.Scan(&sub.ID, &sub.TopicID, &sub.Endpoint)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// UpdateSubscription updates an existing subscription
func (client *DBClient) UpdateSubscription(id int, endpoint string) error {
	stmt, err := client.Db.Prepare("UPDATE subscriptions SET endpoint = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(endpoint, id)
	return err
}

// DeleteSubscription removes a subscription
func (client *DBClient) DeleteSubscription(id int) error {
	stmt, err := client.Db.Prepare("DELETE FROM subscriptions WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	return err
}
