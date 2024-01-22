package helpers

import (
	"encoding/json"
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

	client := &http.Client{}

	req, err := http.NewRequest("POST", webOrderURL, nil)
	if err != nil {
		return err
	}

	// Add HTTP headers to the request
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return nil
}
