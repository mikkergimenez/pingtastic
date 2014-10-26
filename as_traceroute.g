package main

import (
    "bufio"
    "fmt"
    //"io"
    "net"
    "os"
    "os/exec"
    "regexp"
    "strings"
    //"time"
)

func main() {
    fmt.Printf("About to Run Traceroutes \n")
    readFile("oix-full-snapshot-latest.dat")
    //oix-full-snapshot-latest.dat
    //time.Sleep(10 * time.Second)
    ipChan := make(chan []string)
    //asChan := make(chan []string)
    waitChan := make(chan int)
    maxPerIteration := 100;
    iteration := 4000
    waitCount := maxPerIteration // * iteraion
    currCount := 0
    var tablePos int = len(routingTableList)

    for j := 0; j < iteration; j++ {
        for i := tablePos; i > tablePos - maxPerIteration; i-- {
            go runTraceroute(routingTableList[i - 1], ipChan, waitChan)
            //fmt.Printf("%s\n", routingTableList[i - 1])
        }
        for i := 0; i < waitCount - 10; i++ {
            fmt.Printf("%d\n", currCount)
            traceRouteRaw := <- ipChan
            fmt.Printf("%q\n", traceRouteRaw)
            go makeAsPath(traceRouteRaw)
            //fmt.Printf("%q\n", <- asChan)
            currCount += <- waitChan
        }
        //time.Sleep(10 * time.Second)
        tablePos = tablePos - maxPerIteration
    }

    for i := 0; i < iteration * 10; i++ {
        fmt.Printf("%d\n", currCount)
        traceRouteRaw := <- ipChan
        fmt.Printf("%q\n", traceRouteRaw)
        go makeAsPath(traceRouteRaw)
        //fmt.Printf("%q\n", <- asChan)
        currCount += <- waitChan
    }

    close(ipChan)
    close(waitChan)
    //fmt.Printf("%q\n", routingTableList)
    fmt.Printf("Traceroutes Done \n")
}

var routingTableList = make([]string, 1)
var prevAsn string

func processResults(c chan []string) {}

func makeAsPath(traceRouteRaw []string) {
    var asPath = make([]string, 0)
    prevAsn := ""
    for i := 0; i < len(traceRouteRaw); i++ {
        tempAsString := strings.Split(traceRouteRaw[i], ",")
        asn := tempAsString[2]
        if asn == prevAsn {
            //do nothing
        } else {
            asPath = append(asPath, asn)
        }
        prevAsn = asn

    }
    fmt.Printf("%q\n", asPath)
    //asChan <- asPath
    return
}

func runTraceroute(targetHost string, ipChan chan []string, waitChan chan int) {
    app := "traceroute"
    arg0 := "-f" //-f 2
    arg1 := "1" //first ttl
    arg2 := "-m"
    arg3 := "25" //max ttl
    arg4 := "-q"
    arg5 := "3" //number of tries per hop
    arg6 := "-w"
    arg7 := "2" //wait time (in seconds)
    arg8 := "-n" //no dns resolution
    arg9 := "64" //packet size, bytes

    cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, targetHost, arg9) //execute the commands above, with the host passed
    out, err := cmd.Output()                                                       //output the results of the command

    if err != nil {
        fmt.Printf(err.Error())
        return
    }
    //fmt.Printf("%s\n", string(out))
    traceResult := findIPaddr(string(out))
    //fmt.Printf("%q\n", traceResult)
    var ipArray = make([]string, 0)
    dnsPtr := ""
    asnResult := ""
    for i := 2; i < len(traceResult); i++ {
        currIpAddr := traceResult[i][0]
        reverseDNSaddr, _ := net.LookupAddr(currIpAddr)
        if len(reverseDNSaddr) > 0 {
            //fmt.Printf("%s\n", reverseDNSaddr[0])
            dnsPtr = reverseDNSaddr[0] //if it has a PTR, put it in the PTR array section
        } else {
            dnsPtr = "noPTR" //everything that does not have a PTR listed
        }

        reverseIPaddrResult := reverseIPaddr(currIpAddr)                 //call the function to reverse the IP address
        dnsQueryAddress := reverseIPaddrResult + ".origin.asn.cymru.com" //query team cymru database for IP to ASN
        ipToAsnLookup, _ := net.LookupTXT(dnsQueryAddress)               //get TXT record result
        if len(ipToAsnLookup) > 0 {
            asnIndex := strings.Index(ipToAsnLookup[0], " ") //if so, parse out the ASN from the result
            queryString := ipToAsnLookup[0]
            asnResult = queryString[0:asnIndex]
        } else {
            asnResult = "noASN"
        }

        ipRowString := currIpAddr + "," + dnsPtr + "," + asnResult 
        ipArray = append(ipArray, ipRowString)
        
    }
    //fmt.Printf("%d\n", len(traceIpArray))
    //fmt.Printf("%q\n", ipArray)
    ipChan <- ipArray
    waitChan <- 1
    return
}

func reverseIPaddr(ip string) string {
    ipOctetArray := strings.Split(ip, ".")
    var reversedOctetArray [4]string
    count := 0
    for i := range ipOctetArray {
        reversedOctetArray[count] = ipOctetArray[len(ipOctetArray)-1-i]
        count++
    }
    reversedIPaddr := reversedOctetArray[0] + "." + reversedOctetArray[1] + "." + reversedOctetArray[2] + "." + reversedOctetArray[3]
    return reversedIPaddr
}

func findIPaddr(inputText string) [][]string {
    re, err := regexp.Compile(`[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`) //basic IP match in regex. All IPS in traceroute should be valid
    if err != nil {
        fmt.Printf(err.Error())
    }
    ipArray := re.FindAllStringSubmatch(inputText, -1)
    return ipArray
}

var prevIpAddr string

func buildIpArray(ipAddr string) []string {
    if ipAddr == prevIpAddr {
        return routingTableList
    }
    prevIpAddr = ipAddr
    //fmt.Printf("%s\n", ipAddr)
    routingTableList = append(routingTableList, ipAddr)
    //fmt.Printf("%q\n", routingTableList)
    return routingTableList
}

func parseFileLine(fileLine string) {
    ipArray := findIPaddr(fileLine)
    if len(ipArray) > 0 {
        ipString := ipArray[0][0]
        //fmt.Printf("%s\n", ipString)
        buildIpArray(ipString)
    } else {
        fmt.Printf("err_empty_array\n")
    }
}

func readFile(fileName string) {
    fi, err := os.Open(fileName)
    if err != nil {
        fmt.Printf(err.Error())
    }
    scanner := bufio.NewScanner(fi)
    for scanner.Scan() {
        parseFileLine(scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        fmt.Printf(err.Error())
    }
}

/*

func writeLines(line string, path string) { 
    f, err := os.OpenFile(path, os.O_APPEND, 0666)
    if err != nil {
        fmt.Printf(err.Error())
    } 
    _, err = io.WriteString(f, line)
    if err != nil {
        fmt.Printf(err.Error())
    } 
    f.Close()
    return
}
*/