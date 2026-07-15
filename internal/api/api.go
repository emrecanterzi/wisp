package api

import (
	"fmt"
	"net/http"
)

type API struct {
	mux *http.ServeMux
}

func NewAPI() *API {
	return &API{
		mux: http.NewServeMux(),
	}
}

func (a *API) Start() {
	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", a.mux)
}

func (a *API) Stop() {

}

func (a *API) RegisterHandler(pattern string, handler http.Handler) {
	a.mux.Handle(pattern, handler)
}

func (a *API) RegisterFuncHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	a.mux.HandleFunc(pattern, handler)
}
