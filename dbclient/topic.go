package dbclient

// CreateTopic inserts a new topic
func (client *DBClient) CreateTopic(name string) (int64, error) {
	stmt, err := client.Db.Prepare("INSERT INTO topics (name) VALUES (?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(name)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetTopic retrieves a topic by ID
func (client *DBClient) GetTopic(id int) (*Topic, error) {
	row := client.Db.QueryRow("SELECT id, name FROM topics WHERE id = ?", id)
	topic := &Topic{}
	err := row.Scan(&topic.ID, &topic.Name)
	if err != nil {
		return nil, err
	}
	return topic, nil
}

// UpdateTopic updates an existing topic
func (client *DBClient) UpdateTopic(id int, name string) error {
	stmt, err := client.Db.Prepare("UPDATE topics SET name = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, id)
	return err
}

// DeleteTopic removes a topic
func (client *DBClient) DeleteTopic(id int) error {
	stmt, err := client.Db.Prepare("DELETE FROM topics WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	return err
}
