package models

import "database/sql"

// User represents the user model for gophish.
type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username" sql:"not null;unique"`
	Hash     string `json:"-"`
	ApiKey   string `json:"api_key" sql:"not null;unique"`
}

// GetUser returns the user that the given id corresponds to. If no user is found, an
// error is thrown.
func GetUser(id int64) (User, error) {
	u := User{}
	err := db.Where("id=?", id).First(&u).Error
	if err != nil {
		return u, err
	}
	return u, nil
}

// GetUserByAPIKey returns the user that the given API Key corresponds to. If no user is found, an
// error is thrown.
func GetUserByAPIKey(key string) (User, error) {
	u := User{}
	err := db.Where("api_key = ?", key).First(&u).Error
	if err != nil {
		return u, err
	}
	return u, nil
}

// GetUserByUsername returns the user that the given username corresponds to. If no user is found, an
// error is thrown.
func GetUserByUsername(username string) (User, error) {
	Logger.Println("Get user by username...")
	u := User{}
	err := db.Where("username = ?", username).First(&u).Error
	if err != sql.ErrNoRows {
		return u, ErrUsernameTaken
	} else if err != nil {
		return u, err
	}
	return u, nil
}

// PutUser updates the given user
func PutUser(u *User) error {
	err := db.Save(u).Error
	return err
}
