package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"text/template"
	"time"
)

const templateStr = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>moscowti.me</title>
	<style>
		body { background-color: white; }
		.container { overflow: hidden;  display: flex; flex-direction: column; min-height: 75vh; align-items: center; justify-content: center; }
		.time { font-size: 20em; }
		.footer { font-family: Arial, Helvetica, sans-serif; font-size: 1em; }
	</style>
</head>
<body>
	<div class="container">
		<div class="time">{{ . }}</div>
		<div class="footer">Powered by <a href="https://www.coindesk.com/price/bitcoin">CoinDesk</a></div>
	</div>
</body>
</html>`

type Price struct {
	Time struct {
		Updated string `json:"updated"`
	} `json:"time"`
	BPI struct {
		USD struct {
			Code        string  `json:"code"`
			Symbol      string  `json:"symbol"`
			Rate        string  `json:"rate"`
			Description string  `json:"description"`
			RateFloat   float64 `json:"rate_float"`
		} `json:"USD"`
	} `json:"bpi"`
}

type Service struct {
	Client   *http.Client
	URL      string
	Template *template.Template
}

func (svc *Service) getPrice() (*Price, error) {

	req, err := http.NewRequest(http.MethodGet, svc.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := svc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var price Price
	if err := json.Unmarshal(body, &price); err != nil {
		return nil, err
	}

	return &price, nil
}

func (svc *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	price, err := svc.getPrice()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mTime := math.Floor(100000000 / price.BPI.USD.RateFloat)

	if err := svc.Template.Execute(w, mTime); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {

	tmpl, err := template.New("test").Parse(templateStr)
	if err != nil {
		log.Fatal(err)
	}

	svc := &Service{
		Client: &http.Client{
			Timeout: time.Second * 5,
		},
		URL:      `https://api.coindesk.com/v1/bpi/currentprice.json`,
		Template: tmpl,
	}

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      svc,
		Addr:         ":3000",
	}

	log.Println("listening on port 3000")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
