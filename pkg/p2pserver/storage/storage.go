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

package storage

import (
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	dao  *gorm.DB
	once = new(sync.Once)
)

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
		db.AutoMigrate(&Subnet{})

		dao = db
	})
}

type Subnet struct {
	Pubkey string
}

func InsertStat(id string) error {
	return dao.Create(&Subnet{Pubkey: id}).Error
}

func CheckSubnet(id string) bool {
	data := new(Subnet)
	affected := dao.Where("pubkey = ?", id).First(data).RowsAffected
	return affected > 0
}

func GetAllSubnet() []string {
	list := []string{}
	ids := []*Subnet{}
	dao.Find(&ids)
	for _, v := range ids {
		list = append(list, v.Pubkey)
	}
	return list
}
