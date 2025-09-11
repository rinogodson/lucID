package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("go-sqlite3", "./database/app.db")
	if err != nil {
		fmt.Print("Error on db: ", err)
	}
	createTable := `CREATE TABLE IF NOT EXISTS ppl (
		uid TEXT UNIQUE,
		name TEXT,
		team_no INTEGER,
		team_name TEXT,
		age INTEGER,
		id INTEGER PRIMARY KEY AUTOINCREMENT
	)`
	_, err = db.Exec(createTable)
	if err != nil {
		fmt.Print("ERROR: ", err)
	}
}

type Person struct {
	ID       int    `json:"id"`
	UID      string `json:"uid"`
	Name     string `json:"name"`
	TeamNo   int    `json:"team_no"`
	TeamName string `json:"team_name"`
	Age      int    `json:"age"` // just a number
}

func main() {
	initDB()
	r := gin.Default()

	r.POST("/new", func(ctx *gin.Context) {
		var bob Person
		err := ctx.ShouldBindJSON(&bob)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err = db.Exec("INSERT INTO ppl (uid, name, team_no, team_name, age) VALUES (?,?,?,?,?)",
			bob.UID,
			bob.Name,
			bob.TeamNo,
			bob.TeamName,
			bob.Age)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			fmt.Print("Error:", err)
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{"message": "this bud's one of our team now"})
	})

	r.GET("/allofem", func(ctx *gin.Context) {
		rows, err := db.Query("SELECT * FROM ppl")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close() // getting a warning here and ignoring it, why can't? Golang errors are a pain in the ...
		var people []Person
		for rows.Next() {
			var bob Person                                                       // bob goated here too!
			rows.Scan(&bob.UID, &bob.Name, &bob.TeamNo, &bob.TeamName, &bob.Age) // getting a warning here and ignoring it
			people = append(people, bob)
		}
		ctx.JSON(http.StatusOK, people)
	})

	r.GET("/uid/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		var bob Person // bob again :)
		err := db.QueryRow("SELECT * FROM ppl WHERE id = ?", id).Scan(&bob.UID, &bob.Name, &bob.TeamNo, &bob.TeamName, &bob.Age)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "this is a person from the quantum realm, not available here..."})
			return
		}
		ctx.JSON(http.StatusOK, bob)
	})

	r.PUT("/update/:uid", func(ctx *gin.Context) {
		id := ctx.Param("uid")
		var bob Person
		err := ctx.ShouldBindJSON(&bob)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err = db.Exec("UPDATE ppl SET name=?, team_no=?, team_name=?, age=? WHERE uid=?",
			bob.Name,
			bob.TeamNo,
			bob.TeamName,
			bob.Age,
			id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "it's updated!"})
	})

	r.DELETE("/kick/:uid", func(ctx *gin.Context) {
		id := ctx.Param("uid")
		_, err := db.Exec("DELETE FROM ppl WHERE id=?", id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "User deleted"})
	})

	r.Run(":8080")
}
