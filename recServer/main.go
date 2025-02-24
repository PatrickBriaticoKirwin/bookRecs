package main

import (
	"os"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/gocolly/colly"
	_ "github.com/jackc/pgx/v5/stdlib"
)


type ImportRequest struct {
	GoodreadsURL    string `json:"url"`
	RecommenderName string `json:"name"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://patrickbriaticokirwin.github.io", "http://localhost:8000"}, // Allow GitHub Pages frontend
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/books", func(c *gin.Context) {
		recommender := c.Query("recommender")

		query := `
		SELECT 
		b.title, 
		b.author, 
		b.genre, 
		STRING_AGG(r.name, ', ') AS recommenders
		FROM books b
		JOIN recommendations rec ON b.id = rec.book_id
		JOIN recommenders r ON rec.recommender_id = r.id
		`
		args := []interface{}{}

		if recommender != "" {
			query += " WHERE r.name = $1"
			args = append(args, recommender)
		}
		query += "GROUP BY b.id, b.title, b.author, b.genre;"

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{ "error": fmt.Sprintf("Failed to query books: %v", err)})
			return
		}
		defer rows.Close()

		var books []map[string]string
		for rows.Next() {
			var title, author, genre, recName string
			if err := rows.Scan(&title, &author, &genre, &recName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan book data"})
				return
			}
			books = append(books, map[string]string{
				"title":       title,
				"author":      author,
				"genre":       genre,
				"recommender": recName,
			})
		}
		c.JSON(http.StatusOK, books)
	})

	// Fetch unique recommendations by a single person
	r.GET("/books/unique", func(c *gin.Context) {
		recommender := c.Query("recommender")

		if recommender == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Recommender parameter is required"})
			return
		}

		query := `
		SELECT b.title, b.author, b.genre
		FROM books b
		JOIN recommendations rec ON b.id = rec.book_id
		JOIN recommenders r ON rec.recommender_id = r.id
		WHERE r.name = $1
		AND b.id NOT IN (
			SELECT rec.book_id
			FROM recommendations rec
			JOIN recommenders r2 ON rec.recommender_id = r2.id
			WHERE r2.name != $1
		)
		`

		rows, err := db.Query(query, recommender)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query unique books"})
			return
		}
		defer rows.Close()

		var books []map[string]string
		for rows.Next() {
			var title, author, genre string
			if err := rows.Scan(&title, &author, &genre); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan book data"})
				return
			}
			books = append(books, map[string]string{
				"title":  title,
				"author": author,
				"genre":  genre,
				"recommender": recommender,
			})
		}
		c.JSON(http.StatusOK, books)
	})
	r.POST("/import-request", func(c *gin.Context) {
		var req ImportRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		if !strings.Contains(req.GoodreadsURL, "goodreads.com") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request URL must be from goodreads"})
			return	
		}


		_, err := db.Exec(
			"INSERT INTO book_import_requests (recommender_name, goodreads_url) VALUES ($1, $2)",
			req.RecommenderName, req.GoodreadsURL,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store request"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Request submitted for approval"})
	})

	r.POST("/approve-import/{id}", func(c *gin.Context) {
		requestID := c.Param("id")

		// Get request details
		var req struct {
			RecommenderName string
			GoodreadsURL    string
		}
		err := db.QueryRow(
			"SELECT recommender_name, goodreads_url FROM book_import_requests WHERE id = $1 AND status = 'pending'",
			requestID,
		).Scan(&req.RecommenderName, &req.GoodreadsURL)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
			return
		}

		// Ensure recommender exists
		var recommenderID int
		err = db.QueryRow("SELECT id FROM recommenders WHERE name = $1", req.RecommenderName).Scan(&recommenderID)
		if err == sql.ErrNoRows {
			err = db.QueryRow("INSERT INTO recommenders (name) VALUES ($1) RETURNING id", req.RecommenderName).Scan(&recommenderID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recommender"})
				return
			}
		}

		// Scrape Goodreads & Insert books (placeholder function)
		books := ScrapeGoodreads(req.GoodreadsURL)
		for _, book := range books {
			var bookID int
			err = db.QueryRow("SELECT id FROM books WHERE title = $1 AND author = $2", book.Title, book.Author).Scan(&bookID)
			if err == sql.ErrNoRows {
				err = db.QueryRow("INSERT INTO books (title, author, genre) VALUES ($1, $2, $3) RETURNING id",
				book.Title, book.Author, book.Genre).Scan(&bookID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert book"})
					return
				}
			}

			_, err = db.Exec("INSERT INTO recommendations (book_id, recommender_id) VALUES ($1, $2)", bookID, recommenderID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert recommendation"})
				return
			}
		}

		// Update request status
		_, err = db.Exec("UPDATE book_import_requests SET status = 'approved' WHERE id = $1", requestID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update request status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Books imported successfully"})
	})


	r.Run(":8080") // Start server
}


type Book struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	ISBN          string `json:"isbn"`
	PublishedDate string `json:"published_date"`
	Genre         string `json:"genre"`
}

func ScrapeGoodreads(goodReadsUrl string) []Book {
	c := colly.NewCollector()
	var books []Book

	c.OnHTML("tr.bookalike.review", func(e *colly.HTMLElement) {
		title := e.ChildText("td.field.title a")
		author := e.ChildText("td.field.author a")
		isbn := e.ChildText("td.field.isbn .value")
		publishedDate := e.ChildText("td.field.date_pub div.value")


		books = append(books, Book{
			Title:    title,
			Author:   author,
			ISBN:     isbn,
			PublishedDate: publishedDate,
			Genre: "Unknown"})
	})

	err := c.Visit(goodReadsUrl)
	if err != nil {
		log.Fatal(err)
	}
	books = GetGenres(books)
	return books;
}

func GetGenres(books []Book) []Book {
	normalizeTitle := func(title string) string {
		return strings.TrimSpace(strings.ToLower(title))
	}

	isMatchingTitle := func(requested, retrieved string) bool {
		reqNorm := normalizeTitle(requested)
		retNorm := normalizeTitle(retrieved)
		return strings.Contains(reqNorm, retNorm) || strings.Contains(retNorm, reqNorm)
	}

	for i := range books {
		baseURL := "https://www.googleapis.com/books/v1/volumes"
		query := url.QueryEscape(books[i].Title + " " + books[i].Author)
		apiURL := fmt.Sprintf("%s?q=%s", baseURL, query)

		resp, err := http.Get(apiURL)
		if err != nil {
			fmt.Printf("Error fetching data for %s: %v\n", books[i].Title, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("API returned status %d for %s\n", resp.StatusCode, books[i].Title)
			continue
		}

		var result struct {
			Items []struct {
				VolumeInfo struct {
					Title      string   `json:"title"`
					Authors    []string `json:"authors"`
					Categories []string `json:"categories"`
				} `json:"volumeInfo"`
			} `json:"items"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Error decoding JSON for %s: %v\n", books[i].Title, err)
			continue
		}

		if len(result.Items) == 0 {
			fmt.Printf("No matching books found for %s\n", books[i].Title)
			continue
		}

		volume := result.Items[0].VolumeInfo

		if !isMatchingTitle(books[i].Title, volume.Title) {
			fmt.Printf("Mismatch: requested '%s', got '%s'\n", books[i].Title, volume.Title)
			continue
		}

		if len(volume.Categories) > 0 {
			books[i].Genre = volume.Categories[0]
		} else {
			books[i].Genre = "Unknown"
		}

		fmt.Printf("Updated: %s -> Genre: %s\n", books[i].Title, books[i].Genre)

		time.Sleep(1 * time.Second) // Sleep to avoid rate limiting
	}

	return books
}

