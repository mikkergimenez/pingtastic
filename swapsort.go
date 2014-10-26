/*
 * This is a highly inefficient but adaptive sort process to attempt to put urls which are "close" to each other
 * in order, so patterns emerge in the grid.  Swapsort looks at a URL and the one earlier in the sort, and if the 
 * current URL has a first, second or third IP hop as being different, it will swap them.  The order in which they
 * are sorted, will be relatively random, as there is no correlation between lower IP Addresses and anything.
 *
 * Though this is currently sorting on the first three hops, it would likely be better to sort on first three
 * "AS" hops(where AS in this case would be the domain name of the hop)  That would be an easy fix, as we already
 * have the "AS Path" data kept with the URL.
 */

package main

import (
    "fmt"
    "strconv"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)


func do_swap(pre_urlid int, urlid int, order int, dbcon *sql.DB) {
    /*
     * If one URL has a lower IP address than the other URL, swap them.
     */
    var updateQuery string

    updateQuery = "UPDATE urls SET host_order=NULL WHERE id="+strconv.Itoa(pre_urlid)
    _, err := dbcon.Exec(updateQuery)
    if err != nil {
        fmt.Printf("Error (do_swap 1) : ", err)
    }

    updateQuery = "UPDATE urls SET host_order=NULL WHERE id="+strconv.Itoa(urlid)
    _, err = dbcon.Exec(updateQuery)
    if err != nil {
        fmt.Printf("Error (do_swap 2) : ", err)
    }

    updateQuery = "UPDATE urls SET host_order="+strconv.Itoa(order)+" WHERE id="+strconv.Itoa(pre_urlid)
    _, err = dbcon.Exec(updateQuery)
    if err != nil {
        fmt.Printf("Error (do_swap 3) : ", err)
    }

    updateQuery = "UPDATE urls SET host_order="+strconv.Itoa(order - 1)+" WHERE id="+strconv.Itoa(urlid)
    _, err = dbcon.Exec(updateQuery)
    if err != nil {
        fmt.Printf("Error (do_swap 4) : ", err)
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
            fmt.Printf("Error (swapSort 1) : ", err)
        }

        updateQuery = "UPDATE urls SET host_order="+strconv.Itoa(next_order)+" WHERE id="+strconv.Itoa(urlid)
        _, err = dbcon.Exec(updateQuery)
        if err != nil {
            fmt.Printf("Error (swapSort 2) : ", err)
        }

        return
    } else if host_order == 1 {
        return
    } else {

        urlquery = "SELECT id from urls where host_order="+strconv.Itoa(host_order - 1)+" LIMIT 1"
        err = dbcon.QueryRow(urlquery).Scan(&pre_urlid)
        if err != nil {
            fmt.Printf("Error (swapSort 3) : ", err)
        }

        urlquery = "SELECT hop1, hop2, hop3 from traceroutes where url_id="+strconv.Itoa(urlid)
        err = dbcon.QueryRow(urlquery).Scan(&pre_hop1, &pre_hop2, &pre_hop3)
        if err != nil {
            fmt.Printf("Error (swapSort 4) : ", err)
        }

        urlquery = "SELECT hop1, hop2, hop3 from traceroutes where url_id="+strconv.Itoa(pre_urlid)
        err = dbcon.QueryRow(urlquery).Scan(&hop1, &hop2, &hop3)
        if err != nil {
            fmt.Printf("Error (swapSort 5) : ", err)
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

