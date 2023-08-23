package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

// FCMDB stores the users, their FCM identifiers, pubkeys, and their subscriptions
type FCMDB struct {
	db *sql.DB
}

type User struct {
	user_id string
	token   string
}

type PublicKey struct {
	user_id   string
	algorithm int
	is_auth   int
	key_bytes []byte
}

type Subscription struct {
	url              string
	subscribed_users []string
}

func NewFCMDB(DB *sql.DB) *FCMDB {
	fcmdb := &FCMDB{
		db: DB,
	}
	return fcmdb
}

func (DB *FCMDB) InitializeTables() error {
	query := `
        CREATE TABLE Users (
			user_id TEXT NOT NULL,
 			token TEXT NOT NULL,
 			PRIMARY KEY(user_id)
        );
		
		CREATE TABLE PublicKeys (
			user_id TEXT NOT NULL,
			algorithm INTEGER NOT NULL,
			is_auth INTEGER NOT NULL,
			key_bytes BLOB NOT NULL,
			PRIMARY KEY(user_id)
		);

		CREATE TABLE Subscriptions (
			url TEXT NOT NULL,
 			subscribed_users TEXT NOT NULL,
 			PRIMARY KEY(url)
        );

    `

	_, err := DB.db.Exec(query)
	if err != nil {
		fmt.Println("Failed to create table:", err)
		return err
	}
	return nil
}

func (DB *FCMDB) GetUser(user_id string) *User {
	rows, err := DB.db.Query(`
    SELECT user_id, token FROM Users WHERE user_id = ?;
    `, user_id)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var user_id string
		var token string
		if err := rows.Scan(&user_id, &token); err != nil {
			fmt.Println(err)
			return nil
		}
		return &User{
			user_id: user_id,
			token:   token,
		}
	}
	return nil
}

func (DB *FCMDB) GetUserTokens(user_ids []string) []string {

	var user_tokens []string

	for _, user_id := range user_ids {
		rows, err := DB.db.Query(`
    	SELECT user_id, token FROM Users WHERE user_id = ?;
    	`, user_id)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		defer rows.Close()

		for rows.Next() {
			var user_id string
			var token string
			if err := rows.Scan(&user_id, &token); err != nil {
				fmt.Println(err)
				return nil
			}
			user_tokens = append(user_tokens, token)
		}
	}

	return user_tokens
}

func (DB *FCMDB) GetPublicKey(user_id string) *PublicKey {
	rows, err := DB.db.Query(`
    SELECT user_id, algorithm, is_auth, key_bytes FROM PublicKeys WHERE user_id = ?;
    `, user_id)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var user_id string
		var algo int
		var is_auth int
		var key_bytes []byte
		if err := rows.Scan(&user_id, &algo, &is_auth, &key_bytes); err != nil {
			fmt.Println(err)
			return nil
		}
		return &PublicKey{
			user_id:   user_id,
			algorithm: algo,
			is_auth:   is_auth,
			key_bytes: key_bytes,
		}
	}
	return nil
}

func (DB *FCMDB) GetSubscription(url string) *Subscription {

	rows, err := DB.db.Query(`
    SELECT url, subscribed_users FROM Subscriptions WHERE url = ?;
    `, url)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		var usersT string
		if err := rows.Scan(&url, &usersT); err != nil {
			fmt.Println(err)
			return nil
		}
		users := strings.Split(usersT, ",")
		return &Subscription{
			url:              url,
			subscribed_users: users,
		}
	}
	return nil
}

func (DB *FCMDB) GetSubscriptions() []*Subscription {

	rows, err := DB.db.Query(`
    SELECT url, subscribed_users FROM Subscriptions;
    `)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	var subscriptions []*Subscription

	for rows.Next() {
		var url string
		var usersT string
		if err := rows.Scan(&url, &usersT); err != nil {
			fmt.Println(err)
			return nil
		}
		users := strings.Split(usersT, ",")
		subscriptions = append(subscriptions, &Subscription{
			url:              url,
			subscribed_users: users,
		})
	}

	// Check for any errors that occurred during the iteration.
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return nil
	}
	return subscriptions
}

func (DB *FCMDB) UpdateUser(user_id, token string) {
	stmt, err := DB.db.Prepare(`
    INSERT INTO Users (user_id, token) VALUES (?, ?);
    `)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()

	_, err = DB.db.Exec(`
    DELETE FROM Users WHERE user_id = ?;
    `, user_id)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = stmt.Exec(user_id, token)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// TODO: Support "algo" and "auth" options
func (DB *FCMDB) UpdateKey(user_id string, keybytes []byte) {
	stmt, err := DB.db.Prepare(`
    INSERT INTO PublicKeys (user_id, algorithm, is_auth, key_bytes) VALUES (?, ?, ?, ?);
    `)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()

	_, err = DB.db.Exec(`
    DELETE FROM PublicKeys WHERE user_id = ?;
    `, user_id)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = stmt.Exec(user_id, 0, 0, keybytes)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (DB *FCMDB) UpdateSubscription(user_id, url string) {
	stmt, err := DB.db.Prepare(`
    INSERT INTO Subscriptions (url, subscribed_users) VALUES (?, ?);
    `)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()

	var users = []string{user_id}

	Subscription := DB.GetSubscription(url)
	if Subscription != nil {
		users = append(Subscription.subscribed_users, user_id) // TODO: check if user_id is valid
		_, err = DB.db.Exec(`
    	DELETE FROM Subscriptions WHERE url = ?;
    	`, url)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	users = removeDuplicates(users)

	usersT := strings.Join(users, ",")

	_, err = stmt.Exec(url, usersT)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (DB *FCMDB) PrintUsers() {

	rows, err := DB.db.Query(`
    SELECT user_id, token FROM Users;
    `)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	// Print the results.
	for rows.Next() {
		var user_id string
		var token string
		if err := rows.Scan(&user_id, &token); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(user_id, token)
	}

	// Check for any errors that occurred during the iteration.
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return
	}
}

func (DB *FCMDB) PrintKeys() {

	rows, err := DB.db.Query(`
    SELECT user_id, key_bytes FROM PublicKeys;
    `)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	// Print the results.
	for rows.Next() {
		var user_id string
		var key []byte
		if err := rows.Scan(&user_id, &key); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(user_id, string(key))
	}

	// Check for any errors that occurred during the iteration.
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return
	}
}

func (DB *FCMDB) PrintSubscriptions() {

	rows, err := DB.db.Query(`
    SELECT url, subscribed_users FROM Subscriptions;
    `)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	// Print the results.
	for rows.Next() {
		var url string
		var usersT string
		if err := rows.Scan(&url, &usersT); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(url, usersT)
	}

	// Check for any errors that occurred during the iteration.
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return
	}
}

func removeDuplicates(strs []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, s := range strs {
		if !seen[s] {
			result = append(result, s)
			seen[s] = true
		}
	}
	return result
}
