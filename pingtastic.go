/*
 * A Program that reads data from the alexa top 1000 sites and pings them periodically.
 *
 * This program can be run either as a server:
 *
 * ./pingtastic server
 *
 * Periodically however, and prior to first run, you will want to download 
 * the Alexa Top 1000 and import the data to the databse.
 * 
 * This will be done by running:
 *
 * ./pingtastic getAlexa
 * 
 */ 
package main

import (
	// "net/http"
	"fmt"
	//"github.com/tatsushid/go-fastping"
	"database/sql"
	"github.com/aeden/traceroute"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net"
	"os"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	//"syscall"
	//"os/signal"
)

type Configuration struct {
	DBUser		string
	DBPass		string
	DBDatabase	string
	DBHost		string
	DBPort		string
	AlexaURL	string
}

type MysqlServer struct {
	url      string
	user     string
	password string
	database string
}

type Ping struct {
	urlid  string
	rtt    int
	url    string
	ipaddr string
}

type Website struct {
	Url       string
	Name      string
	PingValue int
}


func doPing(pingTime Ping) int {

	start := time.Now()
	_, err := net.Dial("tcp", pingTime.ipaddr+":80")
	if err != nil {
		// handle error
	}
	elapsed := time.Since(start)

	// fmt.Println(pingTime.ipaddr, " ", elapsed, " ", elapsed.Seconds(), " ", int(elapsed))

	return_time := int(elapsed.Seconds() * 1000)

	return return_time
}

func get_day_average(db *sql.DB, ping Ping, hour int) {
	var day_query string 
	var update_day_query string

	day_query = "SELECT latency FROM pings WHERE time_of >= now() - INTERVAL 1 DAY and url_id="+ping.urlid+";"
    rows, err := db.Query(day_query)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    total := 0
    count := 0
    for rows.Next() {
        var latency int
        if err := rows.Scan(&latency); err != nil {
            log.Fatal(err)
        }
        total += latency
        count += 1
    }

    var day_average int
    if count > 0 {
    	day_average = total / count
    } else {
    	day_average = 0
    }

    update_day_query = "INSERT INTO latency_averages (url_id, past_24_hours) VALUES ("+ping.urlid+", "+strconv.Itoa(day_average)+") ON DUPLICATE KEY UPDATE past_24_hours="+strconv.Itoa(day_average)
    
    _, err = db.Exec(update_day_query)
    if err != nil {
        fmt.Printf("Error (get_day_average) : ", err)
    }

	// This function calculates the average latency for the past 24 hours
}

func get_hour_average(db *sql.DB, ping Ping, hour int) {
	// This function calculates the average latency for this hour of the day
	var hour_query string
	var update_hour_query string
	hour_query = "SELECT latency FROM pings WHERE hour_of="+strconv.Itoa(hour)+" and url_id="+ping.urlid+";"
    rows, err := db.Query(hour_query)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    total := 0
    count := 0
    for rows.Next() {
        var latency int
        if err := rows.Scan(&latency); err != nil {
            log.Fatal(err)
        }
        total += latency
        count += 1
    }

    var hour_average int
    if count > 0 {
    	hour_average = total / count
    } else {
    	hour_average = 0
    }

    update_hour_query = "INSERT INTO latency_averages (url_id, hour_"+strconv.Itoa(hour)+") VALUES ("+ping.urlid+", "+strconv.Itoa(hour_average)+") ON DUPLICATE KEY UPDATE hour_"+strconv.Itoa(hour)+"="+strconv.Itoa(hour_average)
    _, err = db.Exec(update_hour_query)
    
    if err != nil {
        fmt.Printf("Error (get_hour_average) : ", err)
    }


}

func writeToDB(db *sql.DB, ping Ping) {
	var dbquery string
	var urlquery string
	var url_update_query string
	var urlid int
	var current_hour int

	urlquery = "SELECT id from urls where ip ='" + ping.ipaddr + "'"
	err := db.QueryRow(urlquery).Scan(&urlid)

	const layout = "2014-12-30 03:04:58"

	current_hour = time.Now().Hour()
	dbquery = "INSERT INTO pings (url_id, latency, time_of, hour_of) VALUES ('" + strconv.Itoa(urlid) + "', '" + strconv.Itoa(ping.rtt) + "', NOW(), '"+strconv.Itoa(current_hour)+"')"
	
	_, err = db.Exec(dbquery)
	if err != nil {
		log.Fatal(err)
	}

	url_update_query = "UPDATE urls SET latency='" + strconv.Itoa(ping.rtt) + "' WHERE id='" + strconv.Itoa(urlid) + "'"

	_, err = db.Exec(url_update_query)
	if err != nil {
		log.Fatal(err)
	}

	get_hour_average(db, ping, current_hour)
	get_day_average(db, ping, current_hour)

}

func printHop(db *sql.DB, hop traceroute.TracerouteHop, ping Ping) {
	var dbquery string
	var urlquery string
	var urlid int

	urlquery = "SELECT id from urls where ip ='" + ping.ipaddr + "'"
	err := db.QueryRow(urlquery).Scan(&urlid)

	addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
	if hop.TTL < 31 {
		dbquery = "INSERT INTO traceroutes (url_id, hop" + strconv.Itoa(hop.TTL) + ", name" + strconv.Itoa(hop.TTL) + ", updated ) VALUES ('" + strconv.Itoa(urlid) + "', '" + addr + "', '" + hop.Host + "', NOW()) ON DUPLICATE KEY UPDATE hop" + strconv.Itoa(hop.TTL) + "='" + addr + "', updated=NOW(), name" + strconv.Itoa(hop.TTL) + "='" + hop.Host + "'"
		_, err = db.Exec(dbquery)
		if err != nil {
			log.Fatal(err)
		}
	}
	swapSort(urlid, db)
}

func address(address [4]byte) string {
	return fmt.Sprintf("%v.%v.%v.%v", address[0], address[1], address[2], address[3])
}

func doTraceroute(db *sql.DB, ping Ping) {
	/*
		var m = flag.Int("m", traceroute.DEFAULT_MAX_HOPS, `Set the max time-to-live (max number of hops) used in outgoing probe packets (default is 64)`)
		var q = flag.Int("q", 1, `Set the number of probes per "ttl" to nqueries (default is one probe).`)

		flag.Parse()
		host := flag.Arg(0)
		options := traceroute.TracerouteOptions{}

		options.SetRetries(*q - 1)
		options.SetMaxHops(*m + 1)
	*/
	options := traceroute.TracerouteOptions{}
	options.SetRetries(0)
	options.SetMaxHops(30)

	ipAddr := ping.ipaddr

	// fmt.Printf("traceroute to %v, %v hops max, %v byte packets\n", ipAddr, options.MaxHops(), options.PacketSize())

	c := make(chan traceroute.TracerouteHop, 0)
	go func() {
		for {
			hop, ok := <-c
			if !ok {
				fmt.Println()
				return
			}
			printHop(db, hop, ping)
		}
	}()

	_, err := traceroute.Traceroute(ipAddr, &options, c)
	if err != nil {
		fmt.Printf("Error (doTraceroute) : ", err)
	}
}

func getDBCon(db MysqlServer) *sql.DB {
	dbcon, err := sql.Open("mysql", db.user+":"+db.password+"@"+db.url+"/"+db.database)
	if err != nil {
		log.Fatal(err)
	}
	return dbcon
}

func calculatePath(db *sql.DB, ping Ping) {
	var path1 string
	var path2 string
	var path3 string

	names := make([]interface{}, 30)
	container := make([]string, 30)
	for i, _ := range container {
		names[i] = &container[i]
	}

	urlquery := "SELECT name1, name2, name3, name4, name5, name6, name7, name8, name9, name10, name11, name12, name13, name14, name15, name16, name17, name18, name19, name20, name21, name22, name23, name24, name25, name26, name27, name28, name29, name30 FROM traceroutes WHERE url_id=" + ping.urlid
	err := db.QueryRow(urlquery).Scan(names...)
	if err != nil {
		// fmt.Printf("Error: ", err)
	}

	p := 0
	for _, name := range container {
		if strings.ContainsRune(name, '.') {
			full_domain_name := strings.Split(name, ".")
			domain_name := full_domain_name[len(full_domain_name)-3] + "." + full_domain_name[len(full_domain_name)-2]
			if p == 0 {
				path1 = domain_name
				p++
			} else if p == 1 && domain_name != path1 {
				path2 = domain_name
				p++
			} else if p == 2 && domain_name != path2 {
				path3 = domain_name
				break
			}
		}
	}
	updateQuery := "UPDATE urls SET path1='" + path1 + "', path2='" + path2 + "', path3='" + path3 + "' WHERE id=" + ping.urlid
	_, err = db.Exec(updateQuery)

	if err != nil {
		fmt.Printf("Error (calculate_path) : ", err)
	}
}

func runServer(db MysqlServer) {

	dbcon := getDBCon(db)

	fmt.Println("Server Started")
	var alexaList [1001]string
	alexaList = getAlexaList(db)
	var x int = 0
	var y int = 0
	for _ = range time.Tick(1 * time.Second) {

		for i, ip := range alexaList {
			fmt.Println("Doing ", i)
			// var ping chan Ping = make(chan Ping)
			var ping Ping
			ping.urlid = strconv.Itoa(i)
			ping.ipaddr = ip
			go func() {
				//var fastPingResult *fastping.Pinger
				fastPingResult := doPing(ping)
				ping.rtt = fastPingResult

				if x == y {
					doTraceroute(dbcon, ping)
					calculatePath(dbcon, ping)
				}

				//ping.rtt = int(fastPingResult.MaxRTT)
				writeToDB(dbcon, ping)
			}()
			time.Sleep(300 * time.Millisecond)
			if x > 9 {
				x = 0
			} else {
				x++
			}

		}
		if y > 9 {
			y = 0
		} else {
			y++
		}
		time.Sleep(300 * time.Second)
	}
}

func get_config() Configuration {
	file, _ := os.Open("conf.json")
	defer file.Close()
	fmt.Println(file)
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
  		fmt.Println("error:", err)
	}
	fmt.Println(configuration)
	return configuration	
}

func main() {

	config := get_config()
	fmt.Println(config)

	var db MysqlServer
	db.url = "tcp("+config.DBHost+":"+config.DBPort+")"
	db.user = config.DBUser
	db.password = config.DBPass
	db.database = config.DBDatabase

	//argsWithProg := os.Args
	if os.Args[1] == "getAlexa" {
		downloadAlexaList(db)
	} else {
		runServer(db)
	}

}
