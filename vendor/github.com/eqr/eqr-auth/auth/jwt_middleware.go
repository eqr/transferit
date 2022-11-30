package auth

import (
	"fmt"
	"log"
	"strconv"

	"github.com/eqr/eqr-auth/config"
	"github.com/eqr/eqr-shared/web_common"
	"github.com/golang-jwt/jwt"

	"github.com/gin-gonic/gin"
)

const AuthCookie = "Authentication"

func AuthorizeJWT(cfg *config.Config, loginService LoginService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(AuthCookie)
		if tokenString == "" || err != nil {
			web_common.Redirect(c, "/login")
			return
		}

		service := JWTAuthService(cfg)
		token, err := service.ValidateToken(tokenString)
		if err != nil {
			log.Printf("error validating token: %v", err.Error())
			web_common.Redirect(c, "/login")
			return
		}

		if !token.Valid {
			log.Println("user token is invalid: ", tokenString)
			web_common.Redirect(c, "/login")
			return
		}

		id, err := setUserId(token, c)
		if err != nil {
			log.Println(err)
			web_common.Redirect(c, "/login")
			return
		}

		login, err := loginService.GetUserLogin(id)
		if err != nil {
			log.Println(err)
			web_common.Redirect(c, "/login")
			return
		}

		tokenValue, err := service.GenerateToken(login, id)
		if err != nil {
			log.Println(err)
			web_common.Redirect(c, "/login")
			return
		}

		SetAuthCookie(tokenValue, c, cfg.Deploy.Host)
	}
}

func setUserId(token *jwt.Token, c *gin.Context) (uint64, error) {
	claims := token.Claims.(jwt.MapClaims)
	userId := claims["userId"]
	c.Set("UserId", userId)
	r, err := strconv.ParseInt(fmt.Sprintf("%v", userId), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("wrong user id in claims: %v", userId)
	}

	return uint64(r), nil
}

func GetUserId(c *gin.Context) (uint64, error) {
	if userId, exists := c.Get("UserId"); !exists {
		err := fmt.Errorf("no user id set up in request context")
		web_common.ShowError(c, err)
		return 0, err
	} else {
		if id, err := strconv.Atoi(fmt.Sprintf("%v", userId)); err != nil {
			err = fmt.Errorf("Incorrect user id in request context: %v", userId)
			web_common.ShowError(c, err)
			return 0, err
		} else {
			return uint64(id), nil
		}
	}
}

func SetAuthCookie(token string, c *gin.Context, host string) {
	c.SetCookie(AuthCookie, token, 360000, "/", host, true, false)
}
