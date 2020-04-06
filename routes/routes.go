package routes

import (
	"net/http"
	"github.com/gorilla/mux"
	"../middleware"
	"../sessions"
	"../models"
	"../utils"
)

func NewRouter() *mux.Router {

	router := mux.NewRouter()
	// wrap the indexHandlers with AuthRequired() to make check if user is already logged in to session
	router.HandleFunc("/", middleware.AuthRequired(indexGetHandler)).Methods("GET")
	router.HandleFunc("/", middleware.AuthRequired(indexPostHandler)).Methods("POST")

	router.HandleFunc("/login", loginGetHandler).Methods("GET")
	router.HandleFunc("/login", loginPostHandler).Methods("POST")

	router.HandleFunc("/logout", logoutGetHandler).Methods("GET")

	router.HandleFunc("/register", registerGetHandler).Methods("GET")
	router.HandleFunc("/register", registerPostHandler).Methods("POST")

	// static file server. Static files are in the static folder
	fileServer := http.FileServer(http.Dir("./static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileServer))

	router.HandleFunc("/{username}", middleware.AuthRequired(userGetHandler)).Methods("GET")

	// lets us create routes and determine the HTTP method used for the route
	return router
}

func indexGetHandler(w http.ResponseWriter, r *http.Request) {

	// get the userId from the session so we know who is accessing the home page
	session, _ := sessions.Store.Get(r, "session")
	untypedUserId := session.Values["user_id"]
	userId, ok := untypedUserId.(int64)
	if !ok {
		utils.InternalServerError(w)
		return
	}

	// get the user object by the userId
	user, err := models.GetUserById(userId)
	if err != nil {
		utils.InternalServerError(w)
		return 
	}

	// get the username from the user object
	username, err := user.GetUsername()
	if err != nil {
		utils.InternalServerError(w)
		return
	}

	// retrieve 10 most recent updates from ALL updates, execute the template with the data passed to an anonymous struct
	updates, err := models.GetAllUpdates()
	if err != nil { 
		utils.InternalServerError(w)
		return 
	}
	utils.ExecuteTemplate(w, "index.html", struct { 
		Title string 
		Updates []*models.Update
		DisplayForm bool
		User string
	} {
		Title: "All updates",
		Updates: updates,
		DisplayForm: true,
		User: username,
	})
}

func indexPostHandler(w http.ResponseWriter, r *http.Request) {
	
	// get the user from the session so we know who is posting
	session, _ := sessions.Store.Get(r, "session")
	untypedUserId := session.Values["user_id"]
	userId, ok := untypedUserId.(int64)
	if !ok {
		utils.InternalServerError(w)
		return
	}

	// get the post body from the form
	r.ParseForm()
	body := r.PostFormValue("update")

	// add the comment to the Redis db and check that it inserted successfully
	err := models.PostUpdate(userId, body)

	if err != nil {
		utils.InternalServerError(w)
		return 
	}
	// redirect to the home page after an insert
	http.Redirect(w, r, "/", 302)
}

func userGetHandler(w http.ResponseWriter, r *http.Request) {

	// retrieve the logged in user id from the session
	session, _ := sessions.Store.Get(r, "session")
	untypedUserId := session.Values["user_id"]
	userId, ok := untypedUserId.(int64)
	if !ok {
		utils.InternalServerError(w)
		return
	}

	// get the current user object by the userId
	user, err := models.GetUserById(userId)
	if err != nil {
		utils.InternalServerError(w)
		return 
	}

	// get the current user's username from the user object
	username, err := user.GetUsername()
	if err != nil {
		utils.InternalServerError(w)
		return
	}

	// get the username at the end of the url path -> localhost/visitedUsername
	vars := mux.Vars(r)
	visitedUsername := vars["username"]

	// get the visited user from the database
	visitedUser, err := models.GetUserByUsername(visitedUsername)
	if err != nil {
		utils.InternalServerError(w)
		return 
	}

	// get the visited user's id so we can then get their update posts
	visitedUserId, err := visitedUser.GetId()
	if err != nil {
		utils.InternalServerError(w)
		return 
	}

	// get all of the posts JUST for the visited User, not all posts
	updates, err := models.GetUpdates(visitedUserId)
	if err != nil { 
		utils.InternalServerError(w)
		return 
	}
	utils.ExecuteTemplate(w, "index.html", struct { 
		Title string 
		Updates []*models.Update
		DisplayForm bool
		User string
	} {
		Title: visitedUsername,
		Updates: updates,
		DisplayForm: userId == visitedUserId,
		User: username,
	})
}


// TODO: add userPostHandler so hitting the post button on the form from the user page doesn't redirect to all posts, but instead redirects to the users page

// show the login page when a GET request is made to login
func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "login.html", nil)
}


func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	// get the username and password from the login form
	r.ParseForm()
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	// Authenticate the user, respond to user with errors (in HTML format) if the username doesn't exist or if the password is wrong
	user, err := models.AuthenticateUser(username, password)
	if err != nil {
		switch err {
		case models.ErrUserNotFound:
			utils.ExecuteTemplate(w, "login.html", "unknown user")
		case models.ErrInvalidLogin:
			utils.ExecuteTemplate(w, "login.html", "incorrect password")
		default:
			utils.InternalServerError(w)
		}
		return
	}

	// get the ID from the user object that was returned by AuthenticateUser
	userId, err := user.GetId()
	if err != nil {
		utils.InternalServerError(w)
		return
	}

	// get the session store and save the user's id as the current session user_id
	session, _ := sessions.Store.Get(r, "session")
	session.Values["user_id"] = userId
	session.Save(r, w)

	// Redirect to the index page
	http.Redirect(w, r, "/", 302)
}


func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
	// clear the user id from the sessions and redirect to the login page
	session, _ := sessions.Store.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)

	// redirecting to the index page actually takes us to the login page since the session storage should have been cleared
	http.Redirect(w, r, "/", 302)
}

// show the register page when a GET request is made 
func registerGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "register.html", nil)
}

func registerPostHandler(w http.ResponseWriter, r *http.Request) {
	// get the username and password from the registraton form
	r.ParseForm()
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	// Save the username and password in the DB
	err := models.RegisterUser(username, password)

	// if the registration fails due to an invalid username, tell the user with some HTML
	if err == models.ErrUsernameTaken {
		utils.ExecuteTemplate(w, "register.html", "username taken")
		return
	} else if err != nil {
		utils.InternalServerError(w)
		return 
	}

	// if all is well, redirect to the login page
	http.Redirect(w, r, "/login", 302)
}