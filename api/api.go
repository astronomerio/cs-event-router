package api

import (
	"context"
	"net/http"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/astronomerio/event-router/api/routes"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	log   = logrus.WithField("package", "api")
	Debug = false
)

type Client struct {
	handlers []routes.RouteHandler
}

func NewClient() *Client {
	return &Client{handlers: make([]routes.RouteHandler, 0)}
}

func (c *Client) AppendRouteHandler(rh routes.RouteHandler) {
	c.handlers = append(c.handlers, rh)
}

func (c *Client) Serve(port string, pprof bool, shutdownChan chan struct{}) error {
	logger := log.WithFields(logrus.Fields{"function": "Serve"})
	logger.Debug("Entered Serve")
	var router *gin.Engine
	if Debug {
		gin.SetMode(gin.DebugMode)
		router = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	}

	for _, handler := range c.handlers {
		handler.Register(router)
	}

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	if pprof {
		ginpprof.Wrap(router)
	}

	if string(port[0]) != ":" {
		port = ":" + port
	}
	srv := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error(err)
		}
	}()

	for range shutdownChan {
		logger.Info("Webserver shutting down")
		srv.Shutdown(context.Background())
		return nil
	}
	logger.Info("Shutdown Webserver")
	return nil
}
