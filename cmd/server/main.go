package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"postsys/internal/db"
	"postsys/internal/db/memdb"
	"postsys/internal/db/pgsql"
	"postsys/internal/graph"
	"postsys/internal/service"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
	timeout := mustGetDuration("TIMEOUT")
	idleTimeout := mustGetDuration("IDLE_TIMEOUT")
	shutdownTimeout := mustGetDuration("SHUTDOWN_TIMEOUT") // новая переменная

	port := mustGetEnv("SERVER_PORT")
	storageType := os.Getenv("STORAGE_TYPE")

	var dbInstance db.DB
	var pool *pgxpool.Pool // для закрытия при необходимости

	switch storageType {
	case "postgres":
		connStr := buildDatabaseURL()
		var err error
		pool, err = pgxpool.New(context.Background(), connStr)
		if err != nil {
			log.Fatal("Unable to connect to database:", err)
		}
		dbInstance = pgsql.New(pool)
	default:
		dbInstance = memdb.New()
	}

	svc := service.New(dbInstance)
	resolver := graph.NewResolver(svc)

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	// настройка транспортов и кэшей (без изменений)
	srv.AddTransport(transport.Websocket{})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", graph.WithChildrenMap(srv))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      nil, // используем стандартный мультиплексор
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  idleTimeout,
	}

	go func() {
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	if pool != nil {
		pool.Close()
		log.Println("database connection pool closed")
	}
}

func buildDatabaseURL() string {
	protocol := mustGetEnv("DB_PROTOCOL")
	user := mustGetEnv("DB_USER")
	password := mustGetEnv("DB_PASSWORD")
	host := mustGetEnv("DB_HOST")
	port := mustGetEnv("DB_PORT")
	dbname := mustGetEnv("DB_NAME")
	options := os.Getenv("DB_OPTIONS")

	url := fmt.Sprintf("%s://%s:%s@%s:%s/%s", protocol, user, password, host, port, dbname)
	if options != "" {
		url += "?" + options
	}
	return url
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return val
}

func mustGetDuration(key string) time.Duration {
	str := mustGetEnv(key)
	dur, err := time.ParseDuration(str)
	if err != nil {
		log.Fatalf("invalid %s value: %v", key, err)
	}
	return dur
}
