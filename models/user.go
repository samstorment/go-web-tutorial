package models

import (
	"fmt"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"github.com/go-redis/redis"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidLogin = errors.New("invalid login")
	ErrUsernameTaken = errors.New("username taken")
)

// All of this is writtem in episode 12 - probably need to rewatch
type User struct {
	// reference users by their id in the DB
	id int64
}

func NewUser(username string, hash []byte) (*User, error) {

	// check Redis to see if the username already exists in the DB. If it exists return the specific error we created
	exists, err := client.HExists("user:by-username", username).Result();
	if exists {
		return nil, ErrUsernameTaken
	}

	// The "user:next-id" key has the value of the id used for the last used. We increment that value by 1 for the new user's id. This id is then assigned as a unique id to a user
	id, err := client.Incr("user:next-id").Result()
	if err != nil { return nil, err }

	// Key is a HASH MAP of user info. We can insert values into it with this logic: key["username"] = "John". The actualy syntax is client.HSet(key, "username", "John")
	key := fmt.Sprintf("user:%d", id)

	// Pipeline of commands to send many commands to Redis in one go. Instead of many back and forths
	pipe := client.Pipeline()

	// Insert the values into the DB. Key is a hash map in Redis
	// key["id"] = id
	pipe.HSet(key, "id", id)
	// key["username"] = username
	pipe.HSet(key, "username", username)
	// key["hash"] = hash
	pipe.HSet(key, "hash", hash)
	// Insert the user's id into the "user:by-username" hashmap. Will let us look up the user's id by their username with the same hashmap logic: id = user:by-username["username"]
	pipe.HSet("user:by-username", username, id)

	// execute all FOUR statements in the pipeline
	_, err = pipe.Exec()
	if err != nil { return nil, err }

	// return a reference to a user object with the id we generated. We can use any user object with an id to access info about the user by using the id to look things up
	return &User{ id }, nil
}

// Get the id of the user object. Since id is the only field in the struct, we can just directly return the id from the struct
func (user *User) GetId() (int64, error) {
	return user.id, nil
}

// Get the username of the user object
func (user *User) GetUsername() (string, error) {
	// first, we need to use the correct key hashmap. We want the hashmap for this user, represented as user:id
	key := fmt.Sprintf("user:%d", user.id)
	// Retrieve the user's username from the DB -> return key["username"]
	return client.HGet(key, "username").Result()
}

// Get the hashed password of the user object
func (user *User) GetHash() ([]byte, error) {
	// we need to use the correct key hashmap
	key := fmt.Sprintf("user:%d", user.id)
	// Get the hashed password from the DB -> return key["hash"]
	return client.HGet(key, "hash").Bytes()
}

// Authenticate the user's password by comparing the hashed password in the DB with the password given to us upon login
func (user *User) Authenticate(password string) error {
	// get the user's hashed password from the DB
	hash, err := user.GetHash()
	if err != nil { return err }

	// Compare the hashed password to the given password. Convert given password to byte slice, because thats what bcrypt expects. 
	// If the two passwords don't match, return the correct type of error so we can tell the user their password is no good
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrInvalidLogin
	}
	// the case where there is no error
	return err
}

// Returns reference to a user object with the given id. We can use that object to perform struct methods on the user
func GetUserById(id int64) (*User, error) {
	return &User{ id }, nil
}

// Returns reference to user object with the given username.
func GetUserByUsername(username string) (*User, error) {
	// get the user's ID from Redis -> id = user:by-username[username]. Return the correct type of error if no user with that username exists
	id, err := client.HGet("user:by-username", username).Int64()
	if err == redis.Nil {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	// return a reference to the user object
	return GetUserById(id)
}


func AuthenticateUser(username, password string) (*User, error) {

	// Get a reference to the User object with the given username. Will be an error if the username doesn't exist
	user, err := GetUserByUsername(username)
	if err != nil { return nil, err }
	// Return the User object reference and the potential error from Authenticating that user. Will be an error if the password is wrong
	return user, user.Authenticate(password)
}

func RegisterUser(username, password string) error {

	// use Bcrypt's default encryption cost. The better the cost the more resistant to brute force decryption attempts
	cost := bcrypt.DefaultCost
	// generate a hashed password from the passord the user types in
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil { return err }
	// Store the new username and hashed password in the database
	_, err = NewUser(username, hash)
	return err
}
