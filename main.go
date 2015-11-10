package main

import (
	"log"
	"os"

	"github.com/RichardKnop/go-oauth2-server/config"
	"github.com/RichardKnop/go-oauth2-server/database"
	"github.com/RichardKnop/go-oauth2-server/migrations"
	"github.com/RichardKnop/go-oauth2-server/oauth"
	"github.com/codegangsta/cli"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

var app *cli.App

func init() {
	// Initialise a CLI app
	app = cli.NewApp()
	app.Name = "go-oauth2-server"
	app.Usage = "OAuth 2.0 Server"
	app.Author = "Richard Knop"
	app.Email = "risoknop@gmail.com"
	app.Version = "0.0.0"
}

func main() {
	// Load the configuration, connect to the database
	cnf := config.NewConfig()
	db, err := database.NewDatabase(cnf)
	if err != nil {
		log.Fatal(err)
	}

	// Set the CLI app commands
	app.Commands = []cli.Command{
		{
			Name:   "migrate",
			Usage:  "run migrations",
			Action: func(c *cli.Context) { migrate(db) },
		},
		{
			Name:  "runserver",
			Usage: "run web server",
			Action: func(c *cli.Context) {
				oauth.InitService(cnf, db)
				runServer()
			},
		},
	}

	app.Run(os.Args)
}

func migrate(db *gorm.DB) {
	// Bootsrrap migrations
	if err := migrations.Bootstrap(db); err != nil {
		log.Fatal(err)
	}
	// Run migrations for the oauth service
	if err := oauth.MigrateAll(db); err != nil {
		log.Fatal(err)
	}
}

func runServer() {
	// Start a negroni app
	n := negroni.Classic()

	// Create a router instance
	router := mux.NewRouter().StrictSlash(true)
	// Add routes for the oauth service
	for _, route := range oauth.Routes {
		router.PathPrefix("/oauth").Subrouter().
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	// Set the router
	n.UseHandler(router)

	// Run the server on port 8080
	n.Run(":8080")
}
