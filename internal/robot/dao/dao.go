package dao

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"sync"
)

var (
	dao  *gorm.DB
	once = new(sync.Once)
)

type Stat struct {
	Ip   string
	Port uint16
	Send uint64
	Recv uint64
}

type Config struct {
	Ip   string
	Port uint16
	User string
	Pwd  string
	Db   string
}

func NewDao(c *Config) {
	once.Do(func() {
		url := fmt.Sprintf("%s:%s@(%s:%d)/%s?", c.User, c.Pwd, c.Ip, c.Port, c.Db)
		db, err := gorm.Open("mysql", url)
		if err != nil {
			panic("failed to connect database")
		}

		// Migrate the schema
		db.AutoMigrate(&Stat{})

		dao = db
	})
}

func InsertStat(ip string, port uint16, send, recv uint64) error {
	return dao.Create(&Stat{Ip: ip, Port: port, Send: send, Recv: recv}).Error
}
