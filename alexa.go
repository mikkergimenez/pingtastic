package main

import (
    "archive/zip"
    "log"
    "io"
    "database/sql"
    "net/http"
    "net"
    "strconv"
    "strings"
    "bufio"
    "bytes"
	"os"
	"fmt"
    _ "github.com/go-sql-driver/mysql"
)

func writeAlexaToDB(db MysqlServer, url string, ip string, order int) {
    dbcon, err := sql.Open("mysql", db.user+":"+db.password+"@"+db.url+"/"+db.database)

    if err != nil {
        log.Fatal(err)
    }
    defer dbcon.Close()

    var dbquery string
    
    dbquery = "INSERT IGNORE INTO urls (text, ip, host_order, latency) VALUES ('"+url+"', '"+ip+"', '"+strconv.Itoa(order)+"', '0')"
    _, err = dbcon.Exec(dbquery)

    if err != nil {
        log.Fatal(err)
    }

    // rows.Close()
}

func readCSVFile(csvFile string, db MysqlServer) {
    file, err := os.Open(csvFile)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    
    fmt.Println("Starting CSV Read")

    x := 0
    for scanner.Scan() {
        if x < 1000 {
            // fmt.Println(scanner.Text()) // Println will add back the final '\n'
            url := strings.Split(scanner.Text(), ",")
            fmt.Println(url)
            ra, err := net.ResolveIPAddr("ip4:icmp", url[1])

            if err != nil {
                fmt.Println(err)
            } else {

                writeAlexaToDB(db, url[1], ra.String(), x)
                x++
            }
        } else {
            break
        }
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "reading standard input:", err)
    }
}

func unzipAlexaZip(zipFile string, csvFile string) {
    var buf bytes.Buffer

    logger := log.New(&buf, "logger: ", log.Lshortfile)
    
    reader, err := zip.OpenReader(zipFile)
    
    if err != nil {
        logger.Fatal(err)
    }
    defer reader.Close()

    writer, err := os.Create(csvFile)
    if err != nil {
         logger.Fatal(err)
    }

    defer writer.Close()

    for _, f := range reader.File {
        rc, err := f.Open()
        if err != nil {
            logger.Fatal(err)
        }

        if _, err = io.Copy(writer, rc); err != nil {
            logger.Fatal(err)
        }

        rc.Close()

        fmt.Println("File uncompressed\n")
    }
}

func downloadFromUrl(url string) string {
    tokens := strings.Split(url, "/")
    fileName := tokens[len(tokens)-1]
    fmt.Println("Downloading", url, "to", fileName)

    // TODO: check file existence first with io.IsExist
    output, err := os.Create(fileName)
    if err != nil {
        fmt.Println("Error while creating", fileName, "-", err)
        return "error"
    }
    defer output.Close()

    response, err := http.Get(url)
    if err != nil {
        fmt.Println("Error while downloading", url, "-", err)
        return "error"
    }
    defer response.Body.Close()

    n, err := io.Copy(output, response.Body)
    if err != nil {
        fmt.Println("Error while downloading", url, "-", err)
        return "error"
    }

    fmt.Println(n, "bytes downloaded.")

    return fileName
}

func readFromAlexaDB(db MysqlServer) [1001]string {
    fmt.Println("Getting from DB")
    dbcon, err := sql.Open("mysql", db.user+":"+db.password+"@"+db.url+"/"+db.database)
    if err != nil {
        log.Fatal(err)
    }
    defer dbcon.Close()

    var dbquery string
    dbquery = "SELECT id, ip FROM urls LIMIT 1000"
    rows, err := dbcon.Query(dbquery)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    var alexaList [1001]string

    //i := 0
    for rows.Next() {
        var text string
        var id int
        if err := rows.Scan(&id, &text); err != nil {
            log.Fatal(err)
        }
        alexaList[id] = text
        //i++
    }
    return alexaList
}

func getAlexaList(db MysqlServer) [1001]string {
    return readFromAlexaDB(db)
}

func downloadAlexaList(db MysqlServer) {
    url := "http://s3.amazonaws.com/alexa-static/top-1m.csv.zip"
    var file string
    file = downloadFromUrl(url)
    if file != "error" {
        csvFile := "top-1m.csv"
        unzipAlexaZip(file, csvFile)
        readCSVFile(csvFile, db)
    }
}
