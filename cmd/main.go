package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/OdysseyMomentumExperience/web3-identity-provider/pkg/config"
	"github.com/OdysseyMomentumExperience/web3-identity-provider/pkg/handler"
	"github.com/OdysseyMomentumExperience/web3-identity-provider/pkg/hydra"
	"gopkg.in/yaml.v2"
)

func main() {
	log.Println("Starting guest identity provider")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
	prettyPrint(cfg)

	hydraClient := hydra.NewHydraClient(cfg.AdminURL)
	oidcHandler := handler.NewHandler(hydraClient)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("Listening on %s", addr)
	http.ListenAndServe(addr, oidcHandler)
}

func prettyPrint(cfg *config.Config) {
	d, _ := yaml.Marshal(cfg)
	log.Printf("--- Config ---\n%s\n", d)
}
