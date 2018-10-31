package service

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/microdevs/missy/component"
)

type OptionFunc func(service *Service) error

func NameOption(name string) OptionFunc {
	return func(service *Service) error {
		service.name = name
		return nil
	}
}

func ComponentHostOption() OptionFunc {
	return func(service *Service) error {
		// get host from component env
		env, err := component.Env(Component)
		if err != nil {
			return nil
		}
		service.Host = env.Get("LISTEN_HOST")
		return nil
	}
}

func ComponentPortOption() OptionFunc {
	return func(service *Service) error {
		env, err := component.Env(Component)
		if err != nil {
			return nil
		}
		service.Port = env.Get("LISTEN_PORT")
		return nil
	}
}

func PrometheusOption(name string) OptionFunc {
	return func(service *Service) error {
		service.Prometheus = NewPrometheus(name)
		return nil
	}
}

func RouterOption(router *mux.Router) OptionFunc {
	return func(service *Service) error {
		service.Router = router
		return nil
	}
}

func ServerMuxOption(mux *http.ServeMux) OptionFunc {
	return func(service *Service) error {
		service.ServeMux = mux
		return nil
	}
}
