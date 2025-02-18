package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

type Book struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	ISBN          string `json:"isbn"`
	PublishedDate string `json:"published_date"`
	Genre         string `json:"genre"`
}

func readExistingBooks(filename string) ([]Book, error) {
	var books []Book
	data, err := ioutil.ReadFile(filename)
	if err == nil {
		err = json.Unmarshal(data, &books)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	return books, nil
}

func writeBooksToFile(filename string, books []Book) error {
	updatedData, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}
	return ioutil.WriteFile(filename, updatedData, 0644)
}

func main() {
	filename := "books.json"
	books, err := readExistingBooks(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	c := colly.NewCollector()

	c.OnHTML("tr.bookalike.review", func(e *colly.HTMLElement) {
		// Extract necessary fields
		title := e.ChildText("td.field.title a")
		author := e.ChildText("td.field.author a")
		isbn := e.ChildText("td.field.isbn div.value")
		publishedDate := e.ChildText("td.field.date_pub div.value")
		genre := "Unknown" // Genre might need a second request

		// Check for "it was amazing" rating
		if !strings.Contains(e.ChildText("td.field.rating"), "it was amazing") {
			return // Skip books without this rating
		}

		newBook := Book{Title: title, Author: author, ISBN: isbn, PublishedDate: publishedDate, Genre: genre}
		books = append(books, newBook)
	})

	c.Visit("website")

	if err := writeBooksToFile(filename, books); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Books appended successfully!")
	}
}

