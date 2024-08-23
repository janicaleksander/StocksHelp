package httpapi

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/janicaleksander/StocksHelp/auth"
	"github.com/janicaleksander/StocksHelp/user"
	"log"
	"net/http"
	"os"
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
func (serv *Server) home(w http.ResponseWriter, r *http.Request) error {
	/*	go func() {
			serv.Hub.MakeCurrencyRequest("CHECK", "KAVIX")
			wg.Done()
		}()

		go func() {
			serv.Hub.MakeCurrencyRequest("BUY", "RIVEX")

		}()*/
	err := serv.Hub.Storage.RegisterUser(*user.NewUser("111ahdmin", "111admin@ahjdmin.pl", "admin"))
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
	/*	if err := serv.Hub.Storage.SellResource(id, "KAVIX", 150, 3456.0); err != nil {
		log.Fatal(err)
	}*/
	return nil
}
