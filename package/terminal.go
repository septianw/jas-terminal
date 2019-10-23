package terminal

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"path/filepath"

	// "github.com/gin-gonic/gin"
	"crypto/sha256"
	"encoding/base64"

	// "strconv"
	"time"

	"github.com/google/uuid"
	usr "github.com/septianw/jas-user/package"
	"github.com/septianw/jas/common"
)

const VERSION = Version

type TerminalFull struct {
	TerminalId     string
	Name           string
	Deleted        uint8
	Location_locid uint64
}

type TerminalOut struct {
	TerminalId string  `json:"terminalid" binding:"required"`
	Name       string  `json:"name" binding:"required"`
	Location   string  `json:"location" binding:"required"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

type TerminalIn struct {
	TerminalId string `json:"terminalid" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Location   string `json:"location" binding:"required"`
}

type TerminalUpdate struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type Grant struct {
	GrantType    string `form:"grant_type"`
	Username     string `form:"username"`
	Password     string `form:"password"`
	ClientId     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	Scope        string `form:"scope"`
}

/*
{
  "access_token":"MTQ0NjJkZmQ5OTM2NDE1ZTZjNGZmZjI3",
  "token_type":"bearer",
  "expires_in":3600,
  "refresh_token":"IwOGYzYTlmM2YxOTQ5MGE3YmNmMDFkNTVk",
  "scope":"create"
}
*/
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    uint64 `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

/*
SELECT
	t.terminalid as terminalid,
    t.name as `name`,
    l.name as `location_name`,
    l.latitude latitude,
    l.longitude longitude
FROM terminal AS t
JOIN location AS l
  ON t.location_locid = l.locid
WHERE t.deleted = 0;

*/

func getdbobj() (db *sql.DB, err error) {
	rt := common.ReadRuntime()
	dbs := common.LoadDatabase(filepath.Join(rt.Libloc, "database.so"), rt.Dbconf)
	db, err = dbs.OpenDb(rt.Dbconf)
	return
}

func Query(q string) (*sql.Rows, error) {
	db, err := getdbobj()
	common.ErrHandler(err)
	defer db.Close()

	return db.Query(q)
}

func Exec(q string) (sql.Result, error) {
	db, err := getdbobj()
	common.ErrHandler(err)
	defer db.Close()

	return db.Exec(q)
}

// get query
/*
	sbTerm.WriteString(fmt.Sprintf(`SELECT t.terminalid as terminalid,
    t.name as ` + "`name`" + `, l.name as ` + "`location_name`" + `, l.latitude latitude,
    l.longitude longitude FROM terminal AS t JOIN location AS l
	ON t.location_locid = l.locid WHERE t.deleted = 0`))
*/
// insert query
/*
	sbTerm.WriteString(fmt.Sprintf(`INSERT INTO `+"`terminal`"+` (terminalid, `+
		"`name`"+`, deleted, location_locid) VALUES ('%s', '%s', 0, (
	SELECT locid FROM location WHERE `+"`name`"+` = '%s'))`,
		termin.TerminalId, termin.Name, termin.Location))
*/
// update query
/*
	sbTerm.WriteString(fmt.Sprintf(`UPDATE terminal SET `+"`name`"+` = '%s',
	location_locid = (SELECT locid FROM location WHERE ` + "`name`" + ` = '%s')
	WHERE terminalid = '%s'`,
		termin.TerminalId, termin.Name, termin.Location))
*/
// delete query
/*
	sbTerm.WriteString(fmt.Sprintf(`UPDATE terminal SET deleted = 1 WHERE terminalid = '%s'`,
		termin.TerminalId))
*/

func InsertTerminal(termin TerminalIn) (termout TerminalOut, err error) {
	var sbTerm strings.Builder

	_, err = sbTerm.WriteString(fmt.Sprintf(`INSERT INTO `+"`terminal`"+` (terminalid, `+
		"`name`"+`, deleted, location_locid) VALUES ('%s', '%s', 0, (
	SELECT locid FROM location WHERE `+"`name`"+` = '%s'))`,
		termin.TerminalId, termin.Name, termin.Location))
	if err != nil {
		return
	}
	log.Println(sbTerm.String())

	_, err = Exec(sbTerm.String())
	if err != nil {
		return
	}

	terminals, err := GetTerminal(termin.TerminalId, 0, 0)
	if err != nil {
		return
	}
	if len(terminals) == 0 {
		err = errors.New("Terminal not found.")
		return
	}
	termout = terminals[0]

	return
}

func GetTerminal(id string, limit, offset int64) (terminals []TerminalOut, err error) {
	var sbTerm strings.Builder
	var terminal TerminalOut

	_, err = sbTerm.WriteString(fmt.Sprintf(`SELECT t.terminalid as terminalid,
    t.name as ` + "`name`" + `, l.name as ` + "`location_name`" + `, l.latitude latitude,
    l.longitude longitude FROM terminal AS t JOIN location AS l
	ON t.location_locid = l.locid WHERE t.deleted = 0`))
	if err != nil {
		return
	}

	if id == "" {
		if limit == 0 {
			_, err = sbTerm.WriteString(fmt.Sprintf(" LIMIT 10 OFFSET 0"))
			if err != nil {
				return
			}
		} else {
			_, err = sbTerm.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset))
			if err != nil {
				return
			}
		}
	} else {
		_, err = sbTerm.WriteString(fmt.Sprintf(" AND terminalid = '%s'", id))
		if err != nil {
			return
		}
	}

	log.Println(sbTerm.String())

	rows, err := Query(sbTerm.String())
	if err != nil {
		return
	}

	for rows.Next() {
		err = rows.Scan(&terminal.TerminalId, &terminal.Name, &terminal.Location, &terminal.Latitude, &terminal.Longitude)
		if err != nil {
			return
		}
		terminals = append(terminals, terminal)
	}

	if len(terminals) == 0 {
		err = errors.New("Terminal not found.")
		return
	}

	return
}

func UpdateTerminal(id string, termin TerminalUpdate) (terminal TerminalOut, err error) {
	var sbTerm strings.Builder

	_, err = sbTerm.WriteString(fmt.Sprintf(`UPDATE terminal SET `+"`name`"+` = '%s',
	location_locid = (SELECT locid FROM location WHERE `+"`name`"+` = '%s')
	WHERE terminalid = '%s'`,
		termin.Name, termin.Location, id))
	if err != nil {
		return
	}

	result, err := Exec(sbTerm.String())
	if err != nil {
		return
	}
	raff, err := result.RowsAffected()
	if err != nil {
		return
	}
	log.Printf("Updated terminal(s): %+v", raff)

	terminals, err := GetTerminal(id, 0, 0)
	if err != nil {
		return
	}
	if len(terminals) == 0 {
		err = errors.New("Terminal not found.")
		return
	}
	terminal = terminals[0]

	return
}

func DeleteTerminal(id string) (terminal TerminalOut, err error) {
	var sbTerm strings.Builder

	_, err = sbTerm.WriteString(fmt.Sprintf(`UPDATE terminal SET deleted = 1 WHERE terminalid = '%s'`,
		id))
	if err != nil {
		return
	}

	terminals, err := GetTerminal(id, 0, 0)
	if err != nil {
		return
	}

	if len(terminals) == 0 {
		err = errors.New("Terminal not found.")
		return
	}

	result, err := Exec(sbTerm.String())
	if err != nil {
		return
	}
	raff, err := result.RowsAffected()
	if err != nil {
		return
	}
	log.Printf("Updated terminal(s): %+v", raff)

	terminal = terminals[0]

	return
}

func verify(sbTerm strings.Builder) (verified bool, err error) {
	var count uint
	verified = false

	log.Println(sbTerm.String())
	rows, err := Query(sbTerm.String())
	if err != nil {
		return
	}
	for rows.Next() {
		rows.Scan(&count)
	}

	log.Printf("Record Found: %d\nverified: %+v", count, (count > 0))
	if count > 0 {
		verified = true
	}

	return
}

func VerifyClients(clientId, clientSecret string) (verified bool, err error) {
	var sbTerm strings.Builder

	_, err = sbTerm.WriteString(fmt.Sprintf(`SELECT COUNT(*) AS count
	FROM clientcredential WHERE clientid = '%s'`, clientId))
	if err != nil {
		return
	}

	if strings.Compare(clientSecret, "") != 0 {
		_, err = sbTerm.WriteString(fmt.Sprintf(` AND clientsecret = '%s'`, clientSecret))
		if err != nil {
			return
		}
	}

	return verify(sbTerm)
}

func VerifyTerminal(terminalId string) (verified bool, err error) {
	var sbTerm strings.Builder

	_, err = sbTerm.WriteString(fmt.Sprintf(`SELECT COUNT(*) AS count
	FROM terminal WHERE terminalid = '%s'`, terminalId))
	if err != nil {
		return
	}

	return verify(sbTerm)
}

func VerifyAccessToken(accessToken string) (verified bool, err error) {
	var sbTerm strings.Builder

	_, err = sbTerm.WriteString(fmt.Sprintf(`SELECT COUNT(*) count
	FROM accesstoken WHERE used = 0 AND token = '%s'`, accessToken))
	if err != nil {
		return
	}

	return verify(sbTerm)
}

func GenerateTokens(terminalId string, grant Grant) (response TokenResponse, err error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var sbAkey, sbEkey, sbQkey strings.Builder
	var tokenUsageId int64

	accessTokenExpired := (60 * 60 * 24) + time.Now().Unix()
	refreshTokenExpired := (60 * 60 * 24 * 3) + time.Now().Unix()

	accessExpired := fmt.Sprintf("%d", accessTokenExpired)
	refreshExpired := fmt.Sprintf("%d", refreshTokenExpired)

	sbAkey.WriteString(accessExpired)
	sbAkey.WriteString("$")
	ka := sha256.Sum256([]byte(uuid.New().String()))
	sbAkey.WriteString(string(ka[:]))

	sbEkey.WriteString(refreshExpired)
	sbEkey.WriteString("$")
	ke := sha256.Sum256([]byte(uuid.New().String()))
	sbEkey.WriteString(string(ke[:]))

	accessToken := base64.StdEncoding.EncodeToString([]byte(sbAkey.String()))
	refreshToken := base64.StdEncoding.EncodeToString([]byte(sbEkey.String()))

	/*users*/
	// _, err = usr.FindUser(usr.UserIn{Uname: grant.Username})
	users, err := usr.FindUser(usr.UserIn{Uname: grant.Username})
	log.Println(err)
	log.Println(accessToken, refreshToken)

	sbQkey.WriteString(fmt.Sprintf(`INSERT INTO tokenusage
	(user_uid, user_uname, terminal_terminalid, date)
	VALUES (%d, '%s', '%s', '%s')`,
		users[0].Uid, users[0].Uname, terminalId, time.Now().Format("2006-01-02 15:04:05")))
	qTokenUsage := sbQkey.String()
	fmt.Println(qTokenUsage)
	result, err := Exec(qTokenUsage)
	if err != nil {
		return
	}
	tokenUsageId, err = result.LastInsertId()
	if err != nil {
		return
	}
	sbQkey.Reset()
	log.Println(tokenUsageId)

	sbQkey.WriteString(fmt.Sprintf(`INSERT INTO accesstoken
	(token, timeout, used, tokenusage_usageid)
	VALUES ('%s', %d, 0, %d)`, accessToken, accessTokenExpired, tokenUsageId))
	qAccessToken := sbQkey.String()
	fmt.Println(qAccessToken)
	_, err = Exec(qAccessToken)
	if err != nil {
		return
	}
	sbQkey.Reset()

	sbQkey.WriteString(fmt.Sprintf(`INSERT INTO refreshtoken
	(token, timeout, used, tokenusage_usageid)
	VALUES ('%s', %d, 0, %d)`, refreshToken, refreshTokenExpired, tokenUsageId))
	qRefreshToken := sbQkey.String()
	fmt.Println(qRefreshToken)
	_, err = Exec(qRefreshToken)
	if err != nil {
		return
	}

	response.ExpiresIn = 60 * 60 * 24
	response.TokenType = "bearer"
	response.AccessToken = accessToken
	response.RefreshToken = refreshToken

	return
}

func FetchToken(terminalId string, grant Grant) (response TokenResponse, err error) {
	var t time.Time
	var sbQToken strings.Builder
	var responses []TokenResponse

	t = time.Now()
	year, month, day := t.Date()
	RangeStart := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	RangeEnd := time.Date(year, month, day, 23, 59, 59, 0, t.Location())

	startString := RangeStart.Format("2006-01-02 15:04:05")
	endString := RangeEnd.Format("2006-01-02 15:04:05")

	sbQToken.WriteString(fmt.Sprintf(`select
	act.token as access_token,
	rt.token as refresh_token,
	act.timeout as expires_in
from tokenusage as tu
join refreshtoken as rt on rt.tokenusage_usageid = tu.usageid
join accesstoken as act on act.tokenusage_usageid = tu.usageid
where `+"`date`"+` between '%s' and '%s'
and tu.terminal_terminalid = '%s' and tu.user_uname = '%s'`,
		startString, endString, terminalId, grant.Username))

	fmt.Println(sbQToken.String())

	rows, err := Query(sbQToken.String())
	if err != nil {
		return
	}
	for rows.Next() {
		var sr TokenResponse
		err = rows.Scan(&sr.AccessToken, &sr.RefreshToken, &sr.ExpiresIn)
		if err != nil {
			return
		}

		sr.TokenType = "bearer"

		responses = append(responses, sr)
	}

	if len(responses) > 0 {
		response = responses[0]
	}

	return
}

/*
return :
{
  "access_token":"MTQ0NjJkZmQ5OTM2NDE1ZTZjNGZmZjI3",
  "token_type":"bearer",
  "expires_in":3600,
  "refresh_token":"IwOGYzYTlmM2YxOTQ5MGE3YmNmMDFkNTVk",
  "scope":"create"
}
*/
func IssueTokens(terminalId string, grant Grant) (response TokenResponse, err error) {
	res, err := FetchToken(terminalId, grant)
	if err != nil {
		return
	}

	res1 := &res
	if (TokenResponse{}) == *res1 {
		resg, err := GenerateTokens(terminalId, grant)
		if err != nil {
			return response, err
		}
		response = resg
	} else {
		response = res
	}

	return
}
