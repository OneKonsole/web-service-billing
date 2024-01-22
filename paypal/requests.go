package paypal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	helpers "github.com/OneKonsole/web-service-billing/helpers"
)

// ===================================================================
// Returns owner paypal access token
//
// Parameters:
//
//	(string) clientID : Paypal Client ID (owner)
//	(string) clientSecret : Paypal Client Secret (owner)
//
// Return
//
//	(string) : The access token retrieved or empty if an error occurs
//	(error) : Error during the process or nil if no error occurs
//
// Example:
//
//	accessToken, err := GetAccessToken("xxxx", "yyyy")
//
// ===================================================================
func GetAccessToken(clientID string, clientSecret string) (string, error) {
	url := "https://api-m.sandbox.paypal.com/v1/oauth2/token"

	client := &http.Client{}

	// Paypal necessary items to add in request payload
	payload := strings.NewReader("grant_type=client_credentials&ignoreCache=true&return_authn_schemes=true&return_client_metadata=true&return_unconsented_scopes=true")

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		fmt.Printf("[ERROR] Could not initialize request to get access token : %s\n", err)
		return "", err
	}

	// Add HTTP headers to the request
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+helpers.AggregateClientInformation(clientID, clientSecret))

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] Could not make request to get access token : %s\n", err)
		return "", err
	}
	defer res.Body.Close()

	// Parse response json body
	decoder := json.NewDecoder(res.Body)
	parsedBody := make(map[string]json.RawMessage)

	if err := decoder.Decode(&parsedBody); err != nil {
		fmt.Printf("[ERROR] Invalid payload: %s\n", err)
		return "", err
	}

	// Retrieve the access token from the parsed response body
	accessToken := string(parsedBody["access_token"])

	// When parsing json.RawMessage into string, double quotes are put
	// at the beginning and at the end of the string.
	accessToken = accessToken[1 : len(accessToken)-1]

	return accessToken, nil
}

// ===================================================================
// Create a Paypal order based on the order details requested.
// Asynchronously wait for paiement approval then capture it.
// Calls web-order service in order to produce the order.
//
// Parameters:
//
//	(http.ResponseWriter) w : Used to generate HTTP responses (retrieved by the router)
//	(PaypalOrderInfos) orderInfos : Information about the order to create
//	(string) clientID : Paypal Client ID (owner). Used to retrieve the access token.
//	(string) clientSecret : Paypal Client Secret (owner). Used to retrieve the access token.
//	(string) webOrderUrl : Used to contact web-order service.
//
// Used on:
//
//	(*OrderOrchestrator) o : Object that permits asynchronous management of the order
//
// Example:
//
//	orderOrchestrator.CreateOrder(w, PaypalOrderInfos{...}, "xxxx", "xxxx", "http://localhost:8010/order")
//
// ===================================================================
func (o *OrderOrchestrator) CreateOrder(
	w http.ResponseWriter,
	orderInfos PaypalOrderInfos,
	clientID string,
	clientSecret string,
	webOrderURL string,
) {
	accessToken, err := GetAccessToken(clientID, clientSecret)

	fmt.Print("[INFO] Access token retrieved\n")
	if err != nil {
		fmt.Printf("[ERROR] Invalid client information for authentication : %s\n", err)
	}

	createdOrder, err := createPaypalOrder("https://api-m.sandbox.paypal.com/v2/checkout/orders", accessToken, orderInfos)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[INFO] Paypal order created\n")
	fmt.Printf("[INFO] Web order URL : %s\n", webOrderURL)
	approvalChannel := make(chan bool)

	// Ensures synchronisation on approvalChans var (only 1 function can write at a time)
	o.mutex.Lock()
	o.approvalChans[createdOrder.OrderID] = approvalChannel
	o.mutex.Unlock()

	// Needed to capture the order later
	var captureURL string

	for _, link := range createdOrder.Links {
		if link.Rel == "capture" {
			captureURL = link.Href
		}
	}

	// Create HTTP response for the created order
	// before waiting for client's approval
	helpers.RespondWithJSON(w, http.StatusOK, map[string]string{
		"order_id": createdOrder.OrderID,
		"status":   createdOrder.Status,
	})

	// Goroutine that waits for client approval
	go func(clientCaptureURL string) {
		approved := <-approvalChannel
		if approved {
			err := captureOrder(accessToken, clientCaptureURL)
			if err != nil {
				fmt.Printf("[ERROR] Could not capture order: %s\n", err)
			}
			err = helpers.LaunchOrder(webOrderURL, &orderInfos.Order)
			if err != nil {
				fmt.Printf("[ERROR] %s\n", err)
			}
		}
	}(captureURL)
}

// ===================================================================
// Signal the order creation that the client has approved the order.
//
// Parameters:
//
//	(string) orderID : ID of the created Paypal Order
//	(http.ResponseWriter) w : Used to generate HTTP responses (retrieved by the router)
//	(*http.Request) r : HTTP request used to contact this function
//
// Used on:
//
//	(*OrderOrchestrator) o : Object that permits asynchronous management of the order
//
// Example:
//
//	orderOrchestrator.ApproveOrder("xyYxyZ",w, r)
//
// ===================================================================
func (orderOrchestrator *OrderOrchestrator) ApproveOrder(
	orderID string,
	w http.ResponseWriter,
	r *http.Request,
) {
	// Lock orderOrchestrator operations for other go routines (integrity)
	orderOrchestrator.mutex.Lock()

	// TODO : Validate order status
	if approvalChannel, ok := orderOrchestrator.approvalChans[orderID]; ok {
		// Sends approval signal to channel for this order
		approvalChannel <- true
		// Remove this channel since it has been processed
		delete(orderOrchestrator.approvalChans, orderID)
	}
	// Unlock orderOrchestrator operations for other go routines
	orderOrchestrator.mutex.Unlock()

	// HTTP Response
	helpers.RespondWithJSON(w, http.StatusOK, PaypalOrderResponse{
		Status:  "APPROVED",
		OrderID: orderID,
	})
}

// ===================================================================
// Capture the paiement once approved by the client
//
// Parameters:
//
//	(string) accessToken : Paypal merchnat access token
//	(string) captureURL : Paypal API URL used to capture the paiement
//
// Return:
//
//	(error) : Error during process or nil if no error occurs
//
// Example:
//
//	_ := captureOrder("xyYxyZxxxxYZxZ", "https://api.sandbox.paypal.com/v2/checkout/orders/xyYxyZxxxxYZxZ/capture")
//
// ===================================================================
func captureOrder(accessToken string, captureURL string) error {

	client := &http.Client{}

	req, err := http.NewRequest("POST", captureURL, nil)
	if err != nil {
		return err
	}

	// Add HTTP headers to the request
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// Actually make the request
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// Validate response status code
	if res.StatusCode != http.StatusCreated {
		return errors.New("external error capturing order")
	}

	defer res.Body.Close()

	return nil
}

// ===================================================================
// Paypal API designed request to create an order
//
// Parameters:
//
//		(string) url : Paypal API URL used to create the order
//		(string) accessToken : Paypal merchant access token
//	 (PaypalOrderInfos) : Details about the order requested
//
// Return:
//
//	(PaypalOrderResponse) : Struct used to create a consistant HTTP response
//	(error) : Error during process or nil if no error occurs
//
// Example:
//
//	orderResponse, err := captureOrder(
//		"https://api.sandbox.paypal.com/v2/checkout/orders/xyYxyZxxxxYZxZ/capture",
//		"xyYxyZxxxxYZxZ",
//		PaypalOrderInfos{...})
//
// ===================================================================
func createPaypalOrder(
	url string,
	accessToken string,
	orderInfos PaypalOrderInfos,
) (PaypalOrderResponse, error) {
	currencyCode := orderInfos.CurrencyCode
	amountValue := orderInfos.MaxAmountValue

	bodyMap := map[string]interface{}{
		"purchase_units": []map[string]interface{}{
			{
				"amount": map[string]interface{}{
					"currency_code": currencyCode,
					"value":         amountValue,
				},
			},
		},
		"intent": "CAPTURE",
	}

	bodyJson, err := json.Marshal(bodyMap)
	if err != nil {
		fmt.Printf("[ERROR] Invalid payload: %s\n", err)
		return PaypalOrderResponse{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyJson))

	if err != nil {
		fmt.Printf("[ERROR] Unable to initiate http request: %s\n", err)
		return PaypalOrderResponse{}, err
	}

	// Add request headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}

	// Actually make the request
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] Unable to initiate http client, %s\n", err)
		return PaypalOrderResponse{}, err
	}

	defer res.Body.Close()

	// Create an HTTP response for this order creation
	var orderRes PaypalOrderResponse

	// If the order was created, create the response
	if res.StatusCode == http.StatusCreated {
		decoder := json.NewDecoder(res.Body)

		err = decoder.Decode(&orderRes)
		if err != nil {
			return PaypalOrderResponse{}, err
		}
	} else {
		return PaypalOrderResponse{}, errors.New("err: order could not be created")
	}

	return orderRes, nil
}
