// Command admin is the admin API — content management, sync, and the Telegram
// bridge proxy. Runs locally only; never deployed publicly. It also serves the
// public routes so the admin UI and the bridge scripts need only one base URL.
package main

import (
	"log"

	"video-scraper/config"
	"video-scraper/db"
	"video-scraper/handlers"
	"video-scraper/router"
)

func main() {
	config.Load()
	db.Init()
	handlers.StartBackground() // 6-hourly auto-sync lives with the admin service

	log.Printf("[admin] admin API listening on :%s", config.C.AdminPort)
	if err := router.SetupAdmin().Run(":" + config.C.AdminPort); err != nil {
		log.Fatal(err)
	}
}
