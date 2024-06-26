package persistence

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

// Contains parameters to be used in pgsql connection string
type PgType struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

// Configuration struct
type ConfigType struct {
	Origin string
	Pg     PgType
}

// Represents row from database table "accounts"
type Account struct {
	id        int
	name      string
	pass_hash string
}

// Represents command pulled from database table "commands"
type QueryCommand struct {
	text string
}

var Config ConfigType

// Connection string
var psqlInfo string

// Sets values for Config and psqlInfo
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

	psqlInfo = "host=" + Config.Pg.Host +
		" port=" + strconv.Itoa(Config.Pg.Port) +
		" user=" + Config.Pg.User +
		" password=" + Config.Pg.Password +
		" dbname=" + Config.Pg.Dbname +
		" sslmode=disable"
}

// Authenticates user in database table accounts
func Authenticate(username string, password string) (bool, error) {
	hash := sha256.Sum256([]byte(password))

	hashPass := hex.EncodeToString(hash[:])

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

	if len(accounts) != 0 && accounts[0].pass_hash == hashPass {
		return true, nil
	}

	return false, errors.New("invalid password")
}

// Gets commands from database that are assigned to specified user
func QueryCommands(username string) ([]string, error) {
	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		return nil, err
	}

	defer db.Close()

	var command_list []string

	rows_all, err := db.Query(`
		select commands.text from accounts_commands 
		join commands on commands.id = accounts_commands.command_id
		join accounts on accounts.id = accounts_commands.account_id
		where accounts.name = 'all'
		order by commands.id desc
	`)

	if err != nil {
		return nil, err
	}

	defer rows_all.Close()

	for rows_all.Next() {
		var c QueryCommand

		err := rows_all.Scan(&c.text)

		if err != nil {
			continue
		}
		command_list = append(command_list, c.text)
	}

	rows_user, err := db.Query(fmt.Sprintf(`
		select commands.text from accounts_commands 
		join commands on commands.id = accounts_commands.command_id
		join accounts on accounts.id = accounts_commands.account_id
		where accounts.name = '%s' order by commands.id desc
	`, username))

	if err != nil {
		return nil, err
	}

	defer rows_user.Close()

	for rows_user.Next() {
		var c QueryCommand

		err := rows_user.Scan(&c.text)

		if err != nil {
			continue
		}
		command_list = append(command_list, c.text)
	}

	return command_list, nil
}

// Saves new command for specified user
func SaveCommand(username, command string) error {
	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		return err
	}

	defer db.Close()

	queryString := `
		-- Insert the command into the commands table
		INSERT INTO commands (text) VALUES ('%s');
		
		-- Insert the account command into the accounts_commands table
		INSERT INTO accounts_commands (account_id, command_id)
		VALUES (
			(SELECT id FROM accounts WHERE name = '%s'),
			LASTVAL()
		);
	`

	result, err := db.Exec(fmt.Sprintf(queryString, command, username))

	affected, _ := result.RowsAffected()

	fmt.Fprint(io.Discard, affected)

	if err != nil {
		return err
	}

	return nil
}
