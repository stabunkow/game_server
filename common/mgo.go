package common

import (
	"log"
	"sync"

	"gopkg.in/mgo.v2"
)

var defaultMgo *Mgo
var defaultMgoOnce sync.Once

type Mgo struct {
	session  *mgo.Session
	database string
}

// clone session after use must be closed
type MgoSession struct {
	session  *mgo.Session
	database string
}

func GetMgo() *Mgo {
	defaultMgoOnce.Do(func() {
		defaultMgo = newMgo(":27017", "test")
	})
	return defaultMgo
}

func newMgo(server, database string) *Mgo {
	session, err := mgo.Dial(server)
	if err != nil {
		log.Println("new mgo generate session failed, err:", err)
		return nil
	}
	session.SetMode(mgo.Monotonic, true)

	return &Mgo{
		session:  session,
		database: database,
	}
}

func (m *Mgo) NewSession() *MgoSession {
	return &MgoSession{
		session:  m.session.Clone(),
		database: m.database,
	}
}

func (m *MgoSession) C(table string) *mgo.Collection {
	return m.session.DB(m.database).C(table)
}

func (m *MgoSession) Close() {
	m.session.Close()
}
