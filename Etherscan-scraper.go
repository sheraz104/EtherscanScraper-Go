package main

import (
    "net/http"
    "log"
    "github.com/PuerkitoBio/goquery"
    "fmt"
    "strings"
    "strconv"
    "github.com/ethereum/go-ethereum/ethclient"
    "time"
    "context"
    "math/big"

)

type obj struct {
    sender string `json:sender`
    receiver string `json:receiver`
    value string `json:value`
    timestamp string `json:timestamp`
}

var data[] obj

func main() {

    getPage("", 1, 0, 0)

    firstStepLength: = len(data)

    for i: = 0;
    i < firstStepLength;
    i++{
        getPage(data[i].sender, 2, 0, 0)
    }

    secondStepLength: = len(data)

    for p: = firstStepLength;
    p < secondStepLength;
    p++{
        getPage(data[p].sender, 3, 0, 0)
    }

    for q: = 0;
    q < len(data);
    q++{
        getLastTransaction(data[q].sender, data[q].receiver, q, 0)
    }

    client, err: = ethclient.Dial("https://mainnet.infura.io/QWMgExFuGzhpu2jUr6Pq")
    if err != nil {
        log.Fatal(err)
    }

    for r: = 0;
    r < len(data);
    r++{

        convertedBlock, _: = strconv.ParseInt(data[r].timestamp, 10, 64)
        blockNumber: = big.NewInt(convertedBlock)
        block, err: = client.BlockByNumber(context.Background(), blockNumber)
        if err != nil {
            log.Fatal(err)
        }

        tm: = time.Unix(block.Time().Int64(), 0)
        data[r].timestamp = tm.String()
    }

    fmt.Println(data)
}


func getLastTransaction(addressFrom string, addressTo string, index int, page int) {
    URL: = func() string {
        if page > 0 {
            return "https://etherscan.io/txs?a=" + addressTo + "&p=" + strconv.Itoa(page)
        } else {
            return "https://etherscan.io/txs?a=" + addressTo
        }
    }()

        res,
    err: = http.Get(URL)

    if err != nil {
        log.Fatal(err)
    }

    defer res.Body.Close()

    if res.StatusCode != 200 {
        log.Fatal("Request failed: ", res.StatusCode)
    }


    transactionFound: = false;
    matchFirst: = false;

    doc,
    err: = goquery.NewDocumentFromReader(res.Body)

        doc.Find("table.table.table-hover tbody tr").Each(func(i int, row * goquery.Selection) {

        var value, timestamp string
        IN: = strings.TrimSpace(row.Find("span.label.label-success.rounded").Text())
        if IN == "IN" {

            row.Find("td").Each(func(i int, elem * goquery.Selection) {

                if transactionFound {
                    return;
                }

                if i == 1 {
                    timestamp = strings.TrimSpace(elem.Text());
                }

                if i == 3 {
                    if strings.ToLower(addressFrom) == strings.ToLower(strings.TrimSpace(elem.Text())) {
                        matchFirst = true;
                    }
                }

                if i == 6 {
                    if matchFirst {
                        value = strings.TrimSpace(elem.Text())
                        data[index].timestamp = timestamp;
                        data[index].value = value
                        transactionFound = true;
                    }
                }
            })
        }
    })

        nextPageExists,
    _: = doc.Find("a.btn.btn-default.btn-xs.logout").Attr("href");

    if !transactionFound && len(nextPageExists) > 0 {
        page++
        if page == 1 {
            page++
        }
        getLastTransaction(addressFrom, addressTo, index, page);
    }

}


func getPage(address string, degree int, page int, count int) {
    URL: = func() string {
        if page > 0 {
            return "https://etherscan.io/txs?a=" + address + "&p=" + strconv.Itoa(page)
        } else {
            return "https://etherscan.io/txs?a=" + address
        }
    }()

        fmt.Println("Visiting URL: ", URL)
    res,
    err: = http.Get(URL)

    if err != nil {
        log.Fatal(err)
    }

    defer res.Body.Close()

    if res.StatusCode != 200 {
        log.Fatal("Request failed: ", res.StatusCode)
    }

    doc,
    err: = goquery.NewDocumentFromReader(res.Body)

        doc.Find("table.table.table-hover tbody tr").Each(func(i int, row * goquery.Selection) {

        var sender, receiver string
        IN: = strings.TrimSpace(row.Find("span.label.label-success.rounded").Text())
        if IN == "IN" {

            row.Find("span.address-tag").Each(func(i int, elem * goquery.Selection) {

                if count >= 5 {
                    return;
                }

                if i == 1 {
                    sender = strings.TrimSpace(elem.Text());
                }

                if i == 2 {
                    receiver = strings.TrimSpace(elem.Text());
                    data = append(data, obj {
                        sender: sender,
                        receiver: receiver
                    })
                    count++;
                }
            })
        }
    })

        nextPageExists,
    _: = doc.Find("a.btn.btn-default.btn-xs.logout").Attr("href");

    if count < 5 && len(nextPageExists) > 0 {
        page++
        if page == 1 {
            page++
        }
        getPage(address, degree, page, count);
    }

}