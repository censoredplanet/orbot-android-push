package main

import (
	"flag"
	"fmt"
	"github.com/censoredplanet/orbot-android-push/PushBridge-server/fcmsender"
	"github.com/gin-gonic/gin"
	"log"
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

	//// These two routes are for debugging
	//router.GET("/bridges", getAllBridges)
	//router.GET("/bridges/:country", getBridgesByCountry)
	//
	//// Android Apps will register their tokens here
	//router.POST("/fcm/register", registerFCM)
	//
	//// admin APIs
	//router.POST("/admin/bridges/update", updateBridgesUsingMOAT)
	//router.POST("/admin/bridges/set", updateBridgesManually)
	//router.POST("/admin/fcm/post", notifyFCM)

	// Run the server
	err = router.Run(":8080")
	if err != nil {
		log.Fatalf("Error running Gin server: %v", err)
		return
	}
}

//func sendFeed(url string, fcmsender *fcmsender.FCMSender, tokens []string) error {
//	// TODO: fountain codes here. Maybe add in other metadata (e.g. time/sequence number/...)
//	// Here is how we turn the raw RSS data into packets (i.e. how we are a transport protocol)
//	data, length := getDataPayload(url)
//	if length == 0 || length > 2800 {
//		// TODO: Support sending slices of a large file through FCM
//		return errPacketTooLarge
//	}
//
//	var wg sync.WaitGroup
//	wg.Add(len(tokens))
//	for _, token := range tokens {
//		fcmsender.SendTo(data, token, &wg)
//	}
//	wg.Wait()
//
//	return nil
//}
