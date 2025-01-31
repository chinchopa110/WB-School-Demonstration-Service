package HTTP

import (
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/HTTP"
	"log"
	"net/http"
)

func ServerGetUp(service OrdersServices.IGetService) {
	server := HTTP.NewServer(service)

	log.Println("HTTP server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", server))
}
