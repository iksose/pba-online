package controllers

import (
	// "html/template"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	// "mime/multipart"
	// "path/filepath"
	"io"
	// "bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	// for the proxy
	// "flag"
	// "github.com/elazarl/goproxy"
	"net/url"

	ctx "github.com/gorilla/context"
	// "github.com/gorilla/pat"
	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/iksose/phishClone/auth"
	mid "github.com/iksose/phishClone/middleware"
	"github.com/iksose/phishClone/models"
	"github.com/justinas/nosurf"
)

var templateDelims = []string{"{{%", "%}}"}
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

// var proxy = goproxy.NewProxyHttpServer()

// var target *string
// var target = flag.String("target", "http://stackoverflow.com", "/tags")

// CreateAdminRouter creates the routes for handling requests to the web interface.
// This function returns an http.Handler to be used in http.ListenAndServe().
func CreateAdminRouter() http.Handler {
	// router := pat.New()
	// router.Get("/login", Login2)
	// router.Post("/login2", Login3)
	router := mux.NewRouter()
	// Base Front-end routes
	router.HandleFunc("/login", Login)
	// router.HandleFunc("/login2", Login2)
	router.HandleFunc("/logout", Use(Logout, mid.RequireLogin))
	router.HandleFunc("/elements/{article}/", ReadArticle)
	// router.HandleFunc("/articles/{article}/edit/", Use(EditArticle, mid.RequireLogin))
	router.HandleFunc("/elements/{article}/edit/", EditArticle)
	router.HandleFunc("/submitEdit", Use(submitEdit, mid.RequireLogin))
	// router.HandleFunc("/upload", Use(Upload, mid.RequireLogin))
	router.HandleFunc("/upload", Upload)
	router.HandleFunc("/about", About)
	router.HandleFunc("/products", Products)
	router.HandleFunc("/search", Search)
	router.HandleFunc("/contact", Contact)
	router.HandleFunc("/elements", Elements)
	router.HandleFunc("/new", NewPost)
	router.HandleFunc("/tags", GetTags)
	// router.HandleFunc("/negotiator/{article}/", Nego)

	// router.NotFoundHandler = http.HandlerFunc(notFound)
	// router.HandleFunc("/", Use(Base, mid.RequireLogin))
	router.HandleFunc("/", Base)

	// Create the API routes
	api := router.PathPrefix("/api").Subrouter()
	api = api.StrictSlash(true)
	// api.HandleFunc("/", Use(API, mid.RequireLogin))
	api.HandleFunc("/elements/{elementtype}", ElementsByType)

	// reverseProxy
	proxy := router.PathPrefix("/negotiator/").Subrouter()
	proxy = proxy.StrictSlash(true)
	proxy.HandleFunc("/{path:.*}", Nego)

	// // Setup static file serving
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	//
	// // Setup CSRF Protection
	// csrfHandler := nosurf.New(router)
	// // Exempt API routes and Static files
	// csrfHandler.ExemptGlob("/api/campaigns/*")
	// csrfHandler.ExemptGlob("/api/groups/*")
	// csrfHandler.ExemptGlob("/api/templates/*")
	// csrfHandler.ExemptGlob("/api/import/*")
	// csrfHandler.ExemptGlob("/static/*")
	// return Use(csrfHandler.ServeHTTP, mid.GetContext)
	// router.HandleFunc("/{path:.*}", notFound)
	return Use(router.ServeHTTP, mid.GetContext)
}

//CreateEndpointRouter creates the router that handles phishing connections.
func CreatePhishingRouter() http.Handler {
	router := mux.NewRouter()
	router.PathPrefix("/static").Handler(http.FileServer(http.Dir("./static/endpoint/")))
	router.HandleFunc("/{path:.*}", PhishHandler)
	return router
}

// PhishHandler handles incoming client connections and registers the associated actions performed
// (such as clicked link, etc.)
func PhishHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.Form.Get("rid")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	rs, err := models.GetResult(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rs.UpdateStatus(models.STATUS_SUCCESS)
	c, err := models.GetCampaign(rs.CampaignId, rs.UserId)
	if err != nil {
		Logger.Println(err)
	}
	c.AddEvent(models.Event{Email: rs.Email, Message: models.EVENT_CLICKED})
	w.Write([]byte("It Works!"))
}

// Use allows us to stack middleware to process the request
// Example taken from https://github.com/gorilla/mux/pull/36#issuecomment-25849172
func Use(handler http.HandlerFunc, mid ...func(http.Handler) http.HandlerFunc) http.HandlerFunc {
	for _, m := range mid {
		handler = m(handler)
	}
	return handler
}

type Response map[string]interface{}

func (r Response) String() (s string) {
	b, err := json.Marshal(r)
	if err != nil {
		s = ""
		return
	}
	s = string(b)
	return
}

func ElementsByType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["elementtype"]
	Logger.Println("get all by this type??", category)
	articles, err := models.ElementsByType(category)
	if err != nil {
		Logger.Println(err)
		JSONResponse(w, "ok", http.StatusBadRequest)
		return
	}
	Logger.Println("ok", articles)
	fmt.Fprint(w, Response{"articles": articles})
	return
}

func GetTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	Logger.Println("Get Tags")
	tags, err := models.GetTags()
	if err != nil {
		Logger.Println("Fuck")
	}
	// fmt.Fprint(w, Response{"success": true, "message": "Hello!"})
	fmt.Fprint(w, Response{"tags": tags})
	return
	// msg := models.Response{Success: true, Message: "Settings Updated Successfully", Data: tags}
	// JSONResponse(w, msg, http.StatusOK)

	// JSONResponse(w, "ok", http.StatusOK)
}

type myjar struct {
	jar map[string][]*http.Cookie
}

func (p *myjar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	fmt.Printf("The URL is : %s\n", u.String())
	// fmt.Printf("The cookie being set is : %s\n", cookies)
	p.jar[u.Host] = cookies
}

func (p *myjar) Cookies(u *url.URL) []*http.Cookie {
	fmt.Printf("The URL is : %s\n", u.String())
	// fmt.Printf("Cookie being returned is : %s\n", p.jar[u.Host])
	return p.jar[u.Host]
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func copyHeader(source http.Header, dest *http.Header) {
	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}

func Nego(w http.ResponseWriter, r *http.Request) {
	// path
	vars := mux.Vars(r)
	path := vars["path"]
	Logger.Println("path? ", path)

	//cookies
	jar := &myjar{}
	jar.jar = make(map[string][]*http.Cookie)

	client := &http.Client{}
	// client.Jar = jar
	// target = flag.String("target", "https://www.pbahealth.com/pbao/Default.aspx", "")
	// flag.Parse()
	// uri := *target + r.RequestURI
	uri := "https://www.pbahealth.com/pbao/" + path
	// uri := "https://www.pbahealth.com/pbao/pbaoUnifiedMain.aspx"

	fmt.Println(r.Method + ": " + uri)

	if r.Method == "POST" {
		Logger.Println("Post???")
		body, err := ioutil.ReadAll(r.Body)
		fatal(err)
		fmt.Printf("Body: %v\n", string(body))
	}

	rr, err := http.NewRequest(r.Method, uri, r.Body)
	fatal(err)
	copyHeader(r.Header, &rr.Header)
	// rr.SetBasicAuth("12600370", "test1234")
	// rr.AddCookie(&cookie)
	rr.Header.Set("Cookie", "ASP.NET_SessionId=eqi40555bj2djg45p3jfyx55")

	//jon

	// client := &http.Client{}
	// client.Jar = jar
	resp, err := client.Do(rr)
	fatal(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fatal(err)

	dH := w.Header()
	copyHeader(resp.Header, &dH)
	dH.Add("Requested-Host", rr.Host)

	w.Write(body)

	//end jon

	// Create a client and query the target
	// var transport http.Transport
	// resp, err := transport.RoundTrip(rr)
	// fatal(err)

	// fmt.Printf("Resp-Headers: %v\n", resp.Header)

	// defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// fatal(err)

	// dH := w.Header()
	// copyHeader(resp.Header, &dH)
	// dH.Add("Requested-Host", rr.Host)

	// w.Write(body)
	// w.Header().Set("Content-Type", "text/html")
	// jar := &myjar{}
	// jar.jar = make(map[string][]*http.Cookie)

	// client := &http.Client{}
	// client.Jar = jar

	// /* Authenticate */
	// req, err := http.NewRequest("GET", "https://www.pbahealth.com/pbao/Negotiator/negMain.aspx", nil)
	// req.SetBasicAuth("12600370", "test1234")
	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Printf("Error : %s", err)
	// }
	// Logger.Println("Okay...")
	// // /* Get Details */
	// // req.URL, _ = url.Parse("http://164.99.113.32/Details")
	// // resp, err = client.Do(req)
	// // if err != nil {
	// // 	fmt.Printf("Error : %s", err)
	// // }
	// Logger.Println(resp)
	// // fmt.Fprint(w, Response{"articles": resp})
	// // http.Redirect(w, r, "https://www.pbahealth.com/pbao/Negotiator/negMain.aspx", 302)
	// // w = resp
	// fmt.Fprint(w, r)
	// return
}

func About(w http.ResponseWriter, r *http.Request) {
	//get user
	user := models.User{}
	// check if context "user" is nil
	user2 := ctx.Get(r, "user")
	if user2 == nil {
		Logger.Println("User NIL")
	} else {
		user = ctx.Get(r, "user").(models.User)
	}
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
		// Articles []models.Article
	}{Title: "Dashboard", User: user, Token: nosurf.Token(r)}
	// getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
	var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/about.html", "templates/footer.html"))
	hogeTmpl.ExecuteTemplate(w, "header", params)
}

func Elements(w http.ResponseWriter, r *http.Request) {
	//get articles
	ark, err := models.GetArticles(1)
	if err != nil {
		Logger.Println("FUCK")
	}
	//get user
	user := models.User{}
	// check if context "user" is nil
	user2 := ctx.Get(r, "user")
	if user2 == nil {
		Logger.Println("User NIL")
	} else {
		Logger.Println("User is NOT NIL")
		user = ctx.Get(r, "user").(models.User)
	}

	Logger.Println("Got user?", user)
	// Logger.Println("Got articles?", ark)
	// Example of using session - will be removed.
	params := struct {
		User     models.User
		Title    string
		Flashes  []interface{}
		Token    string
		Articles []models.Article
	}{Title: "Dashboard", User: user, Token: nosurf.Token(r), Articles: ark}
	// getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
	var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/elements.html", "templates/footer.html"))
	hogeTmpl.ExecuteTemplate(w, "header", params)
}

// also handles deletes
func NewPost(w http.ResponseWriter, r *http.Request) {
	//get articles
	//get user
	user := models.User{}
	// check if context "user" is nil
	user2 := ctx.Get(r, "user")
	if user2 == nil {
		Logger.Println("User NIL")
	} else {
		Logger.Println("User is NOT NIL")
		user = ctx.Get(r, "user").(models.User)
	}

	Logger.Println("Got user?", user)
	switch {
	case r.Method == "GET":
		params := struct {
			User    models.User
			Title   string
			Flashes []interface{}
			Token   string
			// Articles []models.Article
		}{Title: "Dashboard", User: user, Token: nosurf.Token(r)}
		// getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
		var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/new.html", "templates/footer.html"))
		hogeTmpl.ExecuteTemplate(w, "header", params)
	case r.Method == "POST":
		Logger.Println("Post to new post")
		decoder := json.NewDecoder(r.Body)
		var t models.Article
		err := decoder.Decode(&t)
		if err != nil {
			// panic()
			Logger.Println("Fuck", err)
			//return 500
			JSONResponse(w, "ok", http.StatusBadRequest)
			return
		}
		Logger.Println("Sweet", t)
		err = models.PostArticle(t)
		if err != nil {
			Logger.Println("Fuck", err)
			JSONResponse(w, "ok", http.StatusBadRequest)
			return
		}
		JSONResponse(w, "ok", http.StatusOK)
	case r.Method == "DELETE":
		Logger.Println("Delete")
		decoder := json.NewDecoder(r.Body)
		var article models.Article
		err := decoder.Decode(&article)
		if err != nil {
			Logger.Println("Fuck", err)
			//return 500
			JSONResponse(w, "ok", http.StatusBadRequest)
		}
		Logger.Println("Success!", article.Id)
		err = models.DeleteArticle(article)
		if err != nil {
			JSONResponse(w, "ok", http.StatusBadRequest)
		}
		//return 500
		JSONResponse(w, "ok", http.StatusOK)
	}
}

func Contact(w http.ResponseWriter, r *http.Request) {
	//get user
	user := models.User{}
	// check if context "user" is nil
	user2 := ctx.Get(r, "user")
	if user2 == nil {
		Logger.Println("User NIL")
	} else {
		user = ctx.Get(r, "user").(models.User)
	}
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
		// Articles []models.Article
	}{Title: "Dashboard", User: user, Token: nosurf.Token(r)}
	// getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
	var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/contact.html", "templates/footer.html"))
	hogeTmpl.ExecuteTemplate(w, "header", params)
}

func Products(w http.ResponseWriter, r *http.Request) {
	//get user
	user := models.User{}
	// check if context "user" is nil
	user2 := ctx.Get(r, "user")
	if user2 == nil {
		Logger.Println("User NIL")
	} else {
		user = ctx.Get(r, "user").(models.User)
	}
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
		// Articles []models.Article
	}{Title: "Dashboard", User: user, Token: nosurf.Token(r)}
	// getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
	var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/products_and_services.html", "templates/footer.html"))
	hogeTmpl.ExecuteTemplate(w, "header", params)
}
func Search(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()
	// Logger.Println("param", param)
	name := param.Get("name")
	Logger.Println("Uhhh", name)
	res, err := models.GetResults(name)
	if err != nil {
		Logger.Println("FUCK")
	}
	//get user
	user := models.User{}
	// check if context "user" is nil
	user2 := ctx.Get(r, "user")
	if user2 == nil {
		Logger.Println("User NIL")
	} else {
		user = ctx.Get(r, "user").(models.User)
	}
	params := struct {
		User       models.User
		Title      string
		Flashes    []interface{}
		Token      string
		Results    []models.Article
		QueryParam string
	}{Title: "Dashboard", User: user, Token: nosurf.Token(r), Results: res, QueryParam: name}
	// getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
	var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/search.html", "templates/footer.html"))
	hogeTmpl.ExecuteTemplate(w, "header", params)
}

func Register(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	params := struct {
		Title   string
		Flashes []interface{}
		User    models.User
		Token   string
	}{Title: "Register", Token: nosurf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)
	switch {
	case r.Method == "GET":
		params.Flashes = session.Flashes()
		session.Save(r, w)
		getTemplate(w, "register").ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to register
		succ, err := auth.Register(r)
		//If we've registered, redirect to the login page
		if succ {
			session.AddFlash(models.Flash{
				Type:    "success",
				Message: "Registration successful!.",
			})
			session.Save(r, w)
			http.Redirect(w, r, "/login", 302)
		} else {
			// Check the error
			m := ""
			if err == models.ErrUsernameTaken {
				m = "Username already taken"
			} else {
				m = "Unknown error - please try again"
				Logger.Println(err)
			}
			session.AddFlash(models.Flash{
				Type:    "danger",
				Message: m,
			})
			session.Save(r, w)
			http.Redirect(w, r, "/register", 302)
		}

	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	session := ctx.Get(r, "session").(*sessions.Session)
	Logger.Println("Session?", session.Values["id"])
	delete(session.Values, "id")
	sessions.Save(r, w) // jonathan -- why wasn't this already here
	// Flash(w, r, "success", "You have successfully logged out")
	// Could do a post to logout and refresh the current page
	http.Redirect(w, r, "/", 302)
}

func Base(w http.ResponseWriter, r *http.Request) {
	//get articles
	ark, err := models.GetArticles(1)
	if err != nil {
		Logger.Println("FUCK")
	}
	//get user
	user := models.User{}
	// check if context "user" is nil
	user2 := ctx.Get(r, "user")
	if user2 == nil {
		Logger.Println("User NIL")
	} else {
		Logger.Println("User is NOT NIL")
		user = ctx.Get(r, "user").(models.User)
	}

	Logger.Println("Got user?", user)
	// Logger.Println("Got articles?", ark)
	// Example of using session - will be removed.
	params := struct {
		User     models.User
		Title    string
		Flashes  []interface{}
		Token    string
		Articles []models.Article
	}{Title: "Dashboard", User: user, Token: nosurf.Token(r), Articles: ark}
	// getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
	var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/body.html", "templates/footer.html"))
	hogeTmpl.ExecuteTemplate(w, "header", params)
}

func ReadArticle(w http.ResponseWriter, r *http.Request) {
	//get user
	user := models.User{}
	// check if context "user" is nil
	user2 := ctx.Get(r, "user")
	if user2 == nil {
		Logger.Println("User NIL")
	} else {
		Logger.Println("User is NOT NIL")
		user = ctx.Get(r, "user").(models.User)
	}
	vars := mux.Vars(r)
	category := vars["article"]
	Logger.Println("Read Article?", category)
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
		Article models.Article
	}{Title: "Dashboard", User: user, Token: nosurf.Token(r), Article: models.GetArticle(category)}
	var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/article.html", "templates/footer.html"))
	hogeTmpl.ExecuteTemplate(w, "header", params)
}

func EditArticle(w http.ResponseWriter, r *http.Request) {
	Logger.Println("Edit an article?")
	//get user
	user := models.User{}
	// check if context "user" is nil
	user2 := ctx.Get(r, "user")
	if user2 == nil {
		Logger.Println("User NIL")
	} else {
		Logger.Println("User is NOT NIL")
		user = ctx.Get(r, "user").(models.User)
	}
	vars := mux.Vars(r)
	category := vars["article"]
	Logger.Println("Read Article?", category)
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
		Article models.Article
	}{Title: "Dashboard", User: user, Token: nosurf.Token(r), Article: models.GetArticle(category)}
	var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/edit.html", "templates/footer.html"))
	hogeTmpl.ExecuteTemplate(w, "header", params)
}

func submitEdit(w http.ResponseWriter, r *http.Request) {
	Logger.Println("Well hello there")
	decoder := json.NewDecoder(r.Body)
	var t models.Article
	err := decoder.Decode(&t)
	if err != nil {
		// panic()
		Logger.Println("Fuck", err)
	}
	Logger.Println(t.Body)
	err = models.EditArticle(t)
	if err != nil {
		Logger.Println("Fuck", err)
		JSONResponse(w, models.Response{Success: false, Message: "lol"}, http.StatusBadRequest)
		return
	}
	// http.Error(w, http.StatusText(500), 500)
	JSONResponse(w, "ok", http.StatusOK)
}

func getImageDimension(imagePath string) (int, int) {
	// Logger.Println("Get image dimensions", imagePath)
	file, err := os.Open(imagePath)
	if err != nil {
		Logger.Println(os.Stderr, "%v\n", err)
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		Logger.Println(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return image.Width, image.Height
}

func Upload(w http.ResponseWriter, r *http.Request) {
	// file, handler, err := r.FormFile("file")
	// if err != nil {
	// 	Logger.Println("Fuck", err)
	// }
	// data, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	Logger.Println("Fuck", err)
	// }
	// err = ioutil.WriteFile("./static/uploads/"+handler.Filename, data, 0777)
	// if err != nil {
	// 	Logger.Println(handler.Filename, err)
	// }

	//get the multipart reader for the request.

	type Dimensions struct {
		X      int
		Y      int
		Width  int
		Height int
	}

	var dimens Dimensions
	var dest string
	var filename string
	msg := models.Response{Success: true, Message: "Settings Updated Successfully", Data: "Ha"}

	reader, err := r.MultipartReader()
	if err != nil {
		Logger.Println("Fuck", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//copy each part to destination.
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if part.FormName() == "avatar_data" {
			Logger.Println("Here", part.FormName())
			j, err := ioutil.ReadAll(part)
			if err != nil {
				Logger.Println("FUCK", err)
			}
			s := string(j)
			// Logger.Println("YES", j)
			Logger.Println("YES", s)
			json.Unmarshal(j, &dimens)
			Logger.Println(dimens)
			// err = json.NewDecoder(io.Reader).Decode(dimens)
			// err = json.Unmarshal([]byte(s), &dimens)
			// if err == nil {
			// 	Logger.Printf("%+v\n", dimens)
			// } else {
			// 	Logger.Println(err)
			// 	Logger.Printf("%+v\n", dimens)
			// }
		}
		//if part.FileName() is empty, skip this iteration.
		if part.FileName() == "" {
			continue
		}
		dst, err := os.Create("/Users/macadmin/gocode/src/github.com/iksose/phishClone/static/uploads/" + part.FileName())
		defer dst.Close()
		dest = "/Users/macadmin/gocode/src/github.com/iksose/phishClone/static/uploads/" + part.FileName()
		filename = part.FileName()
		if err != nil {
			Logger.Println("Fuck", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(dst, part); err != nil {
			Logger.Println("Fuck", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// var thumbnails image.Image
	// img, _ := imaging.Open("/Users/macadmin/test/cry3.jpg")
	img, _ := imaging.Open(dest)
	if err != nil {
		Logger.Println("Fuck", err)
	}

	old_width, old_height := getImageDimension(dest)
	Logger.Println("Width:", old_width, "old_height:", old_height)
	// dst := imaging.New(100*len(thumbnails), 100, color.NRGBA{0, 0, 0, 0})
	//func Crop(img image.Image, rect image.Rectangle) *image.NRGBA
	// type Rectangle struct {
	//    length, width int
	// }
	// rect := image.Rect(0, 0, 200, 500)
	Logger.Println("Applying transformation", dimens)
	// combo := -old_width + dimens.X
	// Logger.Println("Combo", combo)

	thirdnumber := dimens.X + dimens.Width

	fourthnumber := dimens.Y + dimens.Height

	Logger.Println("Third", thirdnumber)

	rect := image.Rect(dimens.X, dimens.Y, thirdnumber, fourthnumber)

	// rect := image.Rect(0, 100, 700, 700)
	//zoom, up, width right, height
	dst := imaging.Crop(img, rect)
	// err = imaging.Save(dst, "/Users/macadmin/test/cry4.jpg")
	err = imaging.Save(dst, dest)
	if err != nil {
		panic(err)
	}
	Logger.Println("Success")
	//display success message.
	// display(w, "upload", "Upload successful.")
	// msg.result = "gj bro"
	msg.Data = "/uploads/" + filename
	JSONResponse(w, msg, http.StatusOK)
}

func Settings(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "POST":
		err := auth.ChangePassword(r)
		msg := models.Response{Success: true, Message: "Settings Updated Successfully"}
		if err == auth.ErrInvalidPassword {
			msg.Message = "Invalid Password"
			msg.Success = false
		} else if err != nil {
			msg.Message = "Unknown Error Occured"
			msg.Success = false
		}
		JSONResponse(w, msg, http.StatusOK)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Login", Token: nosurf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)
	switch {
	case r.Method == "GET":
		Logger.Println("GET login...")
		params.Flashes = session.Flashes()
		session.Save(r, w)
		// templates := template.New("template")
		// templates.Delims(templateDelims[0], templateDelims[1])
		// _, err := templates.ParseFiles("templates/login.html", "templates/flashes.html")
		// if err != nil {
		// 	Logger.Println("FUCK", err)
		// }
		// template.Must(templates, err).ExecuteTemplate(w, "base", params)
		var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/body.html"))
		hogeTmpl.ExecuteTemplate(w, "header", params)
	case r.Method == "POST":
		Logger.Println("POST - Attempt to login")
		//Attempt to login
		succ, err := auth.Login(r)
		if err != nil {
			Logger.Println(err)
		}
		//If we've logged in, save the session and redirect to the dashboard
		if succ {
			session.Save(r, w)
			http.Redirect(w, r, "/", 302)
		} else {
			Flash(w, r, "danger", "Invalid Username/Password")
			http.Redirect(w, r, "/login", 302)
		}
	case r.Method == "PUT":
		Logger.Println("PUT to login")
	}
}

func getTemplate(w http.ResponseWriter, tmpl string) *template.Template {
	templates := template.New("template")
	templates.Delims(templateDelims[0], templateDelims[1])
	_, err := templates.ParseFiles("templates/base.html", "templates/"+tmpl+".html", "templates/flashes.html")
	if err != nil {
		Logger.Println(err)
	}
	return template.Must(templates, err)
}

func checkError(e error, w http.ResponseWriter, m string, c int) bool {
	if e != nil {
		Logger.Println(e)
		w.WriteHeader(c)
		JSONResponse(w, models.Response{Success: false, Message: m}, http.StatusBadRequest)
		return true
	}
	return false
}

func Flash(w http.ResponseWriter, r *http.Request, t string, m string) {
	session := ctx.Get(r, "session").(*sessions.Session)
	session.AddFlash(models.Flash{
		Type:    t,
		Message: m,
	})
	session.Save(r, w)
}
