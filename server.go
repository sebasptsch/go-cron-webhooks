package main

import (
	database_models "go-cron-webhooks/database-models"
	"go-cron-webhooks/graph"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/robfig/cron/v3"
	"github.com/vektah/gqlparser/v2/ast"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	database_url := os.Getenv("DATABASE_URL")
	if port == "" {
		port = defaultPort
	}

	if database_url == "" {
		log.Println("DATABASE_URL not set, using default SQLite database")
		database_url = "cron_webhooks.db"
	}

	db, err := gorm.Open(sqlite.Open(database_url), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&database_models.CronWebhookModel{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	c := cron.New()

	if err != nil {
		log.Fatalf("failed to create cron scheduler: %v", err)
	}

	webookInfo := make(map[uint]cron.EntryID)

	// get webook info from db
	var webhooks []database_models.CronWebhookModel
	if err := db.Find(&webhooks).Error; err != nil {
		panic("failed to fetch webhooks from database")
	}
	for _, webhook := range webhooks {
		if !webhook.Enabled {
			continue // skip disabled webhooks
		}
		// create new cron function
		ce, err := c.AddFunc(webhook.Cron, func() {
			database_models.RunWebhook(db, &webhook)
		})

		if err != nil {
			panic("failed to add cron job")
		}

		webookInfo[webhook.ID] = ce
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		DB:         db,
		C:          c,
		WebhookMap: webookInfo,
	}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)
	c.Start()

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
