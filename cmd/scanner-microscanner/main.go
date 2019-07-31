package main

import (
	"github.com/danielpacak/harbor-scanner-microscanner/pkg/etc"
	"github.com/danielpacak/harbor-scanner-microscanner/pkg/http/api/v1"
	"github.com/danielpacak/harbor-scanner-microscanner/pkg/model"
	"github.com/danielpacak/harbor-scanner-microscanner/pkg/scanner/microscanner"
	"github.com/danielpacak/harbor-scanner-microscanner/pkg/store"
	"github.com/danielpacak/harbor-scanner-microscanner/pkg/store/fs"
	"github.com/danielpacak/harbor-scanner-microscanner/pkg/store/redis"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	cfg, err := etc.GetConfig()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Printf("Starting harbor-scanner-microscanner with config %v", cfg)

	dataStore, err := GetStore(cfg)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	scanner, err := microscanner.NewScanner(cfg, model.NewTransformer(), dataStore)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	apiHandler := v1.NewAPIHandler(scanner)

	router := mux.NewRouter()
	v1Router := router.PathPrefix("/api/v1").Subrouter()

	v1Router.Methods("GET").Path("").HandlerFunc(apiHandler.GetVersion)
	v1Router.Methods("POST").Path("/scan").HandlerFunc(apiHandler.CreateScan)
	v1Router.Methods("GET").Path("/scan/{detailsKey}").HandlerFunc(apiHandler.GetScanResult)

	err = http.ListenAndServe(cfg.Addr, router)
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error: %v", err)
	}
}

func GetStore(cfg *etc.Config) (store.DataStore, error) {
	switch cfg.StoreDriver {
	case etc.StoreDriverFS:
		return fs.NewStore(cfg.FSStore.DataDir)
	case etc.StoreDriverRedis:
		return redis.NewStore(cfg.RedisStore.RedisURL)
	default:
		return nil, nil
	}
}
