package database

import (
	"context"
	"log"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

func DBinstance() driver.Database {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		log.Fatal("Failed to create HTTP connection:", err)
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})
	if err != nil {
		log.Fatal("Failed to create database connection:", err)
	}

	var db driver.Database

	db_exists, err := client.DatabaseExists(context.TODO(), "restaurant")
	if err != nil {
		log.Fatal("Failed to check if database exists:", err)
	}

	if db_exists {
		log.Println("That db exists already")
		db, err = client.Database(context.TODO(), "restaurant")
		if err != nil {
			log.Fatal("Failed to create database:", err)
		}
	} else {
		db, err = client.CreateDatabase(context.TODO(), "restaurant", nil)
		if err != nil {
			log.Fatal("Failed to create database:", err)
		}
	}

	return db
}

func OpenCollection(db driver.Database, collectionName string) driver.Collection {
	var col driver.Collection

	col_exists, err := db.CollectionExists(context.TODO(), collectionName)
	if err != nil {
		log.Fatal("Failed to check if database exists:", err)
	}

	if col_exists {
		log.Println("That collection exists already")
		col, err = db.Collection(context.TODO(), collectionName)
		if err != nil {
			log.Fatal("Failed to create database:", err)
		}
	} else {
		col, err = db.CreateCollection(context.TODO(), collectionName, nil)
		if err != nil {
			log.Fatal("Failed to create collection:", err)
		}
	}

	return col
}
