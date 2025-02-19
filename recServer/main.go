package main

import (
	"os"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	_ "github.com/jackc/pgx/v5/stdlib"
)


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
		AllowOrigins:     []string{"https://patrickbriaticokirwin.github.io"}, // Allow GitHub Pages frontend
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
		GROUP BY b.id, b.title, b.author, b.genre;
		`
		args := []interface{}{}

		if recommender != "" {
			query += " WHERE r.name = $1"
			args = append(args, recommender)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query books"})
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
			})
		}
		c.JSON(http.StatusOK, books)
	})

	r.Run(":8080") // Start server
}

