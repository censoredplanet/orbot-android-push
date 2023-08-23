package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"os"
)

type JsonSubscription struct {
	Key          string
	Token        string
	Subscription []string
}

func InputJSONToDB(filename string, db *FCMDB) {

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var subscription JsonSubscription
	err = json.Unmarshal(data, &subscription)
	if err != nil {
		panic(err)
	}

	userID := uuid.NewString()

	db.UpdateUser(userID, subscription.Token)
	db.UpdateKey(userID, []byte(subscription.Key))
	for _, site := range subscription.Subscription {
		db.UpdateSubscription(userID, site)
	}

}

func isExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
