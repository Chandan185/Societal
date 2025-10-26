package main

import (
	"log"

	"github.com/Chandan185/Societal/internal/db"
	"github.com/Chandan185/Societal/internal/env"
	"github.com/Chandan185/Societal/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://user:adminpassword@localhost/socialnetwork?sslmode=disable")
	conn, err := db.New(addr, "15m", 3, 3)
	if err != nil {
		log.Fatal("Cannot connect to db:", err)
	}
	store := store.NewStorage(conn)
	db.Seed(store)
}
