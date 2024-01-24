package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	oko "github.com/OneKonsole/order-model"
)

// ===========================================================================================================
// Helper to create a HTTP error message. The message will be sent as JSON
// Parameters:
//
//	w (http.ResponseWriter) : Helper object to create HTTP responses
//	code (int) : HTTP code to send
//	message (string) : Error message to send
//
// Examples:
//
//	respondWithError(w, 500, "Couldn't process the order")
//
// ===========================================================================================================
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

// ===========================================================================================================
// Helper to create JSON HTTP responses
// Parameters:
//
//	w (http.ResponseWriter) : Helper object to create HTTP responses
//	code (int) : HTTP code to send
//	payload (interface) : Data to answer with
//
// Examples:
//
//	respondWithJSON(w, 200, new Order(xx,xx,xx,xx)")
//
// ===========================================================================================================
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func LaunchOrder(webOrderURL string, order *oko.Order) error {
	fmt.Printf("[INFO] Trying to call web order for order %d on : %s\n", order.ID, webOrderURL)

	client := &http.Client{}

	orderJSON, err := json.Marshal(order)

	if err != nil {
		errMessage := "[ERROR] Could not marshal order to JSON when producing order.\n"
		fmt.Print(errMessage)
		return err
	}
	req, err := http.NewRequest("POST", webOrderURL, bytes.NewBuffer(orderJSON))
	if err != nil {
		fmt.Printf("[ERROR] Could not initiate a request to web order for order %d\n", order.ID)
		return err
	}

	// Add HTTP headers to the request
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] Could not make request to web order for order %d.\n", order.ID)
		return err
	}

	fmt.Printf("[INFO] Made request to web order for order %d. Status code : %d\n", order.ID, res.StatusCode)

	defer res.Body.Close()

	return nil
}
