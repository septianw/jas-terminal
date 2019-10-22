package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"os"

	// term "github.com/septianw/jas-terminal/package"
	"github.com/septianw/jas/common"
	"github.com/septianw/jas/types"

	// "github.com/stretchr/testify/assert"

	// "strings"

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

var LastPostID int64

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

	// ci, err := cpac.FindContact(contactIn)
	// if err != nil || len(ci) == 0 {
	// 	t.Fail()
	// }
	// t.Logf("\n%+v\n", ci)
	// LastPostID = ci[0].Id
	// cjson, err := json.Marshal(ci[0])
	// if err != nil {
	// 	t.Fail()
	// }

	// assert.Equal(t, q.expect.Code, rec.Code)
	// assert.Equal(t, string(cjson)+"\n", rec.Body.String())
}
