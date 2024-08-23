package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/janicaleksander/StocksHelp/user"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
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
}

type Postgres struct {
	db *sql.DB
}

func NewDB() (*Postgres, error) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

/*
	func generateUpdateQuery(m map[string]float64) string {
		b := "UPDATE currency_table SET exchange_price = CASE "
		mid := " END WHERE currency_name IN "
		var q []string
		var i []string
		for name, value := range m {
			i = append(i, fmt.Sprintf("'%s'", name))
			s := fmt.Sprintf("WHEN currency_name = '%s' THEN %.2f", name, value)
			q = append(q, s)
		}
		e := b + strings.Join(q, " ") + mid + "(" + strings.Join(i, ",") + ")" + ";"
		return e
	}
*/
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
	if err != nil || !b {
		log.Print(err)
		return err
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
	_, err := p.db.Exec(query, userID, 0.0)
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

// assuming that we have enough money to buy
func (p *Postgres) BuyResource(userID uuid.UUID, name string, q float64, purchasePrice float64) error {
	// add resource to resource_table
	query := `INSERT INTO resource_table (user_id, resource, quantity) 
VALUES ($1, $2, $3) 
ON CONFLICT (user_id, resource) 
DO UPDATE SET quantity = resource_table.quantity + EXCLUDED.quantity;
`
	_, err := p.db.Exec(query, userID, name, q)
	if err != nil {
		return err
	}
	// change money in wallet
	err = p.UpdateWalletBalance(purchasePrice, userID)
	if err != nil {
		return err
	}
	// add transaction to history
	query = `INSERT INTO history_table (user_id,resource,quantity,purchase_price,selling_price,purchase,sale,transaction_time) VALUES ($1,$2,$3,$4,0.0,TRUE,FALSE,$5)`
	_, err = p.db.Exec(query, userID, name, q, purchasePrice, time.Now())
	if err != nil {
		return err
	}
	return nil

}

// asume that we have enough resource to sell
func (p *Postgres) SellResource(userID uuid.UUID, name string, q float64, sellingPrice float64) error {
	// add resource to resource_table
	query := `UPDATE resource_table SET quantity = quantity - $1 WHERE user_id = $2 AND resource=$3`

	_, err := p.db.Exec(query, q, userID, name)
	if err != nil {
		return err
	}
	// change money in wallet
	err = p.UpdateWalletBalance(sellingPrice, userID)
	if err != nil {
		return err
	}
	// add transaction to history
	query = `INSERT INTO history_table (user_id,resource,quantity,purchase_price,selling_price,purchase,sale,transaction_time) VALUES ($1,$2,$3,0.0,$4,FALSE,TRUE,$5)`
	_, err = p.db.Exec(query, userID, name, q, sellingPrice, time.Now())
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
