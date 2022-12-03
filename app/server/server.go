package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"path"
	"time"

	"github.com/boltdb/bolt"
	"github.com/eqr/eqr-auth/auth"
	authConfig "github.com/eqr/eqr-auth/config"
	authController "github.com/eqr/eqr-auth/controller"
	authService "github.com/eqr/eqr-auth/service"
	"github.com/eqr/transferit/app/config"
	"github.com/eqr/transferit/app/service"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router       *gin.Engine
	url          string
	internalPort int
}

func New(cfg *config.Config, authCfg *authConfig.Config) (*Server, error) {
	log.Println("starting server")

	db, err := bolt.Open(cfg.Database.Path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("error creating db %v: %v", cfg.Database.Path, err.Error())
	}

	router := gin.New()

	templatesPath := path.Join(cfg.Templates.Path, "*")
	router.LoadHTMLGlob(templatesPath)

	loginService := auth.NewLoginService(db)
	authorized := router.Group("/", auth.AuthorizeJWT(authCfg, loginService))
	authorized.GET("/", showIndex)

	authController.LoginSetup(router, authCfg, loginService)
	url := fmt.Sprintf("%v:%d", cfg.Server.Host, cfg.Server.Port)
	log.Println("running server on ", url)

	if err := authService.SetupRpc(loginService); err != nil {
		return nil, fmt.Errorf("cannot set up internal service: %w", err)
	}

	tranferService := service.New()

	if err := rpc.Register(tranferService); err != nil {
		return nil, fmt.Errorf("cannot set up transfer service: %w", err)
	}

	return &Server{
		router:       router,
		url:          url,
		internalPort: cfg.Server.InternalPort,
	}, nil
}

func (srv *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", srv.internalPort))
	if err != nil {
		return fmt.Errorf("error running internal service: %w", err.Error())
	}

	defer listener.Close()
	go rpc.Accept(listener)

	transferListener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8083))
	if err != nil {
		return fmt.Errorf("error running transfer service: %w", err.Error())
	}

	defer transferListener.Close()
	go rpc.Accept(transferListener)

	err = srv.router.Run(srv.url)
	if err != nil {
		log.Fatal("error running server: ", err.Error())
		return err
	}

	return nil
}

func showIndex(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"index.html",
		gin.H{
			"message": "Index loaded succesfully",
		},
	)
}
