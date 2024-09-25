package test

import "go.mongodb.org/mongo-driver/v2/bson"

// model
type UserModel struct {
	Name, Sex   string
	PhoneNumber Login
	Test        []string
}

// Embed
type Login struct {
	Name string
	Type string
	//ID   bson.ObjectID
}

// Embed
type UserT struct {
	Name  string
	Login Login
	ID    []*bson.ObjectID
}

// model
type GirlModel struct {
	Nam333e string // ggg
	Time    string
}

type LoginType int

const (
	LoginType_UNKNOWN LoginType = iota
	LoginType_TEST
	LoginType_FOO
	LoginType_BAR
	LoginType_INF
)

type AppRole int

const (
	AppRole_UNKNOWN AppRole = iota
	AppRole_ADMIN
	AppRole_USER
)
