package model

import (
	"game_server/common"
	"log"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// User depend on redis and mgo
// use user model as respositroy

type User struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Email     string        `bson:"email" json:"email"`
	Password  string        `bson:"password" json:"-"` // save in hash
	Sid       string        `bson:"sid" json:"-"`      // like token, as user certificate, default ''
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
}

func (m *User) GetId() string {
	return m.Id.Hex()
}

func (m *User) GetEmail() string {
	return m.Email
}

func (m *User) GetPassword() string {
	return m.Password
}

func (m *User) GetSid() string {
	return m.Sid
}

func (m *User) GetCreatedAt() string { // get unix str
	str := strconv.FormatInt(m.CreatedAt.Unix(), 10)
	return str
}

func (m *User) GetUpdatedAt() string { // get unix str
	str := strconv.FormatInt(m.UpdatedAt.Unix(), 10)
	return str
}

func (m *User) SetId(id string) {
	m.Id = bson.ObjectIdHex(id)
}

func (m *User) SetEmail(email string) {
	m.Email = email
}

func (m *User) SetPassword(pwd string) { // use hash
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return
	}
	m.Password = string(hashPwd)
}

func (m *User) SetSid(sid string) {
	m.Sid = sid
}

func (m *User) SetCreatedAt(str string) {
	unix, _ := strconv.ParseInt(str, 10, 64)
	m.CreatedAt = time.Unix(unix, 0)
}

func (m *User) SetUpdatedAt(str string) {
	unix, _ := strconv.ParseInt(str, 10, 64)
	m.UpdatedAt = time.Unix(unix, 0)
}

// only update sid
func (m *User) UpdateSid(sid string) {
	ms := common.GetMgo().NewSession()
	defer ms.Close()
	c := ms.C("users")

	m.UpdatedAt = time.Now()
	c.Update(bson.M{"_id": m.Id}, bson.M{
		"$set": bson.M{
			"sid":        sid,
			"updated_at": m.UpdatedAt,
		},
	})
}

// Find user from mgo by email
func FindUserById(id string) *User {
	ms := common.GetMgo().NewSession()
	defer ms.Close()
	c := ms.C("users")

	usr := &User{}
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(usr)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Println(err)
		}
		return nil
	}

	return usr
}

// Find user from mgo by email
func FindUserByEmail(email string) *User {
	ms := common.GetMgo().NewSession()
	defer ms.Close()
	c := ms.C("users")

	usr := &User{}
	err := c.Find(bson.M{"email": email}).One(usr)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Println(err)
		}
		return nil
	}

	return usr
}

// Find user from mgo by sid
func FindUserBySid(sid string) *User {
	ms := common.GetMgo().NewSession()
	defer ms.Close()
	c := ms.C("users")

	usr := &User{}
	err := c.Find(bson.M{"sid": sid}).One(usr)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Println(err)
		}
		return nil
	}

	return usr
}

// Create user into mgo
func CreateUser(email, password string) {
	ms := common.GetMgo().NewSession()
	defer ms.Close()
	c := ms.C("users")

	usr := &User{}
	usr.SetEmail(email)
	usr.SetPassword(password)
	usr.CreatedAt = time.Now()
	usr.UpdatedAt = time.Now()
	err := c.Insert(usr)
	if err != nil {
		log.Println(err)
	}
}

// Save data into mgo
func (m *User) Save() {
	ms := common.GetMgo().NewSession()
	defer ms.Close()
	c := ms.C("users")

	m.UpdatedAt = time.Now()
	c.Update(bson.M{"_id": m.Id}, bson.M{
		"$set": bson.M{
			"updated_at": m.UpdatedAt, // no save sid, password, email, created_at etc..
		},
	})
}

// User key at redis
func UserRedisKey(id string) string {
	return "users:" + id
}

// Sid key at redis
func SidRedisKey(sid string) string {
	return "sid:" + sid
}

// Redis user exists
func UserStorageExists(id string) bool {
	key := UserRedisKey(id)
	bol, _ := common.GetRedis().HExists(key, "id")
	return bol
}

// Find user from redis by id
func LoadUserById(id string) *User {
	key := UserRedisKey(id)

	uMap, err := common.GetRedis().HGetAll(key)
	if err != nil {
		log.Println(err)
		return nil
	}

	if len(uMap) == 0 {
		return nil
	}

	usr := &User{}
	usr.SetId(uMap["id"])
	usr.SetEmail(uMap["email"])
	usr.SetPassword(uMap["password"])
	usr.SetSid(uMap["sid"])
	usr.SetCreatedAt(uMap["created_at"])
	usr.SetUpdatedAt(uMap["updated_at"])

	return usr
}

// Storage data into redis
func (m *User) Storage() {
	key := UserRedisKey(m.GetId())

	m.UpdatedAt = time.Now()
	common.GetRedis().HMSet(key,
		"id", m.GetId(),
		"email", m.GetEmail(),
		"password", m.GetPassword(),
		"sid", m.GetSid(),
		"created_at", m.GetCreatedAt(),
		"updated_at", m.GetUpdatedAt(),
	)
}

// Clear user storage data
func (m *User) ClearStorage() {
	key := UserRedisKey(m.GetId())
	common.GetRedis().Del(key)
}

// Get user, first check redis, if nil then use mgo
func GetUserById(id string) (usr *User) {
	if UserStorageExists(id) {
		usr = LoadUserById(id)
	} else {
		usr = FindUserById(id)
	}
	return
}

// Set user, first check redis, if nil then use redis
func (m *User) Set() {
	if UserStorageExists(m.GetId()) {
		m.Storage()
	} else {
		m.Save()
	}
}
