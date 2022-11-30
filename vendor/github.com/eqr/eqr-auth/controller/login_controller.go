package controller

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/eqr/eqr-auth/auth"
	"github.com/eqr/eqr-auth/config"

	"github.com/eqr/eqr-shared/web_common"
	"github.com/gin-gonic/gin"
)

type LoginController interface {
	Login(*gin.Context) (string, error)
	Logout(*gin.Context)
}

type loginController struct {
	loginService auth.LoginService
	JWTService   auth.JWTService
	Host         string
}

func LoginHandler(loginService auth.LoginService, jwtService auth.JWTService, cfg *config.Config) *loginController {
	return &loginController{
		loginService: loginService,
		JWTService:   jwtService,
		Host:         cfg.Deploy.Host,
	}
}

func (controller loginController) Login(ctx *gin.Context) (string, error) {
	var credential auth.LoginCredential
	if err := ctx.ShouldBind(&credential); err != nil {
		return "", errors.New("no data found")
	}
	if isUserAuthenticated, userId := controller.loginService.LoginUser(credential.Login, credential.Password); isUserAuthenticated {
		if token, err := controller.JWTService.GenerateToken(credential.Login, userId); err != nil {
			return "", fmt.Errorf("token generation error: %v", err)
		} else {
			return token, nil
		}
	}

	return "", errors.New("user is not authenticated")
}

func (controller loginController) Logout(ctx *gin.Context) {
	auth.SetAuthCookie("", ctx, controller.Host)
}

func login(controller *loginController) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if token, err := controller.Login(ctx); err == nil {
			auth.SetAuthCookie(token, ctx, controller.Host)
			web_common.Redirect(ctx, "/")
			return
		} else {
			time.Sleep(5 * time.Second)
			web_common.ShowErrorMessage(ctx, "Incorrect credentials")
		}
	}
}

func logout(controller *loginController) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller.Logout(ctx)
		web_common.Redirect(ctx, "/")
	}
}

func showLoginPage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(
			http.StatusOK,
			"login.html",
			gin.H{
				"title": "Login",
			},
		)
	}
}

func LoginSetup(server *gin.Engine, cfg *config.Config, loginService auth.LoginService) {
	jwtService := auth.JWTAuthService(cfg)
	loginController := LoginHandler(loginService, jwtService, cfg)
	server.POST("/login", login(loginController))
	server.GET("/login", showLoginPage())
	server.GET("/logout", logout(loginController))
}
