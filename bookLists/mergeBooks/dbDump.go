package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type Book struct {
	Title        string `json:"title"`
	Author       string `json:"author"`
	ISBN         string `json:"isbn"`
	Genre        string `json:"genre"`
	PublishedDate string `json:"published_date"`
}

func main() {
	// Database connection
	db, err := sql.Open("$IDKIFTHISISDEFAULT", "user=$USER dbname=$DBNAME sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Load books from JSON file
	file, err := os.ReadFile("data_standardized.json")
	if err != nil {
		log.Fatal(err)
	}

	var books []Book
	if err := json.Unmarshal(file, &books); err != nil {
		log.Fatal(err)
	}

	// Process each book
	for _, book := range books {
		var bookID int
		var err error
		if book.ISBN != "" {
		// Check by ISBN (normal case)
		err = db.QueryRow(`SELECT id FROM books WHERE isbn = $1`, book.ISBN).Scan(&bookID)
		} else {
		// Check by Title & Author when ISBN is NULL
		err = db.QueryRow(`SELECT id FROM books WHERE title = $1 AND author = $2`, book.Title, book.Author).Scan(&bookID)
		}
		
		if err == sql.ErrNoRows {
			// Book doesn't exist, insert it
			err = db.QueryRow(`
				INSERT INTO books (title, author, isbn, genre, published_date) 
				VALUES ($1, $2, $3, $4, $5) RETURNING id`, 
				book.Title, book.Author, book.ISBN, book.Genre, book.PublishedDate).Scan(&bookID)
			if err != nil {
				log.Printf("Failed to insert book: %v\n", err)
				continue
			}
		} else if err != nil {
			log.Fatal(err)
		}

		// Insert or get recommender
		var recommenderID int
		err = db.QueryRow(`SELECT id FROM recommenders WHERE name = $1`, "Some Guy").Scan(&recommenderID)

		if err == sql.ErrNoRows {
			err = db.QueryRow(`
				INSERT INTO recommenders (name) 
				VALUES ($1) RETURNING id`, "Some Guy").Scan(&recommenderID)
			if err != nil {
				log.Printf("Failed to insert recommender: %v\n", err)
				continue
			}
		} else if err != nil {
			log.Fatal(err)
		}

		// Insert recommendation (link recommender to book)
		_, err = db.Exec(`
			INSERT INTO recommendations (book_id, recommender_id) 
			VALUES ($1, $2) ON CONFLICT DO NOTHING`, bookID, recommenderID)
		if err != nil {
			log.Printf("Failed to insert recommendation: %v\n", err)
		}
	}

	fmt.Println("Books and recommendations added successfully!")
}

