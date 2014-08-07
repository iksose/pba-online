package models

import (
	"errors"
	"regexp"
	"strings"
	"time"
	// "html/template"

	// "github.com/jinzhu/gorm"
)

//Articles is a struct representing a created article
type Article struct {
	Id        int64
	Header    string    `json:"Header"`
	Body      string    `json:"Body"`
	Img       string    `json: "img"`
	Date      time.Time `json:"date"`
	Teaser    string    `json:"teaser"`
	PrettyURL string    `sql:"-"`
	Tags      []int64   `sql:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GetCampaigns returns the campaigns owned by the given user.
// takes an int, why. Returns 4 newest for front page.
func GetArticles(aid int64) ([]Article, error) {
	Logger.Println("Get articles")
	a := []Article{}
	err := db.Order("date desc").Limit(4).Find(&a).Error
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
		a[i].PrettyURL = prettyurl
		// Logger.Println(a[i].Body)
		// a[i].Body = template.HTML(a[i].BodyOld)
	}
	return a, err
}

func GetArticle(title string) Article {
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
	if err != nil {
		Logger.Println("Error", err)
	}
	Logger.Println("Article bod template?", a.Id, a.Img)
	return a
}

func EditArticle(article Article) error {
	Logger.Println("In add article func")
	// db.Table("articles").Where("header = ?", article.Header).Update("Body", article.Body)
	if err != nil {
		Logger.Println("Error", err)
	}
	// Logger.Println(db.Table("articles").Where("header = ?", article.Header).Update("Body", article.Body).RowsAffected)
	count := db.Table("articles").
		Where("header = ?", article.Header).
		Updates(map[string]interface{}{"Body": article.Body, "Img": article.Img,
		"Teaser": article.Teaser}).RowsAffected
	// count := db.Table("articles").Where("header = ?", article.Header).Update("Body", article.Body).RowsAffected
	if err != nil {
		Logger.Println("Fuck")
		// error = err;
		return err
	}
	if count > 0 {
		Logger.Println("Success!")
		// err = nil;
		return nil
	}
	// if err := db.Table("articles").Where("header = ?", article.Header).Updates(map[string]interface{}{"Body": article.Body, "header_img": article.Img}).Error; err != nil {
	//       Logger.Println("Fuck dude", err)
	// }
	var ErrNotFound = errors.New("update failed")
	return ErrNotFound
	// return error;
}

func PostArticle(article Article) error {
	Logger.Println("Add article", article.Tags)

	// save
	err := db.Save(&article).Error
	if err != nil {
		Logger.Println("Fuck", err)
		return err
	}
	Logger.Println("Success")
	// update tag table
	type Sample struct {
		Articles_ID int64 `json: "Articles_ID"`
		Tag_ID      int64 `json: "Tag_ID"`
	}
	for i, _ := range article.Tags {
		Logger.Println("Iterating", article.Tags[i])
		var tTag = Sample{article.Id, article.Tags[i]}
		err = db.Exec("INSERT INTO Articles_tags VALUES (?, ?)", tTag.Articles_ID, tTag.Tag_ID).Error
		if err != nil {
			Logger.Println("Fuck", err)
			return err
		}
	}
	err = nil
	return err
}

func DeleteArticle(article Article) error {
	Logger.Println("Delete article", article.Id)
	db.Delete(&article)
	if err != nil {
		Logger.Println("Fuck")
		return err
	}
	err = nil
	return err
}

func ElementsByType(category string) ([]Article, error) {
	// type Result struct {
	// 	header string `json:"Header"`
	// }
	Logger.Println("Get by category ", category)
	// article := []Article{}
	// result := []Result{}
	var id int64
	var header string
	var body string
	var date time.Time
	var img string
	var teaser string
	var prettyurl string
	var tags []int64
	var time1 time.Time
	var time2 time.Time
	results := make(map[int]Article)
	// var count int
	rows, err := db.Raw("SELECT articles.id, header, body, date, img, teaser"+
		" FROM Articles_Tags"+
		" JOIN articles"+
		" ON Articles_ID = articles.id"+
		" INNER JOIN Tags"+
		" ON Articles_Tags.Tag_ID = Tags.id"+
		" where title = ?", category).Rows()
	defer rows.Close()
	var i = 0
	for rows.Next() {
		rows.Scan(&id, &header, &body, &date, &img, &teaser)
		results[i] = Article{id, header, body, img, date, teaser, prettyurl, tags, time1, time2}
		Logger.Println("wow", img)
		i++
	}
	myArray := make([]Article, i) // length of rows.Next()
	var idx = 0
	for key, _ := range results {
		myArray[idx] = results[key]
		idx++
		// Logger.Println(results[key])
		// myArray[i] = "Hahah"
	}
	// Logger.Println("Complete", results)
	// Logger.Println("Double complete")
	// Logger.Println("Complete", myArray)
	err = nil
	return myArray, err
}
