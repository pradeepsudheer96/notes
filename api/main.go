package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rs/cors"
)

const (
	databaseMaxOpenConns = 20
	databaseMaxIdleConns = 20
)
const (
	dbHost     = "localhost"
	dbPort     = "5432"
	dbUser     = "pradeepchelamala"
	dbPassword = "123456"
	dbName     = "notes"
)

type NoteInput struct {
	ID         int    `json:"-"`
	UserID     string `json:"userid"`
	Title      string `json:"title"`
	BuyTarget  string `json:"buyTarget"`
	SellTarget string `json:"sellTarget"`
	Notes      string `json:"notes"`
	Output     string `json:"output"`
}

type NoteOutput struct {
	ID         int    `json:"id"`
	UserID     string `json:"-"`
	Title      string `json:"title"`
	BuyTarget  string `json:"buyTarget"`
	SellTarget string `json:"sellTarget"`
	Notes      string `json:"notes"`
	Output     string `json:"output"`
}

type application struct {
	client    *http.Client
	db        *sql.DB
	logger    *log.Logger
	errLogger *log.Logger
}

func (app *application) Create(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/create" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	fmt.Println("path is correct")
	if r.Method != http.MethodPost {
		app.errLogger.Print("invalid method upon creation")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("method is correct")
	fmt.Println(r)
	fmt.Println(r.Body)
	var note NoteInput
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		app.errLogger.Print(err)
		return
	}
	fmt.Println("decoded")
	fmt.Println(note)
	stmt := `
		INSERT INTO notes(userid, title, buy_target, sell_target, notes, output)
		VALUES($1, $2, $3, $4, $5, $6);`
	fmt.Println("stmt")
	fmt.Println(note)
	_, err := app.db.Exec(stmt, note.UserID, note.Title, note.BuyTarget, note.SellTarget, note.Notes, note.Output)
	if err != nil {
		app.errLogger.Print(err)
		return
	}
	fmt.Println("executed")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (app *application) GetNotes(w http.ResponseWriter, r *http.Request) {
	// if r.URL.Path != "/notes" {
	// 	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	// 	return
	// }
	fmt.Println("path is correct")
	params := mux.Vars(r)

	if r.Method != http.MethodGet {
		app.errLogger.Print("invalid method upon creation")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("method is correct")
	fmt.Println(params)
	var userid = params["userid"]

	fmt.Println(userid, ":decoded")
	stmt := `
		SELECT id, userid, title, buy_target, sell_target, notes, output
		FROM notes
		WHERE userid = $1;`
	rows, err := app.db.Query(stmt, userid)
	if err != nil {
		app.errLogger.Print(err)
		return
	}
	defer rows.Close()
	fmt.Println("executed")
	notes := []NoteOutput{}
	for rows.Next() {
		var note NoteOutput
		if err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.BuyTarget, &note.SellTarget, &note.Notes, &note.Output); err != nil {
			app.errLogger.Print(err)
			return
		}
		notes = append(notes, note)
	}
	b, err := json.Marshal(notes)
	fmt.Println(b)
	if err != nil {
		return
	}
	w.Write(b)
}

func newApplication() (*application, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	fmt.Println(psqlInfo)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println(err)
		return nil, errors.Wrap(err, "error connecting to database")
	}
	fmt.Println("here2")
	// By default, the max idle connections is 0, meaning each connection
	// will close immediately. Setting the max idle keeps connections open.
	db.SetMaxOpenConns(databaseMaxOpenConns)
	db.SetMaxIdleConns(databaseMaxIdleConns)

	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to database")
	}
	return &application{
		&http.Client{Timeout: time.Second * 60},
		db,
		log.New(os.Stdout, "[Notes] ", log.LstdFlags),
		log.New(os.Stderr, "[Notes] ", log.LstdFlags),
	}, nil
}

func main() {
	fmt.Println("here")
	app, err := newApplication()
	if err != nil {
		return
	}
	router := mux.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Origin", "Content-Type", "X-Auth-Token", "Authorization"},
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})
	defer app.db.Close()

	// ctx = context.Background()
	router.HandleFunc("/create", app.Create).Methods("POST", "OPTIONS")
	router.HandleFunc("/notes/{userid}", app.GetNotes).Methods("GET", "OPTIONS")
	handler := c.Handler(router)
	fmt.Println("Server started")
	log.Fatal(http.ListenAndServe(":6060", handler))

}
