package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"net/url"
	"strings"
	"testing"

	"os"

	// term "github.com/septianw/jas-terminal/package"
	"github.com/google/uuid"
	term "github.com/septianw/jas-terminal/package"
	"github.com/septianw/jas/common"
	"github.com/septianw/jas/types"

	"github.com/stretchr/testify/assert"

	// "strings"
	"reflect"

	"github.com/gin-gonic/gin"
)

type header map[string]string
type headers []header
type payload struct {
	Method string
	Url    string
	Body   io.Reader
}
type expectation struct {
	Code int
	Body string
}
type quest struct {
	pload  payload
	heads  headers
	expect expectation
}
type quests []quest

var termid string

func getArm() (*gin.Engine, *httptest.ResponseRecorder) {
	router := gin.New()
	gin.SetMode(gin.ReleaseMode)
	Router(router)

	recorder := httptest.NewRecorder()
	return router, recorder
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func doTheTest(load payload, heads headers) *httptest.ResponseRecorder {
	var router, recorder = getArm()

	req, err := http.NewRequest(load.Method, load.Url, load.Body)
	log.Printf("%+v", req)
	handleErr(err)

	if len(heads) != 0 {
		for _, head := range heads {
			for key, value := range head {
				req.Header.Set(key, value)
			}
		}
	}
	router.ServeHTTP(recorder, req)

	return recorder
}

func SetupRouter() *gin.Engine {
	return gin.New()
}

func SetEnvironment() {
	var rt types.Runtime
	var Dbconf types.Dbconf

	Dbconf.Database = "ipoint"
	Dbconf.Host = "localhost"
	Dbconf.Pass = "dummypass"
	Dbconf.Port = 3306
	Dbconf.Type = "mysql"
	Dbconf.User = "asep"

	rt.Dbconf = Dbconf
	rt.Libloc = "/home/asep/gocode/src/github.com/septianw/jas/libs"

	common.WriteRuntime(rt)
}

func UnsetEnvironment() {
	os.Remove("/tmp/shinyRuntimeFile")
}

func TestInsertFunc(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	var input term.TerminalIn
	termid = uuid.New().String()

	input.Name = "Kasir jogja47"
	input.Location = "Apt. 954"
	input.TerminalId = termid

	jsonInput, err := json.Marshal(input)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(string(jsonInput))

	q := quest{
		payload{"POST", "/api/v1/terminal/", bytes.NewBuffer(jsonInput)},
		headers{
			header{"X-terminal": "02162aaa-1719-39a3-adef-f2430324f56a"},
		},
		expectation{201, "contact post"},
	}

	rec := doTheTest(q.pload, q.heads)
	t.Log(rec)
	t.Log(rec.Code)
	t.Log(rec.Body.String())

	terminals, err := term.GetTerminal(termid, 0, 0)
	t.Log(terminals, err)
	if err != nil {
		t.Fail()
	}
	t.Log(reflect.DeepEqual(terminals[0], rec.Body.String()))
	termjson, err := json.Marshal(terminals[0])
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, q.expect.Code, rec.Code)
	assert.Equal(t, string(termjson), strings.TrimSpace(rec.Body.String()))
}

func TestGetFunc(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	// var input term.TerminalIn

	terminal, err := term.GetTerminal(termid, 0, 0)
	if err != nil {
		t.Log(terminal)
		t.Fail()
	}
	terminalJson, err := json.Marshal(terminal[0])
	if err != nil {
		t.Log(terminal)
		t.Fail()
	}

	terminals, err := term.GetTerminal("", 2, 0)
	if err != nil {
		t.Log(terminals)
		t.Fail()
	}
	terminalsJson, err := json.Marshal(terminals)
	if err != nil {
		t.Log(terminals)
		t.Fail()
	}

	qs := quests{
		quest{
			payload{"GET", fmt.Sprintf("/api/v1/terminal/%s", termid), nil},
			headers{
				header{"X-terminal": "02162aaa-1719-39a3-adef-f2430324f56a"},
			},
			expectation{200, string(terminalJson)},
		},
		quest{
			payload{"GET", "/api/v1/terminal/all/2/0", nil},
			headers{
				header{"X-terminal": "02162aaa-1719-39a3-adef-f2430324f56a"},
			},
			expectation{200, string(terminalsJson)},
		},
	}

	for _, q := range qs {
		rec := doTheTest(q.pload, q.heads)
		assert.Equal(t, q.expect.Code, rec.Code)
		assert.Equal(t, q.expect.Body, strings.TrimSpace(rec.Body.String()))
		t.Log(rec)
	}
	// assert.Equal(t, string(termjson), strings.TrimSpace(rec.Body.String()))
}

func TestTerminalPutPositive(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	var input term.TerminalUpdate

	input.Name = "Kasir jogja47"
	input.Location = "Suite 094"
	// input.TerminalId = termid

	jsonInput, err := json.Marshal(input)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	q := quest{
		payload{"PUT", fmt.Sprintf("/api/v1/terminal/%s", termid), bytes.NewBuffer(jsonInput)},
		headers{},
		expectation{200, "wow"},
	}

	rec := doTheTest(q.pload, q.heads)
	t.Log(rec)
	terminals, err := term.GetTerminal(termid, 0, 0)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	termJson, err := json.Marshal(terminals[0])
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, q.expect.Code, rec.Code)
	assert.Equal(t, string(termJson), strings.TrimSpace(rec.Body.String()))
}

func TestContactDeletePositive(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()

	// contactUpdatedJSON, err := json.Marshal(cpac.ContactOut{
	// 	LastPostID,
	// 	"Pramitha",
	// 	"Utami",
	// 	"Mr",
	// 	"konsumen",
	// })
	// common.ErrHandler(err)

	q := quest{
		payload{"DELETE", fmt.Sprintf("/api/v1/terminal/%s", termid), nil},
		headers{},
		expectation{200, "wow"},
	}

	terminals, err := term.GetTerminal(termid, 0, 0)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	termJson, err := json.Marshal(terminals[0])
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	rec := doTheTest(q.pload, q.heads)
	t.Log(rec)
	assert.Equal(t, q.expect.Code, rec.Code)
	assert.Equal(t, string(termJson), strings.TrimSpace(rec.Body.String()))
}

func TestLoginFunc(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()

	uv := url.Values{}
	uv.Add("grant_type", "password")
	uv.Add("username", "septianw")
	uv.Add("password", "password")
	uv.Add("client_id", "01d23d2208cc001ceee0b53bf2a8a306476d7f78")

	// logparam := strings.NewReader(uv.Encode())

	q := quest{
		payload{"POST", "/api/v1/terminal/login", bytes.NewBuffer([]byte(uv.Encode()))},
		headers{
			header{"X-terminal": "02162aaa-1719-39a3-adef-f2430324f56a"},
			header{"Content-Type": "application/x-www-form-urlencoded"},
		},
		expectation{201, "contact post"},
	}

	rec := doTheTest(q.pload, q.heads)

	t.Log(rec)
	t.Log(uv.Encode())
}
