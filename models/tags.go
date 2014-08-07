package models

import _ "github.com/jinzhu/gorm"

type Tag struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

func GetTags() ([]Tag, error) {
	Logger.Println("Get tags")
	tags := []Tag{}
	err := db.Find(&tags).Error
	if err != nil {
		Logger.Println("Error")
		return tags, err
	}
	// for i, _ := range tags {
	// 	Logger.Println("Iterating", tags[i].Title)
	// }
	err = nil
	return tags, err
}
