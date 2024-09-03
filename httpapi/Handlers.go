package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/a-h/templ"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/janicaleksander/StocksHelp/auth"
	"github.com/janicaleksander/StocksHelp/customType"
	"github.com/janicaleksander/StocksHelp/static/components"
	"github.com/janicaleksander/StocksHelp/user"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

func GetUserID(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie("jwt-token")
	if err != nil {
		return uuid.Nil, err
	}
	tokenStr := cookie.Value
	claims := &auth.Claims{}
	_, err = jwt.ParseWithClaims(tokenStr, claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		},
	)
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}
func IsLogged(r *http.Request) (bool, error) {
	cookie, err := r.Cookie("jwt-token")
	if err != nil {
		return false, err
	}
	tokenStr := cookie.Value
	claims := &auth.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		},
	)

	if err != nil {
		return false, err
	}
	if !token.Valid {
		return false, err
	}

	return true, nil
}
func (serv *Server) home(w http.ResponseWriter, r *http.Request) error {
	_ = []customType.Stock{
		customType.Stock{
			Name:      "KAVIX",
			Price:     2343.534,
			PricePrev: 2345.234567,
		},
		customType.Stock{
			Name:      "RAVIX",
			Price:     12343.534,
			PricePrev: 25345.234567,
		},
		customType.Stock{
			Name:      "GAVIX",
			Price:     343.534,
			PricePrev: 25.234567,
		},
	}
	Render(w, r, components.Home())

	/*	go func() {
			serv.Hub.MakeCurrencyRequest("CHECK", "KAVIX")
			wg.Done()
		}()

		go func() {
			serv.Hub.MakeCurrencyRequest("BUY", "RIVEX")

		}()*/
	/*	err := serv.Hub.Storage.RegisterUser(*user.NewUser("111ahdmin", "111admin@ahjdmin.pl", "admin"))
		if err != nil {
			return err
		}
		id, err := serv.Hub.Storage.LoginUser("111admin@ahjdmin.pl", "admin")
		if err != nil {
			log.Fatal(err)
		}
		err = auth.CreateJWTCookieUser(w, r, id)
		if err != nil {
			log.Print(err)
		}
		id, err = GetUserID(r)
		fmt.Print(id)
		s := time.Now()
		resp, err := serv.Hub.MakeCurrencyRequest("market1", "BUY", "ZORAX")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(resp)
		price := resp.Data.(float64)
		fmt.Println("P", price)
		fmt.Println(time.Since(s))
		if err := serv.Hub.Storage.BuyResource(id, "ZORAX", 1, price); err != nil {
			log.Fatal(err)
		}
			if err := serv.Hub.Storage.SellResource(id, "KAVIX", 150, 3456.0); err != nil {
			log.Fatal(err)
		}
		return nil*/
	return nil
}

func Render(w http.ResponseWriter, r *http.Request, component templ.Component) error {
	return component.Render(r.Context(), w)
}

func (serv *Server) temp(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		isUp := rand.Intn(2) == 0
		_ = []customType.Stock{
			customType.Stock{
				Name:      "KAVIX",
				Price:     rand.Float64(),
				PricePrev: 2345.234567,
				Up:        isUp,
				Down:      !isUp,
			},
			customType.Stock{
				Name:      "RAVIX",
				Price:     rand.Float64(),
				PricePrev: 25345.234567,
				Up:        !isUp,
				Down:      isUp,
			},
			customType.Stock{
				Name:      "GAVIX",
				Price:     rand.Float64(),
				PricePrev: 25.234567,
				Up:        isUp,
				Down:      !isUp,
			},
		}

		//return Render(w, r, components.StockHome(t))
	}
	return nil
}
func (serv *Server) login(w http.ResponseWriter, r *http.Request) error {
	m := make(map[string]bool)
	if r.Method == http.MethodGet {
		logged, _ := IsLogged(r)

		if logged {
			m["U are already logged in"] = true
			Render(w, r, components.Login(m))
			return nil
		}
		Render(w, r, components.Login(m))

	}
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		id, err := serv.Hub.Storage.LoginUser(email, password)
		if err != nil {
			m["errr"] = true
			Render(w, r, components.Login(m))
			return nil
		}
		err = auth.CreateJWTCookieUser(w, r, id)
		if err != nil {
			log.Println(err)
		}
		serv.Hub.Storage.UpdateWalletBalance(23456, id)
		http.Redirect(w, r, "/home", http.StatusMovedPermanently)
	}
	return nil
}
func (serv *Server) register(w http.ResponseWriter, r *http.Request) error {
	m := make(map[string]bool)
	if r.Method == http.MethodGet {
		Render(w, r, components.Register(m))
		return nil
	}
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")
		if password != confirmPassword {
			m["Different passwords"] = true
			Render(w, r, components.Register(m))
			return nil
		}
		err := serv.Hub.Storage.RegisterUser(*user.NewUser(username, email, password))
		if err != nil {
			m["Not unique"] = true
			Render(w, r, components.Register(m))
			return nil
		}
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	return nil
}
func (serv *Server) sidebar(w http.ResponseWriter, r *http.Request) error {
	Render(w, r, components.Sidebar())
	return nil
}
func (serv *Server) dashboard(w http.ResponseWriter, r *http.Request) error {
	id, err := GetUserID(r)
	if err != nil {
		return err
	}
	username, err := serv.Hub.Storage.GetUsername(id)
	if err != nil {
		return err
	}
	Render(w, r, components.Dashboard(username, time.Now().Format("2006-01-02")))
	return nil
}

func (serv *Server) oneStock(w http.ResponseWriter, r *http.Request) error {
	s, err := serv.Hub.Storage.GetCurrencyList()
	if err != nil {
		return err
	}
	randStock := s[rand.Intn(len(s))]
	resp, err := serv.Hub.MakeCurrencyRequest("market1", "CHECK", randStock)
	if err != nil {
		return err
	}
	Render(w, r, components.Stock(randStock, resp.Data.(float64)))
	return nil
}

func (serv *Server) handleMarket(w http.ResponseWriter, r *http.Request) error {

	s, err := serv.Hub.Storage.GetCurrencyList()
	if err != nil {
		return err
	}
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}

	Render(w, r, components.Market(s))
	return nil
}

func (serv *Server) wykres(w http.ResponseWriter, r *http.Request) error {
	Render(w, r, components.Chart1([]int{35, 3454, 356, 3454, 2, 34, 5678}))
	return nil
}

func (serv *Server) stockInfo(w http.ResponseWriter, r *http.Request) error {
	currencyName := r.URL.Query().Get("cname")
	Render(w, r, components.TransactionPanel(currencyName))
	return nil
}

func (serv *Server) stockPrice(w http.ResponseWriter, r *http.Request) error {
	currencyName := r.URL.Query().Get("pstock")
	resp, err := serv.Hub.MakeCurrencyRequest("market1", "CHECK", currencyName)
	if err != nil {
		return err
	}
	price := resp.Data.(float64)
	return WriteJson(w, 200, price)
}

func (serv *Server) calculate(w http.ResponseWriter, r *http.Request) error {
	currencyName := r.URL.Query().Get("calculate")
	resp, err := serv.Hub.MakeCurrencyRequest("market1", "CHECK", currencyName)
	if err != nil {
		return err
	}
	price := resp.Data.(float64)
	quantity := r.URL.Query().Get("inputQuantity")
	floatValue, err := strconv.ParseFloat(quantity, 64)

	return WriteJson(w, 200, price*floatValue)
}
func (serv *Server) buy(w http.ResponseWriter, r *http.Request) error {
	m := make(map[string]bool)
	id, err := GetUserID(r)
	if err != nil {
		return err
	}
	currencyName := r.URL.Query().Get("buyCurrencyName")
	resp, err := serv.Hub.MakeCurrencyRequest("market1", "BUY", currencyName)
	if err != nil {
		m["Internal problems"] = true
		return Render(w, r, components.Alert(m))

	}
	price := resp.Data.(float64)
	quantity := r.URL.Query().Get("inputQuantity")
	floatValueQ, err := strconv.ParseFloat(quantity, 64)
	err = serv.Hub.Storage.BuyResource(id, currencyName, floatValueQ, -price*floatValueQ)
	if err != nil {
		m["Oh, you do not have enough money to do this"] = true
		return Render(w, r, components.Alert(m))

	}
	_ = fmt.Sprintf("You bought %v", currencyName)
	return nil
}

func (serv *Server) sell(w http.ResponseWriter, r *http.Request) error {
	id, err := GetUserID(r)
	m := make(map[string]bool)
	if err != nil {
		return err
	}
	currencyName := r.URL.Query().Get("sellCurrencyName")
	resp, err := serv.Hub.MakeCurrencyRequest("market1", "SELL", currencyName)
	if err != nil {
		log.Println(err)
		return err
	}
	price := resp.Data.(float64)
	quantity := r.URL.Query().Get("inputQuantity")

	floatValueQ, err := strconv.ParseFloat(quantity, 64)
	err = serv.Hub.Storage.SellResource(id, currencyName, floatValueQ, price*floatValueQ)
	if err != nil {
		log.Println(err)
		m["You do not have enough resource to sell"] = true
		return Render(w, r, components.Alert(m))
	}

	_ = fmt.Sprintf("You sold %v", currencyName)

	return WriteJson(w, 200, struct{}{})
}

func (serv *Server) getCurrencyOwnState(w http.ResponseWriter, r *http.Request) error {
	id, err := GetUserID(r)
	if err != nil {
		return err
	}
	currencyName := r.URL.Query().Get("name")
	q, err := serv.Hub.Storage.GetCurrencyOwnState(id, currencyName)
	if err != nil {
		return WriteJson(w, 200, 0)

	}
	return json.NewEncoder(w).Encode(q)
}
func (serv *Server) wallet(w http.ResponseWriter, r *http.Request) error {
	id, err := GetUserID(r)
	m, err := serv.Hub.Storage.GetYourStocks(id)
	if err != nil {
		return err
	}
	return Render(w, r, components.Wallet(m))
}
func (serv *Server) walletCalculate(w http.ResponseWriter, r *http.Request) error {
	id, err := GetUserID(r)
	m, err := serv.Hub.Storage.GetYourStocks(id)
	if err != nil {
		log.Println(err)
		return err
	}
	var amount float64
	for name, quantity := range m {
		resp, err := serv.Hub.MakeCurrencyRequest("market1", "CHECK", name)
		if err != nil {
			continue
		}
		amount += resp.Data.(float64) * quantity
	}
	err = serv.Hub.Storage.SetWalletBalance(amount, id)
	if err != nil {
		return err
	}
	return WriteJson(w, 200, amount)
}

func (serv *Server) Logout(w http.ResponseWriter, r *http.Request) error {
	b, err := IsLogged(r)
	m := make(map[string]bool)

	if err != nil || !b {
		m["You are not logged in"] = true
		return Render(w, r, components.Alert(m))
	}
	cookie, err := r.Cookie("jwt-token")
	if err != nil {
		return err
	}
	auth.DefaultCookie(cookie)
	http.SetCookie(w, cookie)
	w.Header().Set("HX-Redirect", "/home")
	return nil
}

func (serv *Server) history(w http.ResponseWriter, r *http.Request) error {
	id, err := GetUserID(r)
	if err != nil {
		return err
	}
	s, err := serv.Hub.Storage.GetHistory(id)
	if err != nil {
		return err
	}

	return Render(w, r, components.History(s))
}
