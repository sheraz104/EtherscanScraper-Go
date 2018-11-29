package main

import (
	"github.com/gorilla/mux"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
	"encoding/json"
	"context"
)

type obj struct {
	Sender    string `json:sender`
	Receiver  string `json:receiver`
	Value     string `json:value`
	Timestamp string `json:timestamp`
}



func main(){
	router := mux.NewRouter();
	router.HandleFunc("/{masterWallet}", handler).Methods("GET")
	http.ListenAndServe(":9090", router)
}

func handler(w http.ResponseWriter, r *http.Request) {
	var data []obj
	vars := mux.Vars(r)

	masterWallet := vars["masterWallet"]

	getPage(masterWallet, 1, 0, 0, &data)

	firstStepLength := len(data)

	for i := 0; i < firstStepLength; i++ {
		getPage(data[i].Sender, 2, 0, 0, &data)
	}

	secondStepLength := len(data)

	for p := firstStepLength; p < secondStepLength; p++ {
		getPage(data[p].Sender, 3, 0, 0, &data)
	}

	for q := 0; q < len(data); q++ {
		getLastTransaction(data[q].Sender, data[q].Receiver, q, 0, data)
	}

	getTimestamps(data)

	returnData := data
	data = nil
	json.NewEncoder(w).Encode(&returnData)
	// data = nil
}

func getLastTransaction(addressFrom string, addressTo string, index int, page int, data []obj) {
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
		getLastTransaction(addressFrom, addressTo, index, page, data)
	}

}

func getPage(address string, degree int, page int, count int, data *[]obj) {
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
					return
				}

				if i == 1 {
					Sender = strings.TrimSpace(elem.Text())
				}

				if i == 2 {
					Receiver = strings.TrimSpace(elem.Text())
					*data = append(*data, obj{Sender: Sender, Receiver: Receiver})
					count++
				}
			})
		}
	})

	nextPageExists, _ := doc.Find("a.btn.btn-default.btn-xs.logout").Attr("href")
	fmt.Println(nextPageExists,"existss")
	if count < 5 && len(strings.TrimSpace(nextPageExists)) > 0 {
		page++
		if page == 1 {
			page++
		}
		getPage(address, degree, page, count, data)
	}

}

func getTimestamps(data []obj) {

	client, err := ethclient.Dial("https://mainnet.infura.io/QWMgExFuGzhpu2jUr6Pq")
	if err != nil {
		log.Fatal(err)
	}

	for r := 0; r < len(data); r++ {

		convertedBlock, _ := strconv.ParseInt(data[r].Timestamp, 10, 64)
		blockNumber := big.NewInt(convertedBlock)
		block, err := client.BlockByNumber(context.Background(), blockNumber)
		if err != nil {
			log.Fatal(err)
		}

		tm := time.Unix(block.Time().Int64(), 0)
		data[r].Timestamp = tm.String()
	}
}
