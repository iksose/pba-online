package models

import (
	"errors"
	"log"
	"os"

	"github.com/coopernurse/gorp"
	"github.com/jinzhu/gorm"
	// "github.com/iksose/phishClone/config"
	// _ "github.com/mattn/go-sqlite3"
	_ "database/sql"
	_ "github.com/go-sql-driver/mysql"
	// "code.google.com/p/go.crypto/bcrypt"
)

var Conn *gorp.DbMap
var db gorm.DB
var err error
var ErrUsernameTaken = errors.New("username already taken")
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

const (
	CAMPAIGN_IN_PROGRESS string = "In progress"
	CAMPAIGN_QUEUED      string = "Queued"
	CAMPAIGN_COMPLETE    string = "Completed"
	EVENT_SENT           string = "Email Sent"
	EVENT_OPENED         string = "Email Opened"
	EVENT_CLICKED        string = "Clicked Link"
	STATUS_SUCCESS       string = "Success"
	STATUS_UNKNOWN       string = "Unknown"
	ERROR                string = "Error"
)

// Flash is used to hold flash information for use in templates.
type Flash struct {
	Type    string
	Message string
}

type Response struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// Setup initializes the Conn object
// It also populates the Gophish Config object
func Setup() error {
	db, err = gorm.Open("mysql", "root:root@unix(/tmp/mysql.sock)/test?parseTime=True")
	db.LogMode(false)
	db.SetLogger(Logger)
	if err != nil {
		Logger.Println(err)
		return err
	}
	// If the file already exists, delete it and recreate it
	// _, err = os.Stat(config.Conf.DBPath)
	// if err != nil {
	// 	Logger.Printf("Database not found... creating db at %s\n", config.Conf.DBPath)
	// 	h, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	// 	db.CreateTable(User{})
	// 	db.CreateTable(Target{})
	// 	db.CreateTable(Result{})
	// 	db.CreateTable(Group{})
	// 	db.CreateTable(GroupTarget{})
	// 	db.CreateTable(Template{})
	// 	db.CreateTable(Attachment{})
	// 	db.CreateTable(SMTP{})
	// 	db.CreateTable(Event{})
	// 	db.CreateTable(Campaign{})
	// 	//Create the default user
	// 	init_user := User{
	// 		Username: "jon",
	// 		Hash:     string(h), //gophish
	// 		ApiKey:   "12345678901234567890123456789012",
	// 	}
	// 	err = db.Save(&init_user).Error
	// 	if err != nil {
	// 		Logger.Println(err)
	// 	}
	// }
	return nil
}
