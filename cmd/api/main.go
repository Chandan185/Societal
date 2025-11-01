package main

import (
	"log"

	"github.com/Chandan185/Societal/internal/db"
	"github.com/Chandan185/Societal/internal/env"
	"github.com/Chandan185/Societal/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			Societal API
//	@description	API for Societal Social Network Application
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath					/v1
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//@description

func main() {
	cnf := config{
		addr: env.GetString("PORT", ":8000"),
		db: dbConfig{
			addr:        env.GetString("DB_ADDR", "postgres://user:adminpassword@localhost/socialnetwork?sslmode=disable"),
			maxOpenConn: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConn: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime: env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env:    env.GetString("ENV", "development"),
		apiURL: env.GetString("API_URL", "localhost:8000"),
	}

	//Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	//database
	db, err := db.New(cnf.db.addr, cnf.db.maxIdleTime, cnf.db.maxOpenConn, cnf.db.maxIdleConn)
	if err != nil {
		logger.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()
	logger.Info("Database connection pool established")
	store := store.NewStorage(db)
	app := &application{
		config: cnf,
		store:  store,
		logger: logger,
	}
	mux := app.mount()
	log.Fatal(app.run(mux))
}
