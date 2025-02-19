package main

import (
    "os"
    "database/sql"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
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

    r.GET("/books", func(c *gin.Context) {
        rows, err := db.Query("SELECT title, author, genre FROM books")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query books"})
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

