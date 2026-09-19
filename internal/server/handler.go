package server

import "net/http"

type Handler interface {
	View(http.ResponseWriter, *http.Request)
	Create(http.ResponseWriter, *http.Request)
	//Add(http.ResponseWriter, *http.Request)
	//Update(http.ResponseWriter, *http.Request)
	//Calculate(http.ResponseWriter, *http.Request)
	//Delete(http.ResponseWriter, *http.Request)
}
