package main

import (
	"log"
	"net/http"

	"github.com/YagoSchramm/GoDepot/service"
)

func main() {
	r, cleanup, err := service.Build()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	log.Fatal(http.ListenAndServe(":8080", r))
}
