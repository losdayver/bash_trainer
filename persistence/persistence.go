package persistence

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

type PgType struct {
	host     string
	port     int
	user     string
	password string
	dbname   string
}

type ConfigType struct {
	Origin string
	pg     PgType
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

	// psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	// 	host, port, user, password, "bash_trainer_cache")

	psqlInfo := "host=" + Config.pg.host +
		"port=" + strconv.Itoa(Config.pg.port) +
		"user=" + Config.pg.user +
		"password=" + Config.pg.password +
		"dbname=" + Config.pg.dbname +
		"sslmode=disable"

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		return false, err
	}

	defer db.Close()

	rows, err := db.Query("select * from accounts where name = '" + username + "'")

	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}

	defer rows.Close()

	accounts := []Account{}

	for rows.Next() {
		a := Account{}

		err := rows.Scan(&a.id, &a.name, &a.pass_hash)

		if err != nil {
			fmt.Println(err)
			continue
		}
		accounts = append(accounts, a)
	}

	if accounts[0].pass_hash == hashPass {
		return true, nil
	}

	return false, errors.New("invalid password")
}
