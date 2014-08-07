package models

import (
	"regexp"
	"strings"
)

type Results struct {
	Id     int64  `json:"-"`
	Header string `json:"-"`
	Body   string `json:"-"`
}

func GetResults(queryString string) ([]Article, error) {
	Logger.Println("search....", queryString)
	a := []Article{}
	err := db.Limit(10).Where("body LIKE ?", "%"+queryString+"%").Find(&a).Error
	if err != nil {
		Logger.Println("Error", err)
		return a, err
	}
	//let's make pretty urls from title
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		Logger.Println("Fuck")
	}
	for i, _ := range a {
		prettyurl := reg.ReplaceAllString(a[i].Header, "-")
		prettyurl = strings.ToLower(strings.Trim(prettyurl, "-"))
		a[i].PrettyURL = prettyurl
	}
	return a, err
}
