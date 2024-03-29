package models

import "gopkg.in/mgo.v2/bson"

type User struct {
	ID       bson.ObjectId `bson:"_id" json:"id"`
	Name     string        `bson:"name" json:"name"`
	password string        `bson:"password" json:"password"`
	IsAdmin  bool          `bson:"isadmin" json:"isadmin"`
}
