package httpapi

import (
	"encoding/json"
	"github.com/janicaleksander/StocksHelp/stockapi"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

type Server struct {
	listenAddress string
	Hub           *stockapi.Hub
}

func NewServer(add string, h *stockapi.Hub) *Server {
	return &Server{listenAddress: add, Hub: h}
}
func WriteJson(w http.ResponseWriter, code int, val any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(val)

}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func makeHandler(f apiFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := f(writer, request); err != nil {
			_ = json.NewEncoder(writer).Encode(err)
		}
	}
}

func (serv *Server) Run() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	router := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))
	router.HandleFunc("/home", makeHandler(serv.home))

	log.Printf("Listening to %s \n", serv.listenAddress)
	log.Fatal(http.ListenAndServe(serv.listenAddress, router))
}
