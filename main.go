package main

import (
	"time"
	"os"
	"fmt"
	"log"
	"database/sql"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

type sqlconf struct {
	User string `yaml:"user"`
	Password string `yaml:"pass"`
	Database string `yaml:"db"`
	Sslmode string `yaml:"sslmode"`
}

type session struct {
	Id int
	Day time.Time
	Minutes int
}

func (c *sqlconf) getConf() *sqlconf {
	yamlFile, err := os.ReadFile("sqlconf.yaml")
	if err != nil {
		log.Print(err)
	}
	err = yaml.Unmarshal(yamlFile, c)	
	if err != nil {
		log.Print(err)
	}

	return c
}

func main() {
	var c sqlconf
	c.getConf()

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", c.User, c.Password, c.Database, c.Sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Print("Error while openning database", err)
	}
	defer db.Close()

	rows, err := db.Query("select * from timer")
	if err != nil {
		log.Print("Error while selecting", err)
	}
	defer rows.Close()
	sessions := []session{}

	for rows.Next() {
		s := session{}
		err := rows.Scan(&s.Id, &s.Day, &s.Minutes)
		if err != nil {
			log.Print("Error while parsing", err)
			continue
		}
		sessions = append(sessions, s)
	}
	for _, s := range sessions {
		fmt.Println(s.Id)
	}
}
