package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"strconv"

	"errors"

	"github.com/gin-gonic/gin"
	loc "github.com/septianw/jas-location/package"
	term "github.com/septianw/jas-terminal/package"
	usr "github.com/septianw/jas-user/package"
	"github.com/septianw/jas/common"
)

const VERSION = term.VERSION

/*
  `uid` INT NOT NULL AUTO_INCREMENT,
  `uname` VARCHAR(225) NOT NULL,
  `upass` TEXT NOT NULL,
  `contact_contactid` INT NOT NULL,
*/

/*
ERROR CODE LEGEND:
error containt 4 digits,
first digit represent error location either module or main app
1 for main app
2 for module

second digit represent error at level app or database
1 for app
2 for database

third digit represent error with input variable or variable manipulation
0 for skipping this error
1 for input validation error
2 for variable manipulation error

fourth digit represent error with logic, this type of error have
increasing error number based on which part of code that error.
0 for skipping this error
1 for unknown logical error
2 for whole operation fail, operation end unexpectedly
*/

const DATABASE_EXEC_FAIL = 2200
const MODULE_OPERATION_FAIL = 2102
const INPUT_VALIDATION_FAIL = 2110

var NOT_ACCEPTABLE = gin.H{"code": "NOT_ACCEPTABLE", "message": "You are trying to request something not acceptible here."}
var NOT_FOUND = gin.H{"code": "NOT_FOUND", "message": "You are find something we can't found it here."}

var segments []string

func Bootstrap() {
	fmt.Println("Module location bootstrap.")
}

/*
POST   /user
GET    /user/(:uid)
GET    /user/all/(:offset)/(:limit)
-----
ini masuk ke terminal
GET    /user/login
	basic auth
	return token, refresh token
-----
PUT    /user/(:uid)
DELETE /user/(:uid)
*/

func Router(r *gin.Engine) {
	r.Any("/api/v1/terminal/*path1", deflt)
}

func deflt(c *gin.Context) {
	segments := strings.Split(c.Param("path1"), "/")
	// log.Printf("\n%+v\n", c.Request.Method)
	// log.Printf("\n%+v\n", c.Param("path1"))
	// log.Printf("\n%+v\n", segments)
	// log.Printf("\n%+v\n", len(segments))

	switch c.Request.Method {
	case "POST":
		if strings.Compare(segments[1], "") == 0 {
			PostTerminalInsert(c)
		} else if strings.Compare(segments[1], "login") == 0 {
			PostLoginFunc(c)
		} else {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, loc.NOT_ACCEPTABLE)
		}
		break
	case "GET":
		if strings.Compare(segments[1], "all") == 0 {
			GetTerminalAllHandler(c)
		} else if strings.Compare(segments[1], "") != 0 {
			GetTerminalIdHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusNotAcceptable, loc.NOT_ACCEPTABLE)
		}
		break
	case "PUT":
		if strings.Compare(segments[1], "") != 0 {
			PutTerminalHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, loc.NOT_ACCEPTABLE)
		}
		break
	case "DELETE":
		if strings.Compare(segments[1], "") != 0 {
			DeleteTerminalHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, loc.NOT_ACCEPTABLE)
		}
		break
	default:
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, loc.NOT_ACCEPTABLE)
		break
	}
	// c.String(http.StatusOK, "hai")
}

func dummyResponse(c *gin.Context) {
	c.String(http.StatusOK, "wow")
}

func PostTerminalInsert(c *gin.Context) {
	var input term.TerminalIn

	if err := c.ShouldBindJSON(&input); err != nil {
		common.ErrHandler(err)
		common.SendHttpError(c, common.INPUT_VALIDATION_FAIL_CODE, errors.New("Input not valid."))
		c.Abort()
		return
	}

	terminal, err := term.InsertTerminal(input)
	if err != nil {
		if strings.Compare("Contact not found.", err.Error()) == 0 {
			common.ErrHandler(err)
			common.SendHttpError(c, common.RECORD_NOT_FOUND_CODE, errors.New("Fail to insert terminal."))
			c.Abort()
			return
		} else {
			common.ErrHandler(err)
			common.SendHttpError(c, common.DATABASE_EXEC_FAIL_CODE, errors.New("Database inserting fail."))
			c.Abort()
			return
		}
	}

	c.JSON(http.StatusCreated, terminal)
}

func PostLoginFunc(c *gin.Context) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var grant term.Grant
	var user usr.UserIn
	// var usersOut []usr.UserOut
	var err error
	var passVerified, clientVerified bool
	var tokenResponse term.TokenResponse

	if err = c.ShouldBind(&grant); err != nil {
		common.SendHttpError(c, common.INPUT_VALIDATION_FAIL_CODE, err)
		return
	}

	log.Printf("%+v", grant)

	if grant.GrantType == "password" {
		user.Uname = grant.Username

		passVerified, err = usr.VerifyUser(grant.Username, grant.Password)
		if err != nil {
			common.SendHttpError(c, common.DATABASE_EXEC_FAIL_CODE, err)
			return
		}

		if passVerified {
			clientVerified, err = term.VerifyClients(grant.ClientId, grant.ClientSecret)
			if err != nil {
				// FIXME: ini harusnya bukan database exec fail tapi ditulis begini untuk placeholder.
				common.SendHttpError(c, common.DATABASE_EXEC_FAIL_CODE, err)
				return
			}

		}

		if passVerified && clientVerified {
			// terminal := c.MustGet("terminal").(string)
			terminal := c.GetHeader("X-terminal")
			tokenResponse, err = term.IssueTokens(terminal, grant)
			if err != nil {
				// FIXME: cek, apakah error code ini sudah benar atau belum.
				common.SendHttpError(c, common.MODULE_OPERATION_FAIL_CODE, err)
				return
			}
			c.JSON(http.StatusOK, tokenResponse)
			return
		} else {
			// FIXME: buat status khusus untuk ini di common.
			c.JSON(http.StatusUnauthorized, nil)
		}
	} else {
		common.SendHttpError(c, common.NOT_ACCEPTABLE_CODE, errors.New("Currently only accept password grant."))
	}
}

func GetTerminalIdHandler(c *gin.Context) {
	var records []term.TerminalOut
	var record term.TerminalOut
	var err error
	var segments = strings.Split(c.Param("path1"), "/")
	var id string

	// FIXME: need some filtering.
	id = segments[1]

	records, err = term.GetTerminal(id, 0, 0)
	if err != nil {
		common.SendHttpError(c, common.DATABASE_EXEC_FAIL_CODE, err)
	}
	if len(records) > 0 {
		record = records[0]
	} else {
		common.SendHttpError(c, common.RECORD_NOT_FOUND_CODE, errors.New("You are find something we can't found it here."))
		return
	}

	c.JSON(http.StatusOK, record)
	return

}

func GetTerminalAllHandler(c *gin.Context) {
	var records []term.TerminalOut
	var segments = strings.Split(c.Param("path1"), "/")
	var l, o int64
	var limit, offset int
	var err error

	md, ada := c.Get("middleware")
	log.Println(md, ada)

	if len(segments) == 3 {
		offset = 0
		limit, err = strconv.Atoi(segments[2])
		if err != nil {
			common.ErrHandler(err)
			common.SendHttpError(c, common.INPUT_VALIDATION_FAIL_CODE, errors.New(
				fmt.Sprintf("%+v should be numeric", segments[2])))
			c.Abort()
			return
		}
	} else if len(segments) == 4 {
		offset, err = strconv.Atoi(segments[3])
		if err != nil {
			log.Println(err.Error())
			common.SendHttpError(c, common.INPUT_VALIDATION_FAIL_CODE, errors.New(
				fmt.Sprintf("%+v should be numeric", segments[3])))
			c.Abort()
			return
		}
		limit, err = strconv.Atoi(segments[2])
		if err != nil {
			log.Println(err.Error())
			common.SendHttpError(c, common.INPUT_VALIDATION_FAIL_CODE, errors.New(
				fmt.Sprintf("%+v should be numeric", segments[2])))
			c.Abort()
			return
		}
	} else {
		limit = 10
		offset = 0
	}

	if err == nil { // tidak ada error dari konversi
		l = int64(limit)
		o = int64(offset)
	}

	records, err = term.GetTerminal("", l, o)
	if err != nil {
		common.SendHttpError(c, common.DATABASE_EXEC_FAIL_CODE, err)
		c.Abort()
	}

	c.JSON(http.StatusOK, records)
}

func PutTerminalHandler(c *gin.Context) {
	var segments = strings.Split(c.Param("path1"), "/")
	var id string
	var input term.TerminalUpdate

	// FIXME: need some filtering.
	id = segments[1]

	if err := c.ShouldBindJSON(&input); err != nil {
		common.SendHttpError(c, common.INPUT_VALIDATION_FAIL_CODE, err)
		return
	}

	_, err := term.UpdateTerminal(id, input)

	if err != nil {
		if strings.Compare("Contact not found.", err.Error()) == 0 {
			common.SendHttpError(c, common.RECORD_NOT_FOUND_CODE, err)
			c.Abort()
			return
		} else {
			common.SendHttpError(c, common.DATABASE_EXEC_FAIL_CODE, err)
			c.Abort()
			return
		}
	}

	on, err := term.GetTerminal(id, 0, 0)
	if err != nil {
		common.SendHttpError(c, common.DATABASE_EXEC_FAIL_CODE, err)
		c.Abort()
		return
	}
	log.Println(on)

	c.JSON(http.StatusOK, on[0])
	return
}

func DeleteTerminalHandler(c *gin.Context) {
	var segments = strings.Split(c.Param("path1"), "/")
	var id string

	// FIXME: need some filtering.
	id = segments[1]

	// contacts := cpac.GetContact(id, 0, 0)
	contact, err := term.DeleteTerminal(id)
	if err != nil {
		common.SendHttpError(c, common.DATABASE_EXEC_FAIL_CODE, err)
		return
	} else if (err != nil) && (strings.Compare("Contact not found.", err.Error()) == 0) {
		common.SendHttpError(c, common.RECORD_NOT_FOUND_CODE, err)
		return
	}

	c.JSON(http.StatusOK, contact)
}
