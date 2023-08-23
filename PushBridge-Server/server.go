package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/censoredplanet/orbot-android-push/PushBridge-server/fcmsender"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

func main() {
	var (
		credsFilename string
		dbPath        string
	)
	flag.StringVar(&credsFilename, "credentials", "", "Firebase service account private key json file, can be downloaded from Firebase console (needed)")
	flag.StringVar(&dbPath, "db", "fcmdb.db", "SQLite Database File Path (needed)")
	flag.Parse()

	if len(credsFilename) == 0 {
		println("ERR: credentials file not specified")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		return
	}

	sender := fcmsender.NewFCMSender(credsFilename)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Println("Failed to open database:", err)
		return
	}
	defer db.Close()
	fcmdb := NewFCMDB(db)

	if !isExist(dbPath) {
		if fcmdb.InitializeTables() != nil {
			fmt.Println("Failed to initialize database:", err)
		}
	}

	// Read user subscription into DB
	InputJSONToDB("exampleSubscription.json", fcmdb)

	// Send RSS feed through FCM
	for _, subscription := range fcmdb.GetSubscriptions() {
		err = sendFeed(subscription.url, sender, fcmdb.GetUserTokens(subscription.subscribed_users))
		if err != nil {
			fmt.Println("WARN: error while sending feed. SKIP", err)
		}
	}
}

func sendFeed(url string, fcmsender *fcmsender.FCMSender, tokens []string) error {
	// TODO: fountain codes here. Maybe add in other metadata (e.g. time/sequence number/...)
	// Here is how we turn the raw RSS data into packets (i.e. how we are a transport protocol)
	data, length := getDataPayload(url)
	if length == 0 || length > 2800 {
		// TODO: Support sending slices of a large file through FCM
		return errPacketTooLarge
	}

	var wg sync.WaitGroup
	wg.Add(len(tokens))
	for _, token := range tokens {
		fcmsender.SendTo(data, token, &wg)
	}
	wg.Wait()

	return nil
}

var errPacketTooLarge = errors.New("packet Exceeds Max FCM Packet Size (4000 bytes)")
