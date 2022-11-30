package auth

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/eqr/eqr-auth/config"
	"github.com/gin-gonic/gin"
)

const basicAuthUserKey = "basic_user"

func AuthorizeBasic(cfg *config.Config, loginService LoginService) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			basicErr(c)
			return
		}

		logged, id := loginService.LoginUser(username, password)
		if !logged {
			log.Printf("basic auth: incorrect credentials")
			basicErr(c)
			return
		} else {
			c.Set(basicAuthUserKey, id)
		}
	}
}

func GetBasicUserId(c *gin.Context) (int64, error) {
	if val, exists := c.Get(basicAuthUserKey); !exists {
		return 0, fmt.Errorf("not found in context")
	} else {
		r, err := strconv.ParseInt(fmt.Sprintf("%v", val), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("wrong user id in basic auth: %v", val)
		}

		return r, nil
	}
}

func basicErr(c *gin.Context) {
	realm := "Authorization Required"
	realm = "Basic realm=" + strconv.Quote(realm)

	c.Header("WWW-Authenticate", realm)
	c.AbortWithStatus(http.StatusUnauthorized)
}
