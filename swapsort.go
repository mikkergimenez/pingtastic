package main

import (
    "fmt"
    "strconv"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)


func do_swap(pre_urlid int, urlid int, order int, dbcon *sql.DB) {
    var updateQuery string

    updateQuery = "UPDATE urls SET host_order=NULL WHERE id="+strconv.Itoa(pre_urlid)
    _, err := dbcon.Exec(updateQuery)
    if err != nil {
        fmt.Printf("Error1: ", err)
    }

    updateQuery = "UPDATE urls SET host_order=NULL WHERE id="+strconv.Itoa(urlid)
    _, err = dbcon.Exec(updateQuery)
    if err != nil {
        fmt.Printf("Error2: ", err, "\n")
    }

    updateQuery = "UPDATE urls SET host_order="+strconv.Itoa(order)+" WHERE id="+strconv.Itoa(pre_urlid)
    _, err = dbcon.Exec(updateQuery)
    if err != nil {
        fmt.Printf("Error3: ", err)
    }

    updateQuery = "UPDATE urls SET host_order="+strconv.Itoa(order - 1)+" WHERE id="+strconv.Itoa(urlid)
    _, err = dbcon.Exec(updateQuery)
    if err != nil {
        fmt.Printf("Error4: ", err, "\n")
    }
}

func swapSort(urlid int, dbcon *sql.DB) {
    var host_order int
    var next_order int
    var pre_urlid int
    var urlquery string
    var unusedQuery string
    var updateQuery string
    var hop1 string
    var hop2 string
    var hop3 string
    var pre_hop1 string
    var pre_hop2 string
    var pre_hop3 string

    urlquery = "SELECT host_order from urls where id="+strconv.Itoa(urlid)
    
    err := dbcon.QueryRow(urlquery).Scan(&host_order)
    
    if err != nil {
        // fmt.Println("Null Order, setting to "+strconv.Itoa(host_order))
    }
    
     if host_order <= 0 {
        unusedQuery = `
        SELECT min(unused) AS unused
        FROM (
            SELECT MIN(t1.host_order)+1 as unused
            FROM urls AS t1
            WHERE NOT EXISTS (SELECT * FROM urls AS t2 WHERE t2.host_order = t1.host_order+1)
            UNION
            SELECT 1
            FROM DUAL
            WHERE NOT EXISTS (SELECT * FROM urls WHERE host_order=1)
        ) AS subquery
        `
        err = dbcon.QueryRow(unusedQuery).Scan(&next_order)

        if err != nil {
            fmt.Printf("Error111: ", err)
        }

        updateQuery = "UPDATE urls SET host_order="+strconv.Itoa(next_order)+" WHERE id="+strconv.Itoa(urlid)
        _, err = dbcon.Exec(updateQuery)
        if err != nil {
            fmt.Printf("Error: ", err)
        }

        return
    } else if host_order == 1 {
        return
    } else {

        urlquery = "SELECT id from urls where host_order="+strconv.Itoa(host_order - 1)+" LIMIT 1"
        err = dbcon.QueryRow(urlquery).Scan(&pre_urlid)
        if err != nil {
            fmt.Printf("Error:222 ", err)
        }

        urlquery = "SELECT hop1, hop2, hop3 from traceroutes where url_id="+strconv.Itoa(urlid)
        err = dbcon.QueryRow(urlquery).Scan(&pre_hop1, &pre_hop2, &pre_hop3)
        if err != nil {
            fmt.Printf("Error333: ", err)
        }

        urlquery = "SELECT hop1, hop2, hop3 from traceroutes where url_id="+strconv.Itoa(pre_urlid)
        err = dbcon.QueryRow(urlquery).Scan(&hop1, &hop2, &hop3)
        if err != nil {
            fmt.Printf("Error444: ", err)
        }

        if hop1 < pre_hop1 {
            do_swap(pre_urlid, urlid, host_order, dbcon)
        } else if hop1 == pre_hop1 && hop2 < pre_hop2 {
            do_swap(pre_urlid, urlid, host_order, dbcon)
        } else if hop1 == pre_hop1 && hop2 == pre_hop2 && hop3 < pre_hop3 {
            do_swap(pre_urlid, urlid, host_order, dbcon)
        } else {
            return
        }
    }
}

