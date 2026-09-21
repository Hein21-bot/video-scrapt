package router

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"video-scraper/config"
	"video-scraper/handlers"
	"video-scraper/middleware"
)

// SetupPublic builds the deployed, read-only public API.
func SetupPublic() *gin.Engine {
	r := gin.Default()
	r.Use(corsMW())
	registerPublic(r)
	return r
}

// SetupAdmin builds the local admin API. It also mounts the public routes so an
// operator only needs one base URL, and so the admin UI can preview content.
func SetupAdmin() *gin.Engine {
	r := gin.Default()
	r.Use(corsMW())
	registerPublic(r)
	registerAdmin(r)
	return r
}

func registerPublic(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	api := r.Group("/api")
	api.Use(middleware.APITokenAuth())
	{
		api.GET("/videos", handlers.HandleVideoList)
		api.GET("/search", handlers.HandleSearch)
		api.GET("/actors", handlers.HandleActors)
		api.GET("/categories", handlers.HandleCategoriesList)
		api.GET("/channels", handlers.HandleChannelsList)
		api.GET("/related", handlers.HandleRelated)
		api.GET("/video-url", handlers.HandleVideoURL)
		api.POST("/share", handlers.HandleShareCreate)
		api.GET("/s/:code", handlers.HandleShareGet)
	}

	// Media proxy — no token, browsers load these directly via img/video src.
	r.GET("/api/stream", handlers.HandleStream)
	r.HEAD("/api/stream", handlers.HandleStreamHEAD)
	r.GET("/api/thumb/:id", handlers.HandleThumb)
}

func registerAdmin(r *gin.Engine) {
	r.POST("/api/admin/login", handlers.HandleAdminLogin)

	p := r.Group("/api/admin")
	p.Use(middleware.AdminAuth())
	{
		p.GET("/stats", handlers.HandleAdminStats)
		p.GET("/sync/status", handlers.HandleSyncStatus)
		p.POST("/sync", handlers.HandleSync)
		p.POST("/sync/target", handlers.HandleSyncTarget)
		p.GET("/videos", handlers.HandleAdminVideos)
		p.POST("/videos", handlers.HandleAdminCreateVideo)
		p.DELETE("/videos", handlers.HandleAdminDeleteVideo)
		p.GET("/bridge/status", handlers.HandleBridgeStatus)
		p.POST("/bridge/add", handlers.HandleBridgeAdd)
		p.GET("/bridge/jobs", handlers.HandleBridgeJobs)
		p.GET("/shares", handlers.HandleAdminShares)
		p.GET("/categories", handlers.HandleAdminCategoriesList)
		p.POST("/categories", handlers.HandleAdminCategoryCreate)
		p.PUT("/categories/:id", handlers.HandleAdminCategoryUpdate)
		p.DELETE("/categories/:id", handlers.HandleAdminCategoryDelete)
		p.GET("/channels", handlers.HandleAdminChannelsList)
		p.POST("/channels", handlers.HandleAdminChannelCreate)
		p.PUT("/channels/:id", handlers.HandleAdminChannelUpdate)
		p.DELETE("/channels/:id", handlers.HandleAdminChannelDelete)
	}
}

// corsMW allows the origins in CORS_ORIGINS (comma-separated); with none set it
// allows any localhost origin, which covers local dev for both apps.
func corsMW() gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Range", "Authorization", "X-Api-Token"},
		ExposeHeaders: []string{"Content-Length", "Content-Range", "Accept-Ranges"},
	}
	if list := strings.TrimSpace(config.C.CORSOrigins); list != "" {
		for _, o := range strings.Split(list, ",") {
			if o = strings.TrimSpace(o); o != "" {
				cfg.AllowOrigins = append(cfg.AllowOrigins, o)
			}
		}
	} else {
		cfg.AllowOriginFunc = func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:")
		}
	}
	return cors.New(cfg)
}
