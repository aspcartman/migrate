package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/go-pg/pg"

	"github.com/aspcartman/migrate"
)

func main() {
	var args = struct {
		Folder   string `arg:"positional,required"`
		Addr     string
		User     string
		Password string
		Database string
		Devmode  bool
	}{
		Addr:     "localhost:5432",
		User:     "postgres",
		Password: "postgres",
		Database: "postgres",
	}
	arg.MustParse(&args)

	db := pg.Connect(&pg.Options{
		Addr:     args.Addr,
		User:     args.User,
		Password: args.Password,
		Database: args.Database,
	})

	if err := migrate.MigrateFromFolder(db, args.Folder, args.Devmode); err != nil {
		fmt.Printf("migrations failed: %s\n", err)
		os.Exit(-1)
	}

	fmt.Println("Migrations done")
}
