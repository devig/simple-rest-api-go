package dao

import (
	"log"

	. "simple-rest-api-go/models"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UsersDAO struct {
	Server   string
	Database string
}

var db *mgo.Database

const (
	COLLECTION = "users"
)

// Establish a connection to database
func (m *UsersDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

// Find list of users
func (m *UsersDAO) FindAll() ([]User, error) {
	var users []User
	err := db.C(COLLECTION).Find(bson.M{}).All(&users)
	return users, err
}

// Find a user by id
func (m *UsersDAO) FindById(id string) (User, error) {
	var user User
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&user)
	return user, err
}

// Find a user by name
func (m *UsersDAO) FindByName(name string) (User, error) {
	var user User
	err := db.C(COLLECTION).Find(bson.M{"name": name}).One(&user)
	return user, err
}

//Login
func (m *UsersDAO) Login(name string, password string) (User, error) {
	var user User
	err := db.C(COLLECTION).Find(bson.M{"name": name, "password": password}).One(&user)
	return user, err
}

// Insert a user into database
func (m *UsersDAO) Insert(user User) error {
	err := db.C(COLLECTION).Insert(&user)
	return err
}

// Delete an existing user
func (m *UsersDAO) Delete(user User) error {
	err := db.C(COLLECTION).Remove(&user)
	return err
}

// Update an existing user
func (m *UsersDAO) Update(user User) error {
	err := db.C(COLLECTION).UpdateId(user.ID, &user)
	return err
}
