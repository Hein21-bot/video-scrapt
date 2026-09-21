// Command api is the public, read-only API. This is the deployed service —
// it has no admin routes and does not run the background scraper.
package main

import (
	"log"

	"video-scraper/config"
	"video-scraper/db"
	"video-scraper/router"
)

func main() {
	config.Load()
	db.Init()

	log.Printf("[api] public API listening on :%s", config.C.Port)
	if err := router.SetupPublic().Run(":" + config.C.Port); err != nil {
		log.Fatal(err)
	}
}
