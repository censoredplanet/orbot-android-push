package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/censoredplanet/orbot-android-push/PushBridge-server/fcmsender"
	"github.com/gin-gonic/gin"
)

// for simplicity, I use global variables here (not the best practice)
var (
	fcmDB     *FCMDB
	fcmSender *fcmsender.FCMSender
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

	// sender
	fcmSender = fcmsender.NewFCMSender(credsFilename)

	// DB
	var err error
	fcmDB, err = NewFCMDB(dbPath)
	if err != nil {
		log.Fatalf("Cannot open database: %v\n", err)
		return
	}
	defer fcmDB.Close()

	router := gin.Default()

	// These two routes are for debugging
	router.GET("/bridges", getAllBridges)
	router.GET("/bridges/:country", getBridgesByCountry)

	// Android Apps will register their tokens here
	router.POST("/fcm/register", registerFCM)

	// admin APIs
	router.POST("/admin/bridges/update", updateBridgesUsingMOAT)
	router.POST("/admin/bridges/set", updateBridgesManually)
	router.POST("/admin/fcm/post", notifyFCM)

	// Run the server
	err = router.Run("0.0.0.0:8888")
	if err != nil {
		log.Fatalf("Error running Gin server: %v", err)
		return
	}
}
