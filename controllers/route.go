package controllers

import (
	"html/template"
	"log"
	"net/http"
	"os"

	ctx "github.com/gorilla/context"
	// "github.com/gorilla/pat"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/iksose/phishClone/auth"
	mid "github.com/iksose/phishClone/middleware"
	"github.com/iksose/phishClone/models"
	"github.com/justinas/nosurf"
)

var templateDelims = []string{"{{%", "%}}"}
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

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
	router.HandleFunc("/articles/{article}/", ReadArticle)
	// router.HandleFunc("/", Use(Base, mid.RequireLogin))
	router.HandleFunc("/", Base)
	router.HandleFunc("/settings", Use(Settings, mid.RequireLogin))
	// Create the API routes
	api := router.PathPrefix("/api").Subrouter()
	api = api.StrictSlash(true)
	api.HandleFunc("/", Use(API, mid.RequireLogin))
	api.HandleFunc("/reset", Use(API_Reset, mid.RequireLogin))
	api.HandleFunc("/campaigns/", Use(API_Campaigns, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}", Use(API_Campaigns_Id, mid.RequireAPIKey))
	api.HandleFunc("/groups/", Use(API_Groups, mid.RequireAPIKey))
	api.HandleFunc("/groups/{id:[0-9]+}", Use(API_Groups_Id, mid.RequireAPIKey))
	api.HandleFunc("/templates/", Use(API_Templates, mid.RequireAPIKey))
	api.HandleFunc("/templates/{id:[0-9]+}", Use(API_Templates_Id, mid.RequireAPIKey))
	api.HandleFunc("/import/group", API_Import_Group)
	//
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
	delete(session.Values, "id")
	Flash(w, r, "success", "You have successfully logged out")
	http.Redirect(w, r, "login", 302)
}

func Base(w http.ResponseWriter, r *http.Request) {
	//get articles
	ark, err := models.GetArticles(1)
	if err != nil{
		Logger.Println("FUCK")
	}
	//get user
	user := models.User{};
	// check if context "user" is nil
	user2 := ctx.Get(r, "user")
	if user2 == nil{
		Logger.Println("User NIL")
	}else{
		Logger.Println("User is NOT NIL")
		user = ctx.Get(r, "user").(models.User)
	}

	Logger.Println("Got user?", user)
	// Logger.Println("Got articles?", ark)
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
		Articles []models.Article
	}{Title: "Dashboard", User: user, Token: nosurf.Token(r), Articles: ark}
	// getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
	var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/body.html", "templates/footer.html"))
	hogeTmpl.ExecuteTemplate(w, "header", params)
}

func ReadArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["article"]
	Logger.Println("Read Article?", category)
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
		Article models.Article
	}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User), Token: nosurf.Token(r), Article: models.GetArticle(category)}
	var hogeTmpl = template.Must(template.New("login").ParseFiles("templates/header.html", "templates/article.html", "templates/footer.html"))
	hogeTmpl.ExecuteTemplate(w, "header", params)
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
