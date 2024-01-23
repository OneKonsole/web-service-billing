package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	helpers "github.com/OneKonsole/web-service-billing/helpers"
	paypalOrder "github.com/OneKonsole/web-service-billing/paypal"

	"github.com/gorilla/mux"
)

type App struct {
	Router            *mux.Router
	AppConf           *AppConf
	OrderOrchestrator *paypalOrder.OrderOrchestrator
	OrderInfos        *paypalOrder.PaypalOrderInfos
}

type AppConf struct {
	ServedPort   string `json:"served_port"`           // e.g. "8010"
	WebOrderURL  string `json:"web_order_service_url"` // e.g. "http://localhost:xxxx/order
	ClientID     string
	ClientSecret string
}

func (a *App) Initialize() {
	fmt.Printf("[INFO] .... INITIALIZING APP ....\n")
	a.Router = mux.NewRouter()
	a.OrderOrchestrator = paypalOrder.NewOrderOchestrator()
	a.initializeRoutes()
}

func (appConf *AppConf) Initialize() {
	fmt.Printf("[INFO] .... INITIALIZING APP CONFIGURATIONS ....\n")
	appConf.ServedPort = os.Getenv("served_port")
	appConf.WebOrderURL = os.Getenv("web_order_service_url")
	appConf.ClientID = os.Getenv("paypal_client_id")
	appConf.ClientSecret = os.Getenv("paypal_client_secret")

	if appConf.ServedPort == "" ||
		appConf.WebOrderURL == "" ||
		appConf.ClientSecret == "" ||
		appConf.ClientID == "" {

		log.Fatal("[ERROR] Could not read env configurations\n")
	}
}

// ===========================================================================================================
// Runs the HTTP server
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// Parameters:
//
//	addr (string): Full URL to use for the server
//
// Examples:
//
//	a.Run("localhost:8010")
//
// ===========================================================================================================
func (a *App) Run() {
	log.Fatal(http.ListenAndServe(":"+a.AppConf.ServedPort, a.Router))
}

func (a *App) getPrices(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[INFO] Default prices requested\n")

	prices := helpers.NewPrices()

	helpers.RespondWithJSON(w, http.StatusOK, prices)
}

func (a *App) validatePodHealth(w http.ResponseWriter, r *http.Request) {
	helpers.RespondWithJSON(w, http.StatusOK, "")
}

func (a *App) approveOrder(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	parsedBody := make(map[string]string)

	if err := decoder.Decode(&parsedBody); err != nil {
		fmt.Printf("Invalid payload: %s", err)
	}
	fmt.Printf("[INFO] Approving order for user %s...\n", a.OrderInfos.Order.UserID)

	a.OrderOrchestrator.ApproveOrder(string(parsedBody["order_id"]), w, r)
}

func (a *App) createOrder(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&a.OrderInfos); err != nil {
		fmt.Printf("[ERROR] Invalid payload: %s\n", err)
	}

	// At the moment, control plane will everytime be enabled
	a.OrderInfos.Order.HasControlPlane = true
	// Calculating the price based on order infos
	price := helpers.CalculatePrice(&a.OrderInfos.Order)
	a.OrderInfos.MaxAmountValue = strconv.Itoa(price)
	fmt.Printf("\n[INFO] Order creation requested by %s\n   ---> Cluster name : %s\n   ---> Control plane : %s\n   ---> Monitoring : %s - %d Go\n   ---> Images storage : %d\n   ---> Alerting : %s\n   ---> Price calculated : (%d %s) \n\n",
		a.OrderInfos.Order.UserID,
		a.OrderInfos.Order.ClusterName,
		strconv.FormatBool(a.OrderInfos.Order.HasControlPlane),
		strconv.FormatBool(a.OrderInfos.Order.HasControlPlane),
		a.OrderInfos.Order.MonitoringStorage,
		a.OrderInfos.Order.ImageStorage,
		strconv.FormatBool(a.OrderInfos.Order.HasControlPlane),
		price,
		a.OrderInfos.CurrencyCode,
	)
	// Call the actual method to manage the new order
	a.OrderOrchestrator.CreateOrder(w, *a.OrderInfos, a.AppConf.ClientID, a.AppConf.ClientSecret, a.AppConf.WebOrderURL)

}

// ===========================================================================================================
// Initialize every HTTP route of our application
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// ===========================================================================================================
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/", a.validatePodHealth).Methods("GET") // Method that only returns "ok" status for kube probes
	a.Router.HandleFunc("/order/approve", a.approveOrder).Methods("POST")
	a.Router.HandleFunc("/order/create", a.createOrder).Methods("POST")
	a.Router.HandleFunc("/order/prices", a.getPrices).Methods("GET")
}
