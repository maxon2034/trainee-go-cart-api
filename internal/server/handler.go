package server

import "net/http"

type Handler interface {
	View(http.ResponseWriter, *http.Request)
	Create(http.ResponseWriter, *http.Request)
	AddItem(http.ResponseWriter, *http.Request)
	UpdateItem(http.ResponseWriter, *http.Request)
	DeleteItem(http.ResponseWriter, *http.Request)
	//Calculate(http.ResponseWriter, *http.Request)
}
