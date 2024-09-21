package httpserver

import (
	"go-waf/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	config  *config.Config
	server  *http.Server
	handler *gin.Engine

	notify chan error
}

func NewHttpServer(conf *config.Config) *HttpServer {
	httpserver := &HttpServer{
		config: conf,
		server: &http.Server{
			Addr: conf.ADDR,
		},
		notify: make(chan error),
	}

	return httpserver
}

func (h *HttpServer) SetHandler(handler *gin.Engine) {
	h.handler = handler
}

func (h *HttpServer) execute() {
	h.server.Handler = h.handler

	var err error
	if h.config.USE_SSL {
		err = h.server.ListenAndServeTLS(h.config.SSL_CERT, h.config.SSL_KEY)
	} else {
		err = h.server.ListenAndServe()
	}

	h.notify <- err
}

func (h *HttpServer) Start() {
	go h.execute()
}

func (h *HttpServer) Notify() <-chan error {
	return h.notify
}

func (h *HttpServer) Stop() {
	h.notify <- h.server.Close()
}
