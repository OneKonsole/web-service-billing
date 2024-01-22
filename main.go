package main

import (
	"os"
)

var a App
var appConf AppConf

func main() {

	// Initialize application configurations and instanciations
	appConf.Initialize()
	a.AppConf = &appConf

	// Init database, field validators, etc...
	a.Initialize()

	a.Run()

	os.Exit(0)

	// accessToken, err := paypal.GetAccessToken(
	// 	"https://api-m.sandbox.paypal.com/v1/oauth2/token",
	// 	"AQaC8NCUwYIr1c9KY_7-qWf6JpBaUwaJ6ncFqFsJbq_89gviTwvOAPAgbyWpLuMHClr0zwgLPbbMSG5h",
	// 	"EF7mO0mChU1VxdY1C5GR6MFjo7jvJqf4SWx3ePY76xOiwi2vZu3Vqxvl9WKC874HykKNqATpcKxOAilx")

	// if err != nil {
	// 	fmt.Printf("Could not get access token: %s", err)
	// 	return
	// }

	// paypal.CreateOrder("https://api-m.sandbox.paypal.com/v2/checkout/orders", accessToken)
}
