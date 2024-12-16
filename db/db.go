package db

import (
	"cmp"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/janicaleksander/StocksHelp/charts"
	"github.com/janicaleksander/StocksHelp/customType"
	"github.com/janicaleksander/StocksHelp/user"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Storage interface {
	UpdatePrice(name string, price float64) error
	CheckFirst() (bool, error)
	SetDefault(m map[string]float64) error
	GetState() (map[string]float64, error)
	RegisterUser(user user.User) error
	LoginUser(email, password string) (uuid.UUID, error)
	SellResource(userID uuid.UUID, name string, q float64, sellingPrice float64) error
	BuyResource(userID uuid.UUID, name string, q float64, purchasePrice float64) error
	CheckBalance(userID uuid.UUID) (float64, error)
	UpdateWalletBalance(x float64, userID uuid.UUID) error
	GetCurrencyList() ([]string, error)
	ChartData(name string) ([]customType.ChartStockInfo, error)
	GetCurrencyOwnState(id uuid.UUID, name string) (float64, error)
	GetYourStocks(userID uuid.UUID) (map[string]float64, error)
	GetUsername(userID uuid.UUID) (string, error)
	SetWalletBalance(x float64, userID uuid.UUID) error
	GetHistory(userID uuid.UUID) ([]customType.TransactionHistory, error)
	GetCurrencyHistory(name string) ([]charts.KlineData, error)
}

type Postgres struct {
	db *sql.DB
}

func NewDB() (*Postgres, error) {

	/*	err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}*/

	dbUser := os.Getenv("DBUSER")
	dbPassword := os.Getenv("DBPASSWORD")
	dbHost := os.Getenv("DBHOST")
	dbName := os.Getenv("DBNAME")
	dbPort := os.Getenv("DBPORT")

	port, _ := strconv.Atoi(dbPort)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=require",
		dbHost, port, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Postgres{db: db}, nil

}
func (p *Postgres) Init() {
	p.CreateCurrencyTable()
	p.CreateUserTable()
	p.CreateResourceTable()
	p.CreateWalletTable()
	p.CreateHistoryTable()
	p.CreateCurrencyHistoryTable()

}

func (p *Postgres) CreateCurrencyTable() {
	query := `CREATE TABLE IF NOT EXISTS currency_table (
        id SERIAL PRIMARY KEY,       
        currency_name VARCHAR(50) NOT NULL,
        exchange_price_prev DECIMAL(18, 6) DEFAULT 0.000,
    	exchange_price DECIMAL(18,6)
    );`

	_, err := p.db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
func (p *Postgres) CreateResourceTable() {
	query := `CREATE TABLE IF NOT EXISTS resource_table (
    user_id UUID,
    resource VARCHAR(50) NOT NULL,
    quantity DECIMAL(18,6)
);`

	_, err := p.db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
	/*	query = `CREATE UNIQUE INDEX user_resource_idx ON resource_table(user_id, resource);`
		_, err = p.db.Exec(query
		if err != nil {
			log.Fatal(err)
		}*/
}
func (p *Postgres) CreateWalletTable() {
	query := `CREATE TABLE IF NOT EXISTS wallet_table (
    user_id UUID,
    money DECIMAL(18,6) 
);`
	_, err := p.db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
func (p *Postgres) CreateHistoryTable() {
	query := `CREATE TABLE IF NOT EXISTS history_table (
    user_id UUID,
    resource VARCHAR(50),
    quantity DECIMAL(18,6),
    purchase_price DECIMAL(18,6) DEFAULT 0.0,
    selling_price DECIMAL(18,6) DEFAULT 0.0,
    purchase BOOLEAN,
    sale BOOLEAN,
    transaction_time TIMESTAMP WITH TIME ZONE
);`
	_, err := p.db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Postgres) CreateUserTable() {
	query := `CREATE TABLE IF NOT EXISTS user_table(
    ID UUID,
    username VARCHAR(50),
    email VARCHAR(50),
    password VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE
)`

	if _, err := p.db.Exec(query); err != nil {
		log.Fatal(err)
	}

}

func (p *Postgres) CreateCurrencyHistoryTable() {
	query := `CREATE TABLE IF NOT EXISTS currency_history (
        id SERIAL PRIMARY KEY,       
        currency_name VARCHAR(50) NOT NULL,
        exchange_price DECIMAL(18, 6) NOT NULL,
        time_at TIMESTAMPTZ DEFAULT NOW()  
    );`

	_, err := p.db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Postgres) SetDefault(m map[string]float64) error {
	s := generateDefaultInsertQuery(m)
	_, err := p.db.Exec(s)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func (p *Postgres) UpdatePrice(name string, price float64) error {
	query := `UPDATE currency_table SET exchange_price_prev = exchange_price, exchange_price = $1 WHERE currency_name = $2;`
	if _, err := p.db.Exec(query, price, name); err != nil {
		fmt.Println(err)
		return err
	}
	query = `INSERT INTO currency_history (currency_name, exchange_price) VALUES ($1, $2);`
	if _, err := p.db.Exec(query, name, price); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (p *Postgres) CheckFirst() (bool, error) {
	query := `SELECT COUNT(*) AS counts FROM currency_table;`
	row := p.db.QueryRow(query)
	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		return true, nil
	}
	return false, nil
}

func (p *Postgres) GetState() (map[string]float64, error) {
	query := `SELECT currency_name, exchange_price_prev,exchange_price FROM currency_table`
	m := make(map[string]float64)
	row, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	for row.Next() {
		var name string
		var pricePrev float64
		var price float64
		if err := row.Scan(&name, &pricePrev, &price); err != nil {
			return nil, err
		}
		m[name] = price
	}

	return m, nil
}

func generateDefaultInsertQuery(m map[string]float64) string {
	b := "INSERT INTO currency_table (currency_name, exchange_price) VALUES"
	var values []string

	for name, price := range m {
		values = append(values, fmt.Sprintf("('%s', %f)", name, price))
	}

	return b + " " + strings.Join(values, ", ")
}

func (p *Postgres) RegisterUser(user user.User) error {
	id := user.ID
	name := user.Name
	email := user.Email
	password := user.Password
	b, err := p.checkUnique(name, email)
	if err != nil {
		return err
	}
	if !b {
		return errors.New("not unique")
	}
	query := `INSERT INTO user_table (ID,username,email,password,created_at) VALUES ($1,$2,$3,$4,$5)`
	_, err = p.db.Exec(query, id, name, email, password, time.Now())
	if err != nil {
		return err
	}
	//set wallet to new user
	err = p.setBalance(id)
	if err != nil {
		return err
	}

	return nil
}
func (p *Postgres) LoginUser(email, password string) (uuid.UUID, error) {
	query := `SELECT id, password FROM user_table WHERE email = $1 LIMIT 1`

	var id uuid.UUID
	var pwd string

	if err := p.db.QueryRow(query, email).Scan(&id, &pwd); err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, fmt.Errorf("user not found")
		}
		return uuid.Nil, err
	}

	err := bcrypt.CompareHashAndPassword([]byte(pwd), []byte(password))
	if err != nil {
		return uuid.Nil, fmt.Errorf("incorrect password")
	}

	return id, nil
}

func (p *Postgres) checkUnique(name, email string) (bool, error) {
	query := `SELECT 1 FROM user_table WHERE username=$1 OR email=$2 LIMIT 1`
	var exists int
	err := p.db.QueryRow(query, name, email).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return false, errors.New("Email or username exists")
	}
	return false, nil
}

// wallet

func (p *Postgres) setBalance(userID uuid.UUID) error {
	query := `INSERT INTO wallet_table  (user_id,money) VALUES ($1 , $2)`
	_, err := p.db.Exec(query, userID, 10000.0)
	if err != nil {
		return err
	}
	return nil
}
func (p *Postgres) CheckBalance(userID uuid.UUID) (float64, error) {
	query := `SELECT money FROM wallet_table WHERE user_id =$1`
	var money float64
	err := p.db.QueryRow(query, userID).Scan(&money)
	if err != nil {
		if err == sql.ErrNoRows {
			return .0, errors.New("uuuid not exists")
		} else {
			return .0, err
		}
	}
	return money, nil
}

func (p *Postgres) BuyResource(userID uuid.UUID, name string, q float64, purchasePrice float64) error {
	var walletMoney float64
	query := `SELECT money FROM wallet_table WHERE user_id=$1`
	err := p.db.QueryRow(query, userID).Scan(&walletMoney)
	if err != nil {
		return err
	}
	fmt.Println(purchasePrice)
	fmt.Println(walletMoney)

	if walletMoney < math.Abs(purchasePrice) {
		return errors.New("not enough money")
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var currentQuantity float64
	querySelect := `SELECT quantity FROM resource_table WHERE user_id=$1 AND resource=$2`
	err = tx.QueryRow(querySelect, userID, name).Scan(&currentQuantity)

	if err == sql.ErrNoRows {
		queryInsert := `INSERT INTO resource_table (user_id, resource, quantity) VALUES ($1, $2, $3)`
		_, err = tx.Exec(queryInsert, userID, name, q)
		if err != nil {
			return err
		}
	} else if err == nil {
		queryUpdate := `UPDATE resource_table SET quantity = quantity + $3 WHERE user_id=$1 AND resource=$2`
		_, err = tx.Exec(queryUpdate, userID, name, q)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	err = p.UpdateWalletBalance(purchasePrice, userID)
	if err != nil {
		return err
	}

	queryHistory := `INSERT INTO history_table (user_id, resource, quantity, purchase_price, selling_price, purchase, sale, transaction_time)
		VALUES ($1, $2, $3, $4, 0.0, TRUE, FALSE, $5)`
	_, err = tx.Exec(queryHistory, userID, name, q, math.Abs(purchasePrice), time.Now())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
func (p *Postgres) SellResource(userID uuid.UUID, name string, q float64, sellingPrice float64) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	var quantity float64
	query := `SELECT quantity FROM resource_table WHERE resource = $1 AND user_id=$2`
	tx.QueryRow(query, name, userID).Scan(&quantity)

	if quantity < q {
		return errors.New("You do not have enugh resource to sell")
	}

	query = `UPDATE resource_table SET quantity = quantity - $1 WHERE user_id = $2 AND resource=$3`

	_, err = tx.Exec(query, q, userID, name)
	if err != nil {
		return err
	}
	err = p.UpdateWalletBalance(sellingPrice, userID)
	if err != nil {
		return err
	}
	query = `INSERT INTO history_table (user_id,resource,quantity,purchase_price,selling_price,purchase,sale,transaction_time) VALUES ($1,$2,$3,0.0,$4,FALSE,TRUE,$5)`
	_, err = tx.Exec(query, userID, name, q, sellingPrice, time.Now())
	if err != nil {
		return err
	}
	return nil

}

func (p *Postgres) UpdateWalletBalance(x float64, userID uuid.UUID) error {
	query := `UPDATE wallet_table SET money = money + $1 WHERE user_id = $2`
	_, err := p.db.Exec(query, x, userID)
	if err != nil {
		return err
	}
	return nil

}
func (p *Postgres) SetWalletBalance(x float64, userID uuid.UUID) error {
	query := `UPDATE wallet_table SET money = money + $1 WHERE user_id = $2`
	_, err := p.db.Exec(query, x, userID)
	if err != nil {
		return err
	}
	return nil

}

func (p *Postgres) ChartData(name string) ([]customType.ChartStockInfo, error) {
	query := `SELECT (exchange_price,time_at) FROM currency_history WHERE currency_name=$1 AND WHERE time_at >= CURRENT_DATE - INTERVAL '30 days'`
	rows, err := p.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	var s []customType.ChartStockInfo
	for rows.Next() {
		var c customType.ChartStockInfo
		var price float64
		var t time.Time
		err := rows.Scan(&price, &t)
		if err != nil {
			log.Print(err)
			continue
		}
		c.Name = name
		c.Price = price
		c.TimeAt = t

		s = append(s, c)

	}
	return s, nil

}

func (p *Postgres) GetCurrencyList() ([]string, error) {
	query := `SELECT currency_name FROM currency_table`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	var s []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			log.Println(err)
			continue
		}
		s = append(s, name)
	}
	return s, nil
}

func (p *Postgres) GetCurrencyOwnState(id uuid.UUID, name string) (float64, error) {
	query := `SELECT quantity FROM resource_table WHERE user_id = $1 AND resource = $2`
	var quantity float64
	err := p.db.QueryRow(query, id, name).Scan(&quantity)
	if err != nil {
		return .0, err
	}
	return quantity, nil
}

func (p *Postgres) GetYourStocks(userID uuid.UUID) (map[string]float64, error) {
	m := make(map[string]float64)
	query := `SELECT resource,quantity FROM resource_table WHERE user_id = $1`
	rows, err := p.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var name string
		var quantity float64
		err = rows.Scan(&name, &quantity)
		if err != nil {
			continue
		}
		m[name] = quantity
	}

	return m, err

}

func (p *Postgres) GetUsername(userID uuid.UUID) (string, error) {
	query := `SELECT username FROM user_table WHERE id=$1`
	var username string
	err := p.db.QueryRow(query, userID).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}
func (p *Postgres) GetHistory(userID uuid.UUID) ([]customType.TransactionHistory, error) {
	query := `SELECT resource,quantity,purchase_price,selling_price,purchase,sale,transaction_time FROM history_table WHERE user_id=$1`
	rows, err := p.db.Query(query, userID)
	if err != nil {
		return []customType.TransactionHistory{}, err
	}
	var s []customType.TransactionHistory
	for rows.Next() {
		var t customType.TransactionHistory
		var Resource string
		var Quantity float64
		var PurchasePrice float64
		var SellingPrice float64
		var Purchase bool
		var Sale bool
		var TransactionTime time.Time
		err := rows.Scan(&Resource, &Quantity, &PurchasePrice, &SellingPrice, &Purchase, &Sale, &TransactionTime)
		if err != nil {
			continue
		}
		t.Resource = Resource
		t.Quantity = Quantity
		t.PurchasePrice = PurchasePrice
		t.SellingPrice = SellingPrice
		t.Purchase = Purchase
		t.Sale = Sale
		t.TransactionTime = TransactionTime
		s = append(s, t)
	}
	return s, nil

}

func (p *Postgres) GetCurrencyHistory(name string) ([]charts.KlineData, error) {
	//query := `SELECT DATE(time_at) AS date, ARRAY_AGG(exchange_price ORDER BY time_at) AS prices,  ARRAY_AGG(time_at ORDER BY time_at) AS times FROM currency_history WHERE currency_name = $1 GROUP BY DATE(time_at) ORDER BY date DESC;`
	query := `SELECT DATE(time_at) AS date, 
       TO_JSON(ARRAY_AGG(exchange_price ORDER BY time_at)) AS prices_json, 
       TO_JSON(ARRAY_AGG(time_at ORDER BY time_at)) AS times_json
FROM currency_history 
WHERE currency_name = $1 
GROUP BY DATE(time_at) 
ORDER BY date ASC;
`
	rows, err := p.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	var s []charts.KlineData
	for rows.Next() {
		var d charts.KlineData
		var date time.Time
		var pricesJSON, timesJSON string
		var prices []float64
		var times []time.Time
		var openPrice float64
		var closePrice float64

		err = rows.Scan(&date, &pricesJSON, &timesJSON)
		if err != nil {
			continue
		}
		if err := json.Unmarshal([]byte(pricesJSON), &prices); err != nil {
			return nil, err
		}
		openPrice = prices[0]
		closePrice = prices[len(prices)-1]
		if err := json.Unmarshal([]byte(timesJSON), &times); err != nil {
			return nil, err
		}
		d.Date = date.Format("2006-01-02")

		slices.SortFunc(prices, func(a, b float64) int {
			return cmp.Compare(a, b)
		})

		if len(prices) > 0 {
			maxV := prices[len(prices)-1]
			minV := prices[0]
			d.Data = [4]float64{openPrice, closePrice, minV, maxV}
		} else {
			d.Data = [4]float64{0, 0, 0, 0}
		}

		s = append(s, d)
	}
	return s, nil
}
