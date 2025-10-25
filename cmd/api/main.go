package main

import (
	"log"

	"github.com/Chandan185/Societal/internal/db"
	"github.com/Chandan185/Societal/internal/env"
	"github.com/Chandan185/Societal/internal/store"
)

const version = "1.0.0"

func main() {
	cnf := config{
		addr: env.GetString("PORT", ":8000"),
		db: dbConfig{
			addr:        env.GetString("DB_ADDR", "postgres://user:adminpassword@localhost/socialnetwork?sslmode=disable"),
			maxOpenConn: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConn: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime: env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
	}
	db, err := db.New(cnf.db.addr, cnf.db.maxIdleTime, cnf.db.maxOpenConn, cnf.db.maxIdleConn)
	if err != nil {
		log.Panic("Error connecting to the database:", err)
	}
	defer db.Close()
	log.Println("Database connection pool established")
	store := store.NewStorage(db)
	app := &application{
		config: cnf,
		store:  store,
	}
	mux := app.mount()
	log.Fatal(app.run(mux))
}
