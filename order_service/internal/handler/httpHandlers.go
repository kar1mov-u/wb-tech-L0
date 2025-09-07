package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"order_service/internal/models"
	"order_service/internal/service"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type OrderServiceHandler struct {
	service *service.OrderService
}

type HttpError struct {
	Code int
	err  error
	msg  string
}

func (he HttpError) Error() string {
	return he.msg
}

type serviceHandle func(w http.ResponseWriter, r *http.Request) error

func (fn serviceHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var httpErr HttpError
	err := fn(w, r)
	if err != nil {
		if errors.As(err, &httpErr) {
			http.Error(w, httpErr.msg, httpErr.Code)
			return
		} else {
			http.Error(w, err.Error(), 500)
		}
	}
}
func (fn serviceHandle) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn.ServeHTTP(w, r)
	}
}

func NewOrderServiceHandler(s *service.OrderService) *OrderServiceHandler {
	return &OrderServiceHandler{service: s}
}

func (h *OrderServiceHandler) GetOrder(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	order, err := h.service.GetOrder(r.Context(), id)
	if err != nil {
		return HttpError{err: err, Code: http.StatusInternalServerError, msg: "falied to retrieve order"}
	}
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(order); err != nil {
		return HttpError{err: err, Code: http.StatusInternalServerError, msg: "falied to write JSON"}
	}
	return nil
}

func (h *OrderServiceHandler) SaveOrder(w http.ResponseWriter, r *http.Request) error {
	var order models.Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		return HttpError{err: err, Code: http.StatusBadRequest, msg: "invalid input data" + err.Error()}
	}
	err = h.service.SaveOrder(r.Context(), order)
	if err != nil {
		return HttpError{err: err, Code: http.StatusInternalServerError, msg: "cannot save order" + err.Error()}
	}
	return nil
}

func (h *OrderServiceHandler) SetRoutes() http.Handler {
	chi := chi.NewRouter()
	chi.Use(middleware.Logger)
	chi.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("working"))
	})
	chi.Get("/order/{id}", serviceHandle(h.GetOrder).HandlerFunc())
	chi.Post("/order/", serviceHandle(h.SaveOrder).HandlerFunc())
	return chi
}
