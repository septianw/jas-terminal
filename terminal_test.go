package terminal

import (
	"encoding/json"
	"testing"

	"log"
	"os"
	"reflect"

	"github.com/google/uuid"
	"github.com/septianw/jas/common"
	"github.com/septianw/jas/types"
)

var LastID string

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

func TestInsertTerminal(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	var termin, termonit TerminalIn
	var err error
	lid := uuid.New()

	LastID = lid.String()
	t.Log(LastID)

	termin.Name = "asmnljk"
	termin.Location = "Suite 819"
	termin.TerminalId = LastID
	_, err = InsertTerminal(termin)
	if err != nil {
		t.Logf("error insert terminal: %+v\n", err)
		t.Fail()
	}

	ters, err := GetTerminal(LastID, 0, 0)
	if err != nil {
		t.Logf("error get terminal after insert: %+v\n", err)
		t.Fail()
	}

	if len(ters) == 0 {
		t.Logf("fail to insert, last inserted: %+v\n", ters)
		t.Fail()
	}

	termonit.TerminalId = ters[0].TerminalId
	termonit.Name = ters[0].Name
	termonit.Location = ters[0].Location

	if !reflect.DeepEqual(termonit, termin) {
		t.Fail()
	}
}

func TestGetTerminal(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	Terminals, err := GetTerminal("", 2, 3)
	t.Log(Terminals, err)
	if err != nil {
		t.Logf("Error: %+v\n", err)
		t.Fail()
	}

	Terminals, err = GetTerminal(LastID, 0, 0)
	t.Log(Terminals, err)
	if err != nil {
		t.Logf("Error: %+v\n", err)
		t.Fail()
	}
}

func TestPutTerminal(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	var termin TerminalIn

	termin.Name = "mamamia"
	termin.Location = "Suite 819"
	termin.TerminalId = LastID

	ters, err := GetTerminal(LastID, 0, 0)
	if err != nil {
		t.Logf("Error: %+v\n", err)
		t.Fail()
	}

	terminal, err := UpdateTerminal(LastID, termin)
	t.Logf("Updated terminal: %+v\n", terminal)
	t.Logf("Error: %+v\n", err)

	if err != nil {
		t.Logf("Error: %+v\n", err)
		t.Fail()
	}

	t.Logf("Ters: %+v\n", ters)
	if reflect.DeepEqual(ters[0], terminal) {
		t.Fail()
	}
}

func TestDeleteTerminal(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	terminal, err := DeleteTerminal(LastID)
	if err != nil {
		t.Logf("Error: %+v\n", err)
		t.Fail()
	}
	_, err = GetTerminal(LastID, 0, 0)
	if err == nil {
		t.Logf("Error: %+v\n", err)
		t.Fail()
	}

	t.Logf("Deleted terminal: %+v", terminal)
}

func TestVerifyClient(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()
	var verified bool
	var err error

	verified, err = VerifyClients("01d23d2208cc001ceee0b53bf2a8a306476d7f78", "")
	if err != nil {
		t.Log(verified, err)
		t.Fail()
	}
	if !verified {
		t.Log(verified, err)
		t.Fail()
	}

	verified, err = VerifyClients("01d23d2208cc001ceee0b53bf2a8a306476d7f78", "59fe3666586748f79243e5d176c9ea702ee9397d70de87ffeb764c0f7bb9ba2d")
	if err != nil {
		t.Log(verified, err)
		t.Fail()
	}
	if !verified {
		t.Log(verified, err)
		t.Fail()
	}

	verified, err = VerifyClients("ini salah", "")
	if err != nil {
		t.Log(verified, err)
		t.Fail()
	}
	if verified {
		t.Log(verified, err)
		t.Fail()
	}

	verified, err = VerifyClients("ini salah", "ini lebih salah lagi")
	if err != nil {
		t.Log(verified, err)
		t.Fail()
	}
	if verified {
		t.Log(verified, err)
		t.Fail()
	}
}

func TestIssueToken(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()
	var g Grant

	g.ClientId = "01d23d2208cc001ceee0b53bf2a8a306476d7f78"
	g.Username = "septianw"
	g.Password = "$2a$10$HL4KenhsWvRlXyDlzUfa3OZHZzs7dkEb2srN8NrGrJMPwJHEfh792"
	g.GrantType = "password"

	terminalId := "1a58a82c-518d-311a-a9ec-5d3be9e4bd3e"

	response, err := IssueTokens(terminalId, g)
	if err != nil {
		t.Fail()
	}

	js, err := json.Marshal(response)
	if err != nil {
		t.Fail()
	}

	t.Logf("response: %+v\nerror: %+v\njson:%+v\n", response, err, string(js))
}
