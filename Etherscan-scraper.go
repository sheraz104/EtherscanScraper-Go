package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
)

type obj struct {
	Sender             string `json:sender`
	Receiver           string `json:receiver`
	Value              string `json:value`
	Timestamp          string `json:timestamp`
	DegreeOfSeparation int    `json:degreeOfSeparation`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/{masterWallet}", handler).Methods("GET")
	http.ListenAndServe(":9090", router)
}

func handler(w http.ResponseWriter, r *http.Request) {
	var data []obj
	vars := mux.Vars(r)

	masterWallet := vars["masterWallet"]

	c0 := make(chan bool)
	go getPage(masterWallet, 1, 0, 0, &data, c0)
	<-c0

	firstStepLength := len(data)
	c1 := make(chan bool)

	for i := 0; i < firstStepLength; i++ {
		go getPage(data[i].Sender, 2, 0, 0, &data, c1)
		// time.Sleep(150 * time.Millisecond)
	}

	for GR1 := 0; GR1 < firstStepLength; GR1++ {
		<-c1
	}

	secondStepLength := len(data)
	c2 := make(chan bool)
	for p := firstStepLength; p < secondStepLength; p++ {
		go getPage(data[p].Sender, 3, 0, 0, &data, c2)
	}

	for GR2 := firstStepLength; GR2 < secondStepLength; GR2++ {
		<-c2
	}

	c3 := make(chan bool)
	for q := 0; q < len(data); q++ {
		go getLastTransaction(data[q].Sender, data[q].Receiver, q, 0, data, c3)

		time.Sleep(200 * time.Millisecond)

	}

	for GR3 := 0; GR3 < len(data); GR3++ {
		<-c3
	}

	c4 := make(chan bool)
	for r := 0; r < len(data); r++ {
		go getTimestamps(data, r, c4)
		// if r%2 == 0 {
		// 	time.Sleep(500 * time.Millisecond)
		// }
	}

	for GR4 := 0; GR4 < len(data); GR4++ {
		<-c4
	}

	returnData := data
	data = nil
	json.NewEncoder(w).Encode(&returnData)
}

func getLastTransaction(addressFrom string, addressTo string, index int, page int, data []obj, c chan bool) {
	URL := func() string {
		if page > 0 {
			return "https://etherscan.io/txs?a=" + addressTo + "&p=" + strconv.Itoa(page)
		} else {
			return "https://etherscan.io/txs?a=" + addressTo
		}
	}()

	res, err := http.Get(URL)

	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatal("Request failed: ", res.StatusCode)
	}

	transactionFound := false
	matchFirst := false

	doc, err := goquery.NewDocumentFromReader(res.Body)

	doc.Find("table.table.table-hover tbody tr").Each(func(i int, row *goquery.Selection) {

		var Value, Timestamp string
		IN := strings.TrimSpace(row.Find("span.label.label-success.rounded").Text())
		if IN == "IN" {

			row.Find("td").Each(func(i int, elem *goquery.Selection) {

				if transactionFound {
					c <- true
					return
				}

				if i == 1 {
					Timestamp = strings.TrimSpace(elem.Text())
				}

				if i == 3 {
					if strings.ToLower(addressFrom) == strings.ToLower(strings.TrimSpace(elem.Text())) {
						matchFirst = true
					}
				}

				if i == 6 {
					if matchFirst {
						Value = strings.TrimSpace(elem.Text())
						data[index].Timestamp = Timestamp
						data[index].Value = Value
						fmt.Println(Value)
						transactionFound = true
					}
				}
			})
		}
	})

	nextPageExists, _ := doc.Find("a.btn.btn-default.btn-xs.logout").Attr("href")

	if !transactionFound && len(strings.TrimSpace(nextPageExists)) > 0 {
		page++
		if page == 1 {
			page++
		}
		go getLastTransaction(addressFrom, addressTo, index, page, data, c)
	}

}

func getPage(address string, degree int, page int, count int, data *[]obj, c chan bool) {
	URL := func() string {
		if page > 0 {
			return "https://etherscan.io/txs?a=" + address + "&p=" + strconv.Itoa(page)
		} else {
			return "https://etherscan.io/txs?a=" + address
		}
	}()

	fmt.Println("Visiting URL: ", URL)
	res, err := http.Get(URL)

	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatal("Request failed: ", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)

	doc.Find("table.table.table-hover tbody tr").Each(func(i int, row *goquery.Selection) {

		var Sender, Receiver string
		IN := strings.TrimSpace(row.Find("span.label.label-success.rounded").Text())
		if IN == "IN" {

			row.Find("span.address-tag").Each(func(i int, elem *goquery.Selection) {

				if count >= 5 {
					c <- true
					return
				}

				if i == 1 {
					Sender = strings.TrimSpace(elem.Text())
				}

				if i == 2 {
					Receiver = strings.TrimSpace(elem.Text())
					*data = append(*data, obj{Sender: Sender, Receiver: Receiver, DegreeOfSeparation: degree})
					count++
				}
			})
		}
	})

	nextPageExists, _ := doc.Find("a.btn.btn-default.btn-xs.logout").Attr("href")
	if count < 5 && len(strings.TrimSpace(nextPageExists)) > 0 {
		page++
		if page == 1 {
			page++
		}
		go getPage(address, degree, page, count, data, c)
	}

}

func getTimestamps(data []obj, r int, c chan bool) {

	client, err := ethclient.Dial("https://mainnet.infura.io/QWMgExFuGzhpu2jUr6Pq")
	if err != nil {
		log.Fatal(err)
	}
	convertedBlock, _ := strconv.ParseInt(data[r].Timestamp, 10, 64)
	blockNumber := big.NewInt(convertedBlock)
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal(err)
	}

	tm := time.Unix(block.Time().Int64(), 0)
	data[r].Timestamp = tm.String()
	c <- true
}
