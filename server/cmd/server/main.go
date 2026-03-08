package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

type ExchangeRate struct {
	UsdBrl struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}
type ExchangeRateSave struct {
	quote      float64 `json:"quote"`
	id         int64   `json:"id"`
	created_at string  `json:"created_at"`
}

func main() {
	fmt.Println("Welcome Lana init")

	db = InitDB()
	defer db.Close()

	http.HandleFunc("/cotacao", HandlerQuote)
	http.ListenAndServe(":8080", nil)
	fmt.Println("Server running in the port 8080...")

}

func Quotation(ctx context.Context) {
	fmt.Println("Quotation...")
	select {
	case <-ctx.Done():
		fmt.Println("Quotation cancelled. Timeout.")
		return
	case <-time.After(time.Millisecond * 10):
		fmt.Println("Quotation checked.")
	}

}
func SearchQuote(ctx context.Context) (*ExchangeRate, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200) // Initializes a timeout Context.
	Quotation(ctx)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Erro: Timeout excedido ao consultar API 200ms")
		}
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)

	if error != nil {
		return nil, error
	}

	var c ExchangeRate
	error = json.Unmarshal(body, &c)
	fmt.Println(&c)
	if error != nil {
		return nil, error
	}
	fmt.Println(&c)

	return &c, nil
}
func HandlerQuote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	quote, error := SearchQuote(ctx)
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := InsertQuote(ctx, db, *quote)
	if err != nil {
		fmt.Println("Erro ao salvar no banco:", err)
	}

	json.NewEncoder(w).Encode(quote)

	FindMany(db)

}
func InitDB() *sql.DB {
	db, err := sql.Open("sqlite", "cotacoes.db")
	if err != nil {
		panic(err)
	}

	query := `
    CREATE TABLE IF NOT EXISTS cotacoes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        quote TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	_, err = db.Exec(query)
	if err != nil {
		panic(err)
	}

	fmt.Println("Banco de dados inicializado com sucesso!")
	return db
}

func InsertQuote(ctx context.Context, db *sql.DB, exchangeRate ExchangeRate) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctx, "insert into cotacoes ( quote ) values ($1)", exchangeRate.UsdBrl.Bid)
	if err != nil {
		// Se o erro for o estouro do tempo, você loga aqui
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Erro: O banco demorou mais de 10ms!")
		}
		return err
	}
	return nil
}
func FindMany(db *sql.DB) error {
	rows, err := db.Query("select id, quote, created_at from cotacoes")
	if err != nil {
		return err
	}
	defer rows.Close()

	var exchangeRateSaves []ExchangeRateSave
	for rows.Next() {
		var er ExchangeRateSave
		err = rows.Scan(&er.id, &er.quote, &er.created_at)
		if err != nil {
			return err
		}
		exchangeRateSaves = append(exchangeRateSaves, er)
	}

	for _, ex := range exchangeRateSaves {
		fmt.Printf(" Cotações: ID[%v] COTA[%v] DATA CONSULTAR[%v]", ex.id, ex.quote, ex.created_at)
		fmt.Println()
	}

	return nil
}
