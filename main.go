package ETH

import(
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"strconv"
)

type Book struct{
	ID int `json:id`
	Title string `json:title`
	Author string `json:author`
	Year string `json:year`
}

var books []Book

func yell(){


	books = append(books, Book{ ID:1, Title:"Intro to programming!", Author:"sheraz arshad", Year:"2019" },
	Book{ ID:2, Title:"2Intro to programming!", Author:"sheraz arshad", Year:"2019" },
	Book{ ID:3, Title:"3Intro to programming!", Author:"sheraz arshad", Year:"2019" },
	Book{ ID:4, Title:"4Intro to programming!", Author:"sheraz arshad", Year:"2019" })

	router := mux.NewRouter()
	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/books", addBook).Methods("POST")
	router.HandleFunc("/books", updateBook).Methods("PUT")
	router.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")
	
	log.Fatal(http.ListenAndServe(":8000", router))
}

func getBooks(w http.ResponseWriter, r *http.Request){
	json.NewEncoder(w).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	
	i, _ := strconv.Atoi(params["id"])

	for _, book := range books{
		if book.ID == i{
			json.NewEncoder(w).Encode(book)
		}
	}
}

func addBook(w http.ResponseWriter, r *http.Request){
	var book Book
	json.NewDecoder(r.Body).Decode(&book)

	books := append(books, book)
	json.NewEncoder(w).Encode(books)
}

func updateBook(w http.ResponseWriter, r *http.Request){
	var book Book
	json.NewDecoder(r.Body).Decode(&book)

	for i, item := range books{
		if item.ID == book.ID{
			books[i] = book
			json.NewEncoder(w).Encode(&book)
		}
	}
}

func deleteBook(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)

	id, _ := strconv.Atoi(params["id"])

	for i, item := range books{
		if item.ID == id{
			books = append(books[:i], books[i+1:]...)
			json.NewEncoder(w).Encode(id)
		}
	}
}