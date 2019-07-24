package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"net/http"
	"os"

	"gopkg.in/mgo.v2/bson"

	. "simple-rest-api-go/config"
	. "simple-rest-api-go/dao"
	. "simple-rest-api-go/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

var config = Config{}
var dao = UsersDAO{}
var store *sessions.CookieStore
var log = logrus.New()

func init() {
	config.Read()

	dao.Server = config.Server
	dao.Database = config.Database
	dao.Connect()

	store = sessions.NewCookieStore([]byte(config.Sessionkey))

	//log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter)                  //default
	log.Formatter.(*logrus.TextFormatter).DisableColors = true // remove colors
	//log.Formatter.(*logrus.TextFormatter).DisableTimestamp = true // remove timestamp from test output
	log.Level = logrus.TraceLevel
	//log.Out = os.Stdout

	var filename string = "logfile.log"
	// Create the log file if doesn't exist. And append to it if it already exists.
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
	}

	log.SetOutput(f)

}

func encodePass(p string) string {
	h := sha256.New()
	h.Write([]byte(p))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// endpoints

// GET list of users
func AllUsers(w http.ResponseWriter, r *http.Request) {

	log.WithFields(logrus.Fields{
		"func": "AllUsers",
	}).Info("Shows all users")

	users, err := dao.FindAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, users)
}

// GET a user by ID
func FindUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user, err := dao.FindById(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}
	respondWithJson(w, http.StatusOK, user)
}

// GET a user by name
func FindUserByName(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user, err := dao.FindByName(params["name"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid User Name")
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJson(w, http.StatusOK, user)
}

// GET a user by Cookie
func FindUserByCookieKey(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)

	//var sess interface{}
	session, err := store.Get(r, "auth-key")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: false,
	}

	_, err2 := dao.FindByName(session.Values["name"].(string))
	if err2 != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid User Name")
		return
	}

	respondWithJson(w, http.StatusOK, session.Values["password"].(string))
}

func Login(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	//var sess interface{}
	password := encodePass(r.FormValue("password"))
	name := r.FormValue("login")

	user, err := dao.Login(name, password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Name and Password")
		return
	} else {

		//var sess interface{}
		session, err := store.Get(r, "auth-key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: false,
		}

		// Set some session values.
		session.Values["name"] = name
		session.Values["password"] = password

		store.Save(r, w, session)
	}

	respondWithJson(w, http.StatusOK, user)
}

// POST a new user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var user User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	user.ID = bson.NewObjectId()
	if err := dao.Insert(user); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusCreated, user)
}

// PUT update an existing user
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := dao.Update(user); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
}

// DELETE an existing user
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := dao.Delete(user); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Define HTTP request routes
func main() {

	r := mux.NewRouter()

	r.HandleFunc("/users", AllUsers).Methods("GET")
	r.HandleFunc("/users", CreateUser).Methods("POST")
	r.HandleFunc("/users", UpdateUser).Methods("PUT")
	r.HandleFunc("/users", DeleteUser).Methods("DELETE")
	r.HandleFunc("/users/{id}", FindUser).Methods("GET")
	//r.HandleFunc("/user/{name}", FindUserByName).Methods("GET")
	//r.Path("/user/{name}").Queries("password", "{password}").HandlerFunc(Login).Methods("GET")
	r.HandleFunc("/free", AllUsers).Methods("GET")
	r.HandleFunc("/user", Login).Methods("POST")
	r.HandleFunc("/user", FindUserByCookieKey).Methods("GET")
	r.HandleFunc("/user/{name}", FindUserByName).Methods("GET")
	r.HandleFunc("/admin/{name}", FindUserByName).Methods("GET")
	if err := http.ListenAndServe(":3000", r); err != nil {
		fmt.Println(err)
	}
}
