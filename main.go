package main

import (
	"os/exec"
	"flag"
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

func (s *session) dates() (int, time.Month, int) {
	return s.Day.Date() 
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

func allRows(db *sql.DB) []session {
	rows, err := db.Query("select * from timer")
	if err != nil {
		log.Println("Error occuring while select", err)
	}
	defer rows.Close()

	sessions := []session{}
	for rows.Next() {
		s := session{}
		err := rows.Scan(&s.Id, &s.Day, &s.Minutes)
		if err != nil {
			log.Println("Error occuring while parsing", err)
		}
		sessions = append(sessions, s)
	}

	return sessions
}

func checkLastOne(db *sql.DB) session {
	row, err := db.Query("select * from timer order by id DESC limit 1")
	if err != nil {
		log.Println("Error while finding")
	}
	s := session{}
	row.Next()
	err = row.Scan(&s.Id, &s.Day, &s.Minutes)
	if err != nil {
		log.Println("Occuring while parsing with error:", err)
	}
	return s
}

func whenClose(db *sql.DB, lastRev session, openProc time.Time) {
	yn, mn, dn := openProc.Date()
	y, m, d := lastRev.Day.Date()
	if yn == y && mn == m && dn == d {
		fmt.Println("all ok")
		_, err := db.Exec("update timer set minutes=minutes+$1 where id=$2;", int((time.Since(openProc))/time.Minute), lastRev.Id)
		if err != nil {
			log.Println("error while updating", err)
		}
		fmt.Println(lastRev.Id)
	} else {
		fmt.Println("new date")
		_, _ = db.Exec("insert into timer(minutes) values($1);", time.Since(openProc))
	}
}

func main() {
	now := time.Now()
	var process string
	flag.StringVar(&process, "exec", "", "Process which we try to exect")
	flag.Parse()
	var c sqlconf
	c.getConf()
	cmd := exec.Command("vim", process)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Println("error while execute programm", err)
	}

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", c.User, c.Password, c.Database, c.Sslmode)

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Println("Error while openning database", err)
	}
	defer db.Close()

	lastSession := checkLastOne(db)
	fmt.Println(lastSession)
	defer whenClose(db, lastSession, now)
}
