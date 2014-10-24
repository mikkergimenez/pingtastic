package main

import (
	"./alexa"
	"fmt"
	"github.com/tatsushid/go-fastping"
	"database/sql"
	//"github.com/go-sql-driver/mysql"
	"net"
	"time"
	//"syscall"
	//"os/signal"	
)

type MysqlServer struct {
	url string
	user string
	password string
	database string
}

type Website struct {
    Url string
    Name string
    PingValue int
}

type Layout struct {
	x_range, y_range int
}

type Ping struct {
	rtt int
	url string
}

func doPing(pingTime chan Ping) {

	p := fastping.NewPinger()

	ra, err := net.ResolveIPAddr("ip4:icmp", os.Args[1])

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p.AddIPAddr(ra)

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	}

	p.OnIdle = func() {
		fmt.Println("finish")
	}

	err = p.Run()

	if err != nil {
		fmt.Println(err)
	}

}

func writeToDB(db MysqlServer, ping chan Ping) {
	con, err := sql.Open("mysql", db.user+":"+db.password+"@/"+db.database)
}

func runServer(db MysqlServer) {
	for _ = range time.Tick(60 * time.Second) {
		var alexaList [1000]string
		for _, website := range alexaList {
			ping := make(chan Ping)
			ping.url = website
			go doPing(ping)
			writeToDB(db, ping)	
		}
		
	}
}

func main() {

	var l Layout

	l.x_range = 40
	l.y_range = 25

	var db MysqlServer
	db.url = "http://localhost/"
	db.user = "pingtastic"
	db.password = "pingtastic"
	db.database = "pingtastic"

	argsWithProg := os.Args
	if os.Args[1] == "getAlexa" {
		getAlexaList(db)
	} else {
		runServer(db)
	}

}

