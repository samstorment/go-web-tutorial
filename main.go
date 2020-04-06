// go build -o a; ./a
package main

import (
	"net/http"
	"./routes"
	"./models"
	"./utils"
)

func main() {
	
	// initialize DB client
	models.Init()
	// Load all the html templates in the templates folder
	utils.LoadTemplates("templates/*.html")
	// get the router that has all of our routes
	router := routes.NewRouter()
	// declare router as our default router
	http.Handle("/", router)
	http.ListenAndServe(":8080", nil);
}

