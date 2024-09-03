package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/janicaleksander/StocksHelp/stockapi"
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

	switch v := val.(type) {
	case int, int32, int64, float32, float64: // Handle numeric types
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := w.Write([]byte(fmt.Sprintf("%v", v)))
		return err
	default:
		return json.NewEncoder(w).Encode(val) // Fallback to JSON encoding for non-numeric types
	}
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
	/*	err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
		}*/
	router := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))
	router.HandleFunc("/home", makeHandler(serv.home))
	router.HandleFunc("/temp", makeHandler(serv.temp))
	router.HandleFunc("/login", makeHandler(serv.login))
	router.HandleFunc("/register", makeHandler(serv.register))
	router.HandleFunc("/profile", makeHandler(serv.sidebar))
	router.HandleFunc("/dashboard", makeHandler(serv.dashboard))
	router.HandleFunc("/onestock", makeHandler(serv.oneStock))
	router.HandleFunc("/market", makeHandler(serv.handleMarket))
	router.HandleFunc("/wykres", makeHandler(serv.wykres))
	router.HandleFunc("/stock", makeHandler(serv.stockInfo))
	router.HandleFunc("/stockPrice", makeHandler(serv.stockPrice))
	router.HandleFunc("/calculate", makeHandler(serv.calculate))
	router.HandleFunc("/buy", makeHandler(serv.buy))
	router.HandleFunc("/sell", makeHandler(serv.sell))
	router.HandleFunc("/getCurrencyState", makeHandler(serv.getCurrencyOwnState))
	router.HandleFunc("/wallet", makeHandler(serv.wallet))
	router.HandleFunc("/walletCalculate", makeHandler(serv.walletCalculate))
	router.HandleFunc("/logout", makeHandler(serv.Logout))
	router.HandleFunc("/history", makeHandler(serv.history))

	log.Printf("Listening to %s \n", serv.listenAddress)
	log.Fatal(http.ListenAndServe(serv.listenAddress, router))
}
