package models

import (
	"fmt"
	"strconv"
)

// reference all updates's by their id
type Update struct {
	id int64
}

func NewUpdate(userId int64, body string) (*Update, error) {
	id, err := client.Incr("update:next-id").Result()
	if err != nil { return nil, err }

	key := fmt.Sprintf("update:%d", id)

	// Pipeline of commands to send many commands in one go. Instead of many back and forths
	pipe := client.Pipeline()
	// Hash Set. set the the key's id, username, and hash to the variable values
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "user_id", userId)
	pipe.HSet(key, "body", body)
	
	// For every new update, we push the update to two hashmaps. "updates" is a global list of all update posts
	// The other list (user:id-updates) is the hasmap of updates JUST for a specific user
	pipe.LPush("updates", id)
	pipe.LPush(fmt.Sprintf("user:%d:updates", userId), id)

	// execute all the statements in the pipeline
	_, err = pipe.Exec()
	if err != nil { return nil, err }

	return &Update{ id }, nil
}

func (update *Update) GetBody() (string, error) {
	key := fmt.Sprintf("update:%d", update.id)
	return client.HGet(key, "body").Result()
}

func (update *Update) GetUser() (*User, error) {

	key := fmt.Sprintf("update:%d", update.id)
	userId, err := client.HGet(key, "user_id").Int64()
	if err != nil { return nil, err }
	return GetUserById(userId)
}


func queryUpdates(key string) ([]*Update, error) {
	updateIds , err := client.LRange(key, 0, 10).Result()
	if err != nil { return nil, err }

	updates := make([]*Update, len(updateIds))
	for i, strId := range updateIds {
		
		intId, err := strconv.Atoi(strId)

		if err != nil { 
			return nil, err
		}

		updates[i] = &Update{ int64(intId) }
	}
	return updates, nil
}


func GetAllUpdates() ([]*Update, error) {
	return queryUpdates("updates")
}

func GetUpdates(userId int64) ([]*Update, error) {

	key := fmt.Sprintf("user:%d:updates", userId)
	return queryUpdates(key)
}

func PostUpdate(userId int64, body string) error {

	_, err := NewUpdate(userId, body)
	return err
}