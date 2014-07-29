package models

import (
    "time"
    "strings"
    "regexp"

    // "github.com/jinzhu/gorm"
)

//Articles is a struct representing a created article
type Article struct {
    ID            int64     `json:"Article_ID"`
    Header        string     `json:"header"`
    Body          string    `json:"body"`
    Date          time.Time `json:"date"`
    NewProp       string
}


// GetCampaigns returns the campaigns owned by the given user.
func GetArticles(aid int64) ([]Article, error) {
    Logger.Println("Get articles")
    a := []Article{}
    err := db.Limit(10).Find(&a).Error
    if err != nil {
        Logger.Println("Error")
        return a, err
    }
    //let's make pretty urls from title
    reg, err := regexp.Compile("[^A-Za-z0-9]+")
    if err != nil {
        Logger.Println("Fuck")
    }
    for i, _ := range a {
        // Logger.Println("Iterating", a[i].Header)
        prettyurl := reg.ReplaceAllString(a[i].Header, "-")
        prettyurl = strings.ToLower(strings.Trim(prettyurl, "-"))
        a[i].NewProp = prettyurl
    }
    return a, err;
}

func GetArticle(title string) (Article){
    Logger.Println("Get one article?", title)
    a := Article{}
    reg, err := regexp.Compile("[^A-Za-z0-9]+")
    if err != nil {
        Logger.Println("Fuck")
    }
    prettyurl := reg.ReplaceAllString(title, " ")
    prettyurl = strings.ToLower(strings.Trim(prettyurl, "-"))
    Logger.Println("Pretty?", prettyurl)
    err = db.Where("header = ?", prettyurl).Find(&a).Error
    if err != nil{
        Logger.Println("Error", err)
    }
    return a
}
