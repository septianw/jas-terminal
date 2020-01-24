package main

import (
	"errors"
	// "fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	term "github.com/septianw/jas-terminal"

	// usr "github.com/septianw/jas-user/package"
	"github.com/septianw/jas/common"
)

type Headers struct {
	Authorization string `header:"Authorization"`
	Terminal      string `header:"x-terminal" binding:"required"`
}

/*
POST /oauth/token HTTP/1.1
Host: authorization-server.com

grant_type=password
&username=user@example.com
&password=1234luggage
&client_id=xxxxxxxxxx
&client_secret=xxxxxxxxxx
*/

/*
curl -H "Authorization: Bearer RsT5OjbzRn430zqMLgV3Ia" \
https://api.authorization-server.com/1/me
*/

func MiddleFunc() gin.HandlerFunc {
	log.Println("verify loaded.")
	return func(c *gin.Context) {
		var h Headers
		var err error
		var terminalVerified, accessTokenVerified bool
		var segments []string

		log.Println(c.Request.URL.String())
		// log.Println("wii woo wii woo")
		c.Set("middleware", "loaded")
		if strings.Compare(c.Request.URL.String(), "/ui/v1/proxy") == 0 {
			c.Next()
		}
		segments = strings.Split(c.Request.URL.String(), "/")
		log.Printf("\n%+v\n", segments)

		if err = c.ShouldBindHeader(&h); err != nil {
			common.ErrHandler(err)
			common.SendHttpError(c, common.INPUT_VALIDATION_FAIL_CODE,
				errors.New("x-terminal header not found. access forbidden."))
			c.Abort()
			return
		}

		if strings.Compare(h.Authorization, "") == 0 {
			// FIXME: Harusnya ini masuk log. siapa, kapan, dan dari terminal mana user login.
			if (strings.Compare(segments[3], "terminal") == 0) &&
				strings.Compare(segments[4], "login") == 0 {
				// ini masuk ke login
				c.Next()
				return
			} else {
				// ini masuk ke tempat lain
				common.SendHttpError(c, common.FORBIDDEN_CODE,
					errors.New("Authorization required. Please login first."))
				c.Abort()
				return
			}
		}

		if terminalVerified, err = term.VerifyTerminal(h.Terminal); !terminalVerified {
			log.Printf("\n$+v\n", err)
			log.Printf("\n$+v\n", terminalVerified)
			common.SendHttpError(c, common.INPUT_VALIDATION_FAIL_CODE,
				errors.New("Terminal not registered. Please contact authorized officer to register your terminal."))
			c.Abort()
			return
		}
		c.Set("terminal", h.Terminal)

		if strings.Compare(h.Authorization, "") != 0 {
			if accessTokenVerified, err = term.VerifyAccessToken(strings.Split(h.Authorization, " ")[1], h.Terminal); !accessTokenVerified {
				log.Printf("\n%+v\n", err)
				log.Printf("\n%+v\n", accessTokenVerified)
				common.SendHttpError(c, common.FORBIDDEN_CODE,
					errors.New("Invalid access token."))
				c.Abort()
				return
			}
		}
		c.Set("accessToken", strings.Split(h.Authorization, " ")[1])
		log.Println(c.Request.URL.Path)

		if terminalVerified && accessTokenVerified {
			c.Next()
		} else {
			common.SendHttpError(c, common.FORBIDDEN_CODE, errors.New("Unauthorized access."))
			c.Abort()
		}

		// term.VerifyClients()
		// usr.VerifyUser()
	}
}
