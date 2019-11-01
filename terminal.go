package terminal

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	"path/filepath"

	// "github.com/gin-gonic/gin"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	// "strconv"
	"time"

	"github.com/google/uuid"
	usr "github.com/septianw/jas-user"
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
	GrantType    string `form:"grant_type" binding:"required"`
	RefreshToken string `form:"refresh_token"`
	Username     string `form:"username"`
	Password     string `form:"password"`
	ClientId     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	Scope        string `form:"scope"`
}

type ClientCredential struct {
	ClientId     string
	ClientSecret string
	ClientName   string
}

type MarkRecordAsObsoleteFn func() (err error)

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

func verify(sbTerm strings.Builder, obsolete MarkRecordAsObsoleteFn) (verified bool, err error) {
	var count uint
	verified = false

	log.Println(sbTerm.String())
	rows, err := Query(sbTerm.String())
	log.Println(sbTerm.String())
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		rows.Scan(&count)
	}

	// obsolete()

	log.Printf("Record Found: %d\nverified: %+v", count, (count > 0))
	if count > 0 {
		err = obsolete()
		log.Println(err)
		if err == nil {
			verified = true
		}
	}

	return
}

// FIXME ini harusnya diperlakukan yang sama seperti verify password.
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

	return verify(sbTerm, func() error { return nil })
}

func VerifyTerminal(terminalId string) (verified bool, err error) {
	var sbTerm strings.Builder

	_, err = sbTerm.WriteString(fmt.Sprintf(`SELECT COUNT(*) AS count
	FROM terminal WHERE terminalid = '%s'`, terminalId))
	if err != nil {
		return
	}

	return verify(sbTerm, func() error { return nil })
}

func VerifyAccessToken(accessToken, terminalId string) (verified bool, err error) {
	var sbTerm strings.Builder
	// var count uint64
	verified = false

	// get valid refresh token associated with
	_, err = sbTerm.WriteString(fmt.Sprintf(`select COUNT(*) as count
from tokenusage as tu
join accesstoken as act on act.tokenusage_usageid = tu.usageid
where tu.terminal_terminalid = '%s' and tu.expired = 0 AND act.token = '%s'`,
		terminalId, accessToken))
	if err != nil {
		return
	}

	fmt.Println(sbTerm.String())
	// sbTerm.Reset()

	// FIXME
	return verify(sbTerm, func() error {
		debug.PrintStack()
		log.Println(terminalId, accessToken)
		var timeout, usageid, delta int64
		var created, upstring string
		var updated sql.NullString
		var expired uint8 = 0
		t := time.Now()

		sbTerm.Reset()
		sbTerm.WriteString(fmt.Sprintf(`SELECT tu.usageid, act.timeout, tu.created, tu.updated
		FROM tokenusage AS tu
		JOIN accesstoken AS act ON act.tokenusage_usageid = tu.usageid
		WHERE tu.terminal_terminalid = '%s'
		AND act.token = '%s'`, terminalId, accessToken))
		log.Println(sbTerm.String())
		rows, err := Query(sbTerm.String())
		if err != nil {
			return err
		}

		for rows.Next() {
			err := rows.Scan(&usageid, &timeout, &created, &updated)
			if updated.Valid {
				upstring = updated.String
			}
			common.ErrHandler(err)
		}
		rows.Close()
		sbTerm.Reset()

		tCreated, err := time.Parse("2006-01-02T15:04:05Z", created)
		if err != nil {
			return err
		}

		upstring = t.Format("2006-01-02 15:04:05")

		// ini masih pakai unix timestamp
		// asumsinya timeout itu far future.
		log.Println(time.Unix(timeout, 0).Format("2006"))
		if strings.Compare(time.Unix(timeout, 0).Format("2006"), "2019") == 0 {
			delta = timeout - t.Unix()
			if delta < 0 {
				expired = 1
			}
		} else {
			age := t.Sub(tCreated)
			tokenAge := tCreated.Add(time.Duration(timeout) * time.Second)
			log.Printf("\nage: %+v, tokenAge: %+v, created: %+v, timeout: %+v, now: %+v", age.Seconds(), tokenAge, tCreated.Unix(), timeout, t.Unix())
			deltaDur := tokenAge.Sub(time.Now())
			log.Printf("\ntokenage - t: %+v\n", deltaDur.Hours())
			if int64(deltaDur.Seconds()) > 86400 {
				delta = 86400
			} else {
				delta = int64(deltaDur.Seconds())
			}
			if delta <= 0 {
				expired = 1
			}
		}
		log.Println(delta)

		sbTerm.WriteString(fmt.Sprintf("update tokenusage as tu, accesstoken as act"+
			" set act.timeout = %d, tu.updated = '%s', tu.expired = %d"+
			" where tu.usageid = act.tokenusage_usageid and tu.terminal_terminalid = '%s'"+
			" and act.token = '%s'", delta, upstring, expired, terminalId, accessToken))
		log.Println(sbTerm.String())

		_, err = Exec(sbTerm.String())
		if err != nil {
			return err
		}

		return nil
	})
}

func VerifyRefreshToken(refreshToken, terminalId string) (verified bool, err error) {
	var sbTerm strings.Builder
	// var count uint64
	verified = false

	// get valid refresh token associated with
	_, err = sbTerm.WriteString(fmt.Sprintf(`SELECT COUNT(*) AS count
FROM tokenusage AS tu
JOIN refreshtoken AS rt ON rt.tokenusage_usageid = tu.usageid
WHERE tu.terminal_terminalid = '%s' AND tu.expired = 0 AND rt.token = '%s'`,
		terminalId, refreshToken))
	if err != nil {
		return
	}

	// FIXME
	return verify(sbTerm, func() error {
		debug.PrintStack()
		log.Println(terminalId, refreshToken)
		var timeout, usageid, delta int64
		var created, upstring string
		var updated sql.NullString
		var expired uint8 = 0
		t := time.Now()

		sbTerm.Reset()
		sbTerm.WriteString(fmt.Sprintf(`SELECT tu.usageid, rt.timeout, tu.created, tu.updated
		FROM tokenusage AS tu
		JOIN refreshtoken AS rt ON rt.tokenusage_usageid = tu.usageid
		WHERE tu.terminal_terminalid = '%s'
		AND rt.token = '%s'`, terminalId, refreshToken))
		log.Println(sbTerm.String())
		rows, err := Query(sbTerm.String())
		if err != nil {
			return err
		}

		for rows.Next() {
			err := rows.Scan(&usageid, &timeout, &created, &updated)
			if updated.Valid {
				upstring = updated.String
			}
			common.ErrHandler(err)
		}
		rows.Close()
		sbTerm.Reset()

		tCreated, err := time.Parse("2006-01-02T15:04:05Z", created)
		if err != nil {
			return err
		}

		upstring = t.Format("2006-01-02 15:04:05")

		// ini masih pakai unix timestamp
		// asumsinya timeout itu far future.
		log.Println(time.Unix(timeout, 0).Format("2006"))
		if strings.Compare(time.Unix(timeout, 0).Format("2006"), "2019") == 0 {
			delta = timeout - t.Unix()
			if delta < 0 {
				expired = 1
			}
		} else {
			age := t.Sub(tCreated)
			tokenAge := tCreated.Add(time.Duration(timeout) * time.Second)
			log.Printf("\nage: %+v, tokenAge: %+v, created: %+v, timeout: %+v, now: %+v", age.Seconds(), tokenAge, tCreated.Unix(), timeout, t.Unix())
			deltaDur := tokenAge.Sub(time.Now())
			log.Printf("\ntokenage - t: %+v\n", deltaDur.Hours())
			if int64(deltaDur.Seconds()) > 259200 {
				delta = 259200
			} else {
				delta = int64(deltaDur.Seconds())
			}
			if delta <= 0 {
				expired = 1
			}
		}
		log.Println(delta)

		sbTerm.WriteString(fmt.Sprintf("update tokenusage as tu, refreshtoken as rt"+
			" set rt.timeout = %d, tu.updated = '%s', tu.expired = %d"+
			" where tu.usageid = rt.tokenusage_usageid and tu.terminal_terminalid = '%s'"+
			" and rt.token = '%s'", delta, upstring, expired, terminalId, refreshToken))
		log.Println(sbTerm.String())

		_, err = Exec(sbTerm.String())
		if err != nil {
			return err
		}

		return nil
	})
}

// Set token expired.
func SetExpired(token, tokenType string) (err error) {
	var sbToken strings.Builder

	/*
		update accesstoken as t, tokenusage as tu
		set tu.expired = 1, tu.updated = '1982-02-06 22:54:43'
		where t.tokenusage_usageid = tu.usageid
			and t.token = '01d69673fc28382ded95d9da8c853aacc914fd97';
	*/

	switch tokenType {
	case "refresh_token":
		sbToken.WriteString("UPDATE accesstoken AS t, tokenusage AS tu")
		break
	case "access_token":
		sbToken.WriteString("UPDATE refreshtoken AS t, tokenusage AS tu")
		break
	default:
		err = errors.New("Sorry, only access_token of refresh_token that can be set expired.")
		return
	}

	sbToken.WriteString(fmt.Sprintf(" SET tu.expired = 1, tu.updated = '%s'"+
		" WHERE t.tokenusage_usageid = tu.usageid AND t.token = '%s'",
		time.Now().Format("2006-01-02 15:04:05"), token))

	log.Printf(sbToken.String())

	_, err = Exec(sbToken.String())
	if err != nil {
		return err
	}

	return
}

func GenerateTokens(terminalId string, grant Grant) (response TokenResponse, err error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var sbAkey, sbEkey, sbQkey strings.Builder
	var tokenUsageId int64
	var queryAccess, queryRefresh string

	accessTokenExpired := (60 * 60 * 24)
	refreshTokenExpired := (60 * 60 * 24 * 3)

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
	// kalau refresh token dan tanpa username
	if (strings.Compare(grant.Username, "") == 0) &&
		(strings.Compare(grant.GrantType, "refresh_token") == 0) {
		sbQkey.WriteString(fmt.Sprintf(`SELECT tu.user_uname FROM tokenusage as tu
JOIN refreshtoken AS rt ON rt.tokenusage_usageid = tu.usageid
WHERE rt.token = '%s'`, grant.RefreshToken))

		rows, err := Query(sbQkey.String())
		if err != nil {
			return response, err
		}

		for rows.Next() {
			rows.Scan(&grant.Username)
		}
		rows.Close()
		sbQkey.Reset()

		sbQkey.WriteString(fmt.Sprintf(`UPDATE tokenusage AS tu, refreshtoken AS rt SET tu.expired = 1
WHERE rt.tokenusage_usageid = tu.usageid AND rt.token = '%s'`, grant.RefreshToken))
		log.Printf("\n%+v\n", sbQkey.String())
		_, err = Exec(sbQkey.String())
		if err != nil {
			return response, err
		}
		sbQkey.Reset()
	}

	if (strings.Compare(grant.Username, "") == 0) &&
		(strings.Compare(grant.GrantType, "client_credentials") == 0) {
		sbQkey.Reset()
		// INSERT INTO credentialusage (clientcredential_clientid, terminal_terminalid, `date`)
		// VALUES ('%s', '%s', '%s');

		sbQkey.WriteString(fmt.Sprintf("INSERT INTO credentialusage (clientcredential_clientid, terminal_terminalid, `date`)"+
			"VALUES ('%s', '%s', '%s')", grant.ClientId, terminalId, time.Now().Format("2006-01-02 15:04:05")))
		result, err := Exec(sbQkey.String())
		if err != nil {
			return response, err
		}

		credentialusageUsageid, err := result.LastInsertId()
		if err != nil {
			return response, err
		}
		sbQkey.Reset()

		sbQkey.WriteString(fmt.Sprintf(`INSERT INTO accesstoken
		(token, timeout, used, credentialusage_credentialusageid)
		VALUES ('%s', %d, 1, %d)`, accessToken, accessTokenExpired, credentialusageUsageid))
		queryAccess = sbQkey.String()
		sbQkey.Reset()

		// FIXME INI MENGHASILKAN QUERY KOSONG.
		sbQkey.WriteString(fmt.Sprintf(`INSERT INTO refreshtoken
		(token, timeout, used, credentialusage_credentialusageid)
		VALUES ('%s', %d, 1, %d)`, refreshToken, refreshTokenExpired, credentialusageUsageid))
		queryRefresh = sbQkey.String()
		// insert into credentialusage (clientcredential_clientid, terminal_terminalid, date)
		// values ('%s', '%s', '%s')
	} else {
		users, err := usr.FindUser(usr.UserIn{Uname: grant.Username})
		log.Println(grant.Username)
		log.Println(err)
		log.Printf("\n%+v\n", users)
		log.Println(accessToken, refreshToken)

		sbQkey.WriteString(fmt.Sprintf(`INSERT INTO tokenusage
		(user_uid, user_uname, terminal_terminalid, created)
		VALUES (%d, '%s', '%s', '%s')`,
			users[0].Uid, users[0].Uname, terminalId, time.Now().Format("2006-01-02 15:04:05")))
		qTokenUsage := sbQkey.String()
		fmt.Println(qTokenUsage)
		result, err := Exec(qTokenUsage)
		if err != nil {
			return response, err
		}
		tokenUsageId, err = result.LastInsertId()
		if err != nil {
			return response, err
		}
		sbQkey.Reset()
		log.Println(tokenUsageId)

		sbQkey.WriteString(fmt.Sprintf(`INSERT INTO accesstoken
		(token, timeout, used, tokenusage_usageid)
		VALUES ('%s', %d, 1, %d)`, accessToken, accessTokenExpired, tokenUsageId))
		queryAccess = sbQkey.String()
		sbQkey.Reset()

		sbQkey.WriteString(fmt.Sprintf(`INSERT INTO refreshtoken
		(token, timeout, used, tokenusage_usageid)
		VALUES ('%s', %d, 1, %d)`, refreshToken, refreshTokenExpired, tokenUsageId))
		queryRefresh = sbQkey.String()

	}
	log.Println(queryAccess)
	_, err = Exec(queryAccess)
	if err != nil {
		return
	}

	log.Println(queryRefresh)
	_, err = Exec(queryRefresh)
	if err != nil {
		return
	}

	response.ExpiresIn = uint64(accessTokenExpired)
	response.TokenType = "bearer"
	response.AccessToken = accessToken
	response.RefreshToken = refreshToken

	return
}

func FetchToken(terminalId string, grant Grant) (response TokenResponse, err error) {
	// var t time.Time
	var sbQToken strings.Builder
	var responses []TokenResponse

	// t = time.Now()
	// year, month, day := t.Date()
	// RangeStart := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	// RangeEnd := time.Date(year, month, day, 23, 59, 59, 0, t.Location())

	// startString := RangeStart.Format("2006-01-02 15:04:05")
	// endString := RangeEnd.Format("2006-01-02 15:04:05")

	sbQToken.WriteString(fmt.Sprintf(`select
	act.token as access_token,
	rt.token as refresh_token,
	act.timeout as expires_in
from tokenusage as tu
join refreshtoken as rt on rt.tokenusage_usageid = tu.usageid
join accesstoken as act on act.tokenusage_usageid = tu.usageid
where tu.expired = 0
and tu.terminal_terminalid = '%s' and tu.user_uname = '%s'`,
		terminalId, grant.Username))

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

// FIXME ini harusnya diperlakukan sama seperti username dan password.
func InsertClientCredential(clientName string) (clientOut ClientCredential, err error) {
	var sbClient strings.Builder
	var rid, rsec []byte
	var clients []ClientCredential

	rid = make([]byte, 16)
	rsec = make([]byte, 32)

	_, err = rand.Read(rid)
	if err != nil {
		return
	}
	_, err = rand.Read(rsec)
	if err != nil {
		return
	}

	clients, err = GetClientCredentials(clientName)
	if err != nil {
		return
	}

	if len(clients) != 0 {
		log.Printf("\nclients: %+v\n", clients)
		err = errors.New("Client with the same name found, client name need to be unique")
		return
	}

	sbClient.WriteString(fmt.Sprintf(`INSERT INTO clientcredential (clientid, clientsecret, clientname, deleted)
	VALUES ('%s', '%s', '%s', 0)`, hex.EncodeToString(rid), hex.EncodeToString(rsec), clientName))

	_, err = Exec(sbClient.String())
	if err != nil {
		return
	}

	clients, err = GetClientCredentials(clientName)
	if err != nil {
		return
	}

	if len(clients) == 0 {
		err = errors.New("Insert client credential fail.")
		return
	}

	clientOut = clients[0]

	return
}

func GetClientCredentials(clientName string) (clientOuts []ClientCredential, err error) {
	var sbClient strings.Builder
	var clientOut ClientCredential
	var clName sql.NullString

	sbClient.WriteString(`SELECT clientid, clientsecret, clientname FROM clientcredential WHERE (deleted = 0 OR deleted IS NULL)`)

	if strings.Compare(clientName, "") != 0 {
		sbClient.WriteString(fmt.Sprintf(` AND clientname = '%s'`, clientName))
	}
	log.Println(sbClient.String())

	rows, err := Query(sbClient.String())
	if err != nil {
		return
	}

	for rows.Next() {
		err = rows.Scan(&clientOut.ClientId, &clientOut.ClientSecret, &clName)
		if err != nil {
			return
			break
		}
		if clName.Valid {
			clientOut.ClientName = clName.String
		}
		clientOuts = append(clientOuts, clientOut)
	}

	return
}

func DeleteClientCredentials(clientName string) (clientOuts []ClientCredential, err error) {
	var sbClient strings.Builder

	if strings.Compare(clientName, "") == 0 {
		err = errors.New("clientName cannot be empty.")
		return
	}

	clients, err := GetClientCredentials(clientName)
	if err != nil {
		return
	}

	if len(clients) > 1 {
		log.Printf("\ndelete target client: %+v\n", clients)
		err = errors.New("Duplicate client credential found, delete failed.")
		return
	}

	if len(clients) == 0 {
		err = errors.New("No client credential found, delete failed.")
		return
	}

	sbClient.WriteString(fmt.Sprintf(`UPDATE clientcredential SET deleted = 1 WHERE clientname = '%s'`, clientName))

	_, err = Exec(sbClient.String())
	if err != nil {
		return
	}

	if len(clients) == 1 {
		clientOuts = clients
	}

	return
}
