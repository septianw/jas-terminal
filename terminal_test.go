package terminal

import (
	"encoding/json"
	"testing"

	"crypto/rand"
	"encoding/hex"
	"log"

	"os"
	"reflect"

	"github.com/google/uuid"
	"github.com/septianw/jas/common"
	"github.com/septianw/jas/types"
)

var LastID string
var Token TokenResponse
var TerminalId string

func SetEnvironment() {
	var rt types.Runtime
	var Dbconf types.Dbconf

	Dbconf.Database = "jasdev"
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

func randWord(n int) string {
	word := make([]byte, n)
	rand.Read(word)

	wordString := hex.EncodeToString(word)
	log.Println(wordString)

	return wordString
}

/*CRUD Terminal start*/
func TestInsertTerminal(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	var termin, termonit TerminalIn
	var err error
	lid := uuid.New()

	LastID = lid.String()
	t.Log(LastID)

	termin.Name = randWord(8)
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

	Terminals, err := GetTerminal("", 2, 0)
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

	var termin TerminalUpdate

	termin.Name = randWord(9)
	termin.Location = "Suite 819"
	// termin.TerminalId = LastID

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

/*CRUD Terminal stop*/

/*CRUD ClientCredential start*/

var cName string

func TestInsertClientCredential(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	cName = randWord(8)
	credential, err := InsertClientCredential(cName)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Logf("%+v", credential)
}

func TestGetClientCredential(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	credential, err := GetClientCredentials(cName)
	if err != nil {
		t.Logf("%+v\n", err)
		t.Fail()
	}

	t.Logf("%+v\n", credential)
}

func TestDeleteClientCredential(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	credential, err := DeleteClientCredentials(cName)

	if err != nil {
		t.Logf("%+v\n", err)
		t.Fail()
	}

	t.Logf("%+v\n", credential)
}

/*CRUD ClientCredential stop*/

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

// Dalam IssueToken terdapat FetchToken dan GenerateToken
func TestIssueToken(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()
	var g Grant

	g.ClientId = "2008e223b4c077f8eaf8e68a23546220"
	g.Username = "reba63"
	g.Password = "password"
	g.GrantType = "password"

	TerminalId = "2e9ba49a-9a42-4cbc-9f66-4359b22b5ff4"

	Token, err := IssueTokens(TerminalId, g)
	t.Logf("\nToken: %+v\n", Token)
	if err != nil {
		t.Fail()
	}

	js, err := json.Marshal(Token)
	if err != nil {
		t.Fail()
	}

	if reflect.DeepEqual(Token, TokenResponse{}) {
		t.Fail()
	}

	t.Logf("response: %+v\nerror: %+v\njson:%+v\n", Token, err, string(js))
}

func TestGenerateToken(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	var err error
	var tk TokenResponse

	// Generate token dari username password
	var g Grant

	g.ClientId = "2008e223b4c077f8eaf8e68a23546220"
	g.Username = "reba63"
	g.Password = "password"
	g.GrantType = "password"

	TerminalId = "2e9ba49a-9a42-4cbc-9f66-4359b22b5ff4"

	tk, err = GenerateTokens(TerminalId, g)
	t.Logf("\n%+v\n", tk)
	if err != nil {
		t.Logf("\n%+v\n", err)
		t.Logf("\n%+v\n", tk)
		t.Fail()
	}

	// Generate token dari refresh_token
	g.GrantType = "refresh_token"
	g.Username = ""
	g.Password = ""
	g.RefreshToken = tk.RefreshToken
	tk, err = GenerateTokens(TerminalId, g)
	t.Logf("\n%+v\n", tk)
	if err != nil {
		t.Logf("\n%+v\n", err)
		t.Logf("\n%+v\n", tk)
		t.Fail()
	}
	Token = tk
}

func TestVerifyToken(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	t.Logf("%+v\n", TerminalId)
	t.Logf("%+v\n", Token.AccessToken)
	t.Logf("%+v\n", Token)

	// var g Grant

	// g.ClientId = "2008e223b4c077f8eaf8e68a23546220"
	// g.Username = "reba63"
	// g.Password = "password"
	// g.GrantType = "password"

	// TerminalId = "2e9ba49a-9a42-4cbc-9f66-4359b22b5ff4"

	// Token, err := IssueTokens(TerminalId, g)
	// if err != nil {
	// 	t.Logf("\nerr: %+v\n", err)
	// }
	// if reflect.DeepEqual(Token, TokenResponse{}) {
	// 	t.Fail()
	// }

	verified, err := VerifyAccessToken(Token.AccessToken, TerminalId)
	// verified, err := VerifyAccessToken("", "")
	t.Logf("%+v, %+v\n", verified, err)
	if !verified {
		t.Fail()
	}

	if err != nil {
		t.Fail()
	}
}

func TestVerifyRefreshToken(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	t.Logf("%+v\n", TerminalId)
	t.Logf("%+v\n", Token)
	verified, err := VerifyRefreshToken(Token.RefreshToken, TerminalId)
	// verified, err := VerifyRefreshToken("", "")
	t.Logf("%+v, %+v\n", verified, err)
	if !verified {
		t.Fail()
	}

	if err != nil {
		t.Fail()
	}
}
