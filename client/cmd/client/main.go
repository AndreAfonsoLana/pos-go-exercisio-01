package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Value struct {
	Valor float64 `json:"valor"`
}
type ExchangeRate struct {
	UsdBrl struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

var NameFile = "cotacao.txt"

func main() {
	fmt.Println("Starting CLIENT USD-BRL ...")

	quoteNow, err := GetQuotation()
	if err != nil {
		panic(err)
	}
	result := CreateFile()
	fmt.Printf("Result CreateFile: %t\n", result)
	if result {
		SaveFile(*quoteNow)
	}
	ReadFile()

}
func GetQuotation() (*Value, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300) // Initializes a timeout Context.
	Quotation(ctx)
	defer cancel()

	req, error := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil) //Prepare Request
	if error != nil {
		return nil, error
	}

	resp, err := http.DefaultClient.Do(req) //Make request
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //Close request

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var value ExchangeRate
	error = json.Unmarshal(body, &value)
	if error != nil {
		return nil, error
	}
	fmt.Println(value.UsdBrl.Bid)
	valueConvert, errorConvert := strconv.ParseFloat(value.UsdBrl.Bid, 64)
	if errorConvert != nil {
		return nil, errorConvert
	}

	var valueReturn Value
	valueReturn.Valor = valueConvert

	return &valueReturn, nil
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
func CreateFile() bool {
	fileExist, err := os.ReadFile(NameFile)

	if fileExist != nil {
		return true
	}

	file, err := os.Create(NameFile)
	if err != nil {
		panic(err)
	}
	if file != nil {
		return true
	}
	return false
}
func SaveFile(value Value) error {
	file, err := os.OpenFile(NameFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	jsonData = append(jsonData, '\n')
	lengthFile, err := file.Write(jsonData)

	if err != nil {
		return err
	}
	fmt.Printf(" File created with sucess! length[%d]", lengthFile)

	defer file.Close()

	return nil
}
func ReadFile() {
	contentsFile, err := os.Open(NameFile)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(contentsFile)
	buffer := make([]byte, 50)
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			break
		}
		fmt.Println(string(buffer[:n]))
	}
}
