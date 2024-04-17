package persistence

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

type PgType struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

type ConfigType struct {
	Origin string
	Pg     PgType
}

type Account struct {
	id        int
	name      string
	pass_hash string
}

var Config ConfigType

func Init() {
	fileContents, err := os.ReadFile("./conf.d/persistence.json")

	if err != nil {
		panic("no persistence.json found!")
	}

	// Unmarshal JSON into the variable
	err = json.Unmarshal(fileContents, &Config)

	if err != nil {
		panic("invalid configuration provided in persistence.json!")
	}
}

func Authenticate(username string, password string) (bool, error) {
	hash := sha256.Sum256([]byte(password))

	hashPass := hex.EncodeToString(hash[:])

	psqlInfo := "host=" + Config.Pg.Host +
		" port=" + strconv.Itoa(Config.Pg.Port) +
		" user=" + Config.Pg.User +
		" password=" + Config.Pg.Password +
		" dbname=" + Config.Pg.Dbname +
		" sslmode=disable"

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		return false, err
	}

	defer db.Close()

	rows, err := db.Query("select * from accounts where name = '" + username + "'")

	if err != nil {
		return false, err
	}

	defer rows.Close()

	accounts := []Account{}

	for rows.Next() {
		a := Account{}

		err := rows.Scan(&a.id, &a.name, &a.pass_hash)

		if err != nil {
			continue
		}
		accounts = append(accounts, a)
	}

	if accounts[0].pass_hash == hashPass {
		return true, nil
	}

	return false, errors.New("invalid password")
}
