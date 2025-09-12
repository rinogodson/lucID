package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"lucid.backend/uidgen"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./database/app.db")
	if err != nil {
		fmt.Print("Error on db: ", err)
	}
	createTable := `CREATE TABLE IF NOT EXISTS ppl (
		uid TEXT UNIQUE,
		name TEXT,
		age INTEGER,
		id INTEGER PRIMARY KEY AUTOINCREMENT,
  	team_id INTEGER,
  	FOREIGN KEY(team_id) REFERENCES teams(id) ON DELETE SET NULL
	)`

	createTeamTable := `CREATE TABLE IF NOT EXISTS teams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT
	)`
	_, err = db.Exec(createTable)
	if err != nil {
		fmt.Print("ERROR: ", err)
	}
	_, err = db.Exec(createTeamTable)
	if err != nil {
		fmt.Print("ERROR: ", err)
	}

	// adding the default team for those singles
	createDefaultTeam := `INSERT OR IGNORE INTO teams (id, name) VALUES (1, 'Single')`
	_, err = db.Exec(createDefaultTeam)
	if err != nil {
		fmt.Print("ERROR: ", err)
	}
}

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Person struct {
	ID     int    `json:"id"`
	UID    string `json:"uid"`
	Name   string `json:"name"`
	Age    int    `json:"age"` // just a number
	TeamID int    `json:"team_id"`
	Team   *Team  `json:"team,omitempty"`
}

func getUID() string {
	for {
		uid := uidgen.UIDGen()
		var exists string
		err := db.QueryRow("SELECT uid FROM ppl WHERE uid = ?", uid).Scan(&exists)
		if err == sql.ErrNoRows {
			return uid
		}
	}
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

		bob.UID = getUID()

		_, err = db.Exec("INSERT INTO ppl (uid, name, age, team_id) VALUES (?,?,?,?)",
			bob.UID,
			bob.Name,
			bob.Age,
			bob.TeamID,
		)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			fmt.Print("Error:", err)
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{"message": "this bud's one of our kind now"})
	})

	r.GET("/allofem", func(ctx *gin.Context) {
		rows, err := db.Query("SELECT uid, name, age, id, team_id FROM ppl")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close() // getting a warning here and ignoring it, why can't? Golang errors are a pain in the ...
		var people []Person
		for rows.Next() {
			var bob Person                                                        // bob goated here too!
			err := rows.Scan(&bob.UID, &bob.Name, &bob.Age, &bob.ID, &bob.TeamID) // getting a warning here and ignoring it
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			people = append(people, bob)
		}
		ctx.JSON(http.StatusOK, people)
	})

	r.GET("/person/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		var bob Person // bob again :)
		err := db.QueryRow("SELECT uid, name, age, id, team_id FROM ppl WHERE uid = ?", id).Scan(&bob.UID, &bob.Name, &bob.Age, &bob.ID, &bob.TeamID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "this is a person from the quantum realm, not available here..."})
			return
		}
		ctx.JSON(http.StatusOK, bob)
	})

	r.PUT("/update/:uid", func(ctx *gin.Context) {
		id := ctx.Param("uid")
		var bob Person
		if err := ctx.ShouldBindJSON(&bob); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var alice Person
		err := db.QueryRow("SELECT uid, name, age, id, team_id FROM ppl WHERE uid = ?", id).
			Scan(&alice.UID, &alice.Name, &alice.Age, &alice.ID, &alice.TeamID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updates := []string{}
		args := []any{}

		if bob.Name != "" && bob.Name != alice.Name {
			updates = append(updates, "name=?")
			args = append(args, bob.Name)
		}
		if bob.Age != 0 && bob.Age != alice.Age {
			updates = append(updates, "age=?")
			args = append(args, bob.Age)
		}
		if bob.TeamID != 0 && bob.TeamID != alice.TeamID {
			updates = append(updates, "team_id=?")
			args = append(args, bob.TeamID)
		}

		if len(updates) == 0 {
			ctx.JSON(http.StatusOK, gin.H{"message": "you want to change smthg? what it is??"})
			return
		}

		query := fmt.Sprintf("UPDATE ppl SET %s WHERE uid=?", strings.Join(updates, ", "))
		args = append(args, id)

		_, err = db.Exec(query, args...)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "updated!"})
	})

	r.DELETE("/kick/:uid", func(ctx *gin.Context) {
		id := ctx.Param("uid")
		_, err := db.Exec("DELETE FROM ppl WHERE uid=?", id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "this guy's deleted"})
	})

	// Teams Stuff

	r.GET("/teams/all", func(ctx *gin.Context) {
		rows, err := db.Query(`
    SELECT ppl.id, ppl.uid, ppl.name, ppl.age, ppl.team_id, teams.id, teams.name
    FROM ppl
    LEFT JOIN teams ON ppl.team_id = teams.id
		`)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer rows.Close()

		var people []Person
		for rows.Next() {
			var bob Person
			var team Team
			err := rows.Scan(&bob.ID, &bob.UID, &bob.Name, &bob.Age, &bob.TeamID, &team.ID, &team.Name)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if team.ID != 0 {
				bob.Team = &team
			}
			people = append(people, bob)
		}
		ctx.JSON(http.StatusOK, people)

	})

	r.GET("/teams/:id", func(ctx *gin.Context) {
		teamIDfromCtx := ctx.Param("id")
		rows, err := db.Query(`
		SELECT ppl.id, ppl.uid, ppl.name, ppl.age, ppl.team_id, teams.id, teams.name
		FROM ppl
		LEFT JOIN teams ON ppl.team_id = teams.id
		WHERE ppl.team_id=?`, teamIDfromCtx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		var nameOftheTeam string
		err = db.QueryRow(`SELECT name FROM teams WHERE id=?`, teamIDfromCtx).Scan(&nameOftheTeam)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "no team like that"})
			return
		}
		// id := ctx.Param("id")
		// var bob Person // bob again :)
		// err := db.QueryRow("SELECT * FROM ppl WHERE id = ?", id).Scan(&bob.UID, &bob.Name, &bob.Age, &bob.ID, &bob.TeamID)
		// if err != nil {
		// 	ctx.JSON(http.StatusNotFound, gin.H{"error": "this is a person from the quantum realm, not available here..."})
		// 	return
		// }
		// ctx.JSON(http.StatusOK, bob)
		//
		// if err != nil {
		// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }

		defer rows.Close()

		var people []Person
		for rows.Next() {
			var bob Person
			var team Team
			err := rows.Scan(&bob.ID, &bob.UID, &bob.Name, &bob.Age, &bob.TeamID, &team.ID, &team.Name)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if team.ID != 0 {
				bob.Team = &team
			}
			people = append(people, bob)
		}
		ctx.JSON(http.StatusOK, people)
	})

	r.POST("/new/team", func(ctx *gin.Context) {
		var newTeam Person
		err := ctx.ShouldBindJSON(&newTeam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := db.Exec("INSERT INTO teams (name) VALUES (?)",
			newTeam.Name,
		)
		lastID, err := res.LastInsertId()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			fmt.Print("Error:", err)
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{"message": "yay! new team", "id": lastID})
	})
	r.Run(":8080")
}
