/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package dao

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

type Config struct {
	Ip   string
	Port uint16
	User string
	Pwd  string
	Db   string
}

func NewDao(c *Config) {
	if db != nil {
		return
	}

	var err error

	url := fmt.Sprintf("%s:%s@(%s:%d)/%s?", c.User, c.Pwd, c.Ip, c.Port, c.Db)
	if db, err = gorm.Open("mysql", url); err != nil {
		panic("failed to connect database")
	}

	var initTable = func(tb interface{}) {
		// db.DropTableIfExists(tb)
		// db.CreateTable(tb)
		db.AutoMigrate(tb)
	}

	initTable(&Stat{})
	initTable(&Subnet{})
}

type Stat struct {
	Ip   string
	Port uint16
	Hash string
	Send uint64
	Recv uint64
}

func InsertStat(ip string, port uint16, hash string, send, recv uint64) error {
	return db.Create(&Stat{Ip: ip, Port: port, Hash: hash, Send: send, Recv: recv}).Error
}

type Subnet struct {
	Pubkey string
}

func CheckSubnet(id string) bool {
	data := new(Subnet)
	affected := db.Where("pubkey = ?", id).First(data).RowsAffected
	return affected > 0
}

func GetAllSubnet() []string {
	list := []string{}
	ids := []*Subnet{}
	db.Find(&ids)
	for _, v := range ids {
		list = append(list, v.Pubkey)
	}
	return list
}
