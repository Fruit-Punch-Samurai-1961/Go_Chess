package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sheshan1961/chessapp/cmd/websockets"
	"github.com/sheshan1961/chessapp/pkg/models/mysql"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

//Struct for all of our config settings
type Config struct {
	Addr           string
	StaticDir      string
	DataSourceName string
}

//define contextkey type to pass into request context
type contextKey string
var contextKeyUser = contextKey("user")


//make an application struct which will essentially be a method which will allow us to use dependency variables across the package
//1. errorLog is a log.Logger which will use the errLog personal logger defined in main
//2. infoLog is a log.Logger which will use the infoLog personal logger defined in main
//3. games will contain the Game Model object
//4. hub will store the websocket connection which maintains a set of clients, rooms and sends messages to the clients
//5. templateCache maintains a dict of string (key) and templates (value)
//6. sizedBufferPool a bufferpool to make it so that when we check for errors, it's more efficient rather than creating a new buffer each time
type application struct {
	errorLog        *log.Logger
	infoLog         *log.Logger
	games           *mysql.GameModel
	hub             *websockets.Hub
	templateCache   map[string]*template.Template
	sizedBufferPool *SizedBufferPool
	sessionManager  *scs.SessionManager
	users           *mysql.UserModel
}

//function to connect our database with the driver so we can access it
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	//Command Line flags to store our HTTP address and static file dir
	cfg := new(Config)
	flag.StringVar(&cfg.Addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.StaticDir, "static-dir", "./ui/static", "Path to Static Dir")
	flag.StringVar(&cfg.DataSourceName, "dsn", "username:password/chessapp?parseTime=true", "MySQL Database for chess games")
	flag.Parse()

	//connect to our chess app database
	db, err := openDB(cfg.DataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//personal loggers: 1 for info and another for error
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	//Initalize empty template cache...
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errLog.Fatal(err)
	}

	//declare new session manager and set lifetime to expire after 12 hours
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour

	//make hub for all connections and pass in the database, so we can save the last move made once the player exits
	hub := websockets.NewHub(db)
	go hub.Run()
	//make instance of application
	app := &application{
		errorLog:        errLog,
		infoLog:         infoLog,
		games:           &mysql.GameModel{DB: db},
		hub:             hub,
		templateCache:   templateCache,
		sizedBufferPool: NewSizedBufferPool(48, 7000),
		sessionManager:  sessionManager,
		users:           &mysql.UserModel{DB: db},
	}

	//create a struct to hold non-default TLS settings
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	//make a http.server struct to set the error logs to the new error log we made
	srv := &http.Server{
		Addr:         cfg.Addr,
		ErrorLog:     errLog,
		Handler:      app.routes(cfg),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", cfg.Addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errLog.Fatal(err)

}
