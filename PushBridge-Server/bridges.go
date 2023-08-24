package main

import (
	"sync"

	"github.com/censoredplanet/orbot-android-push/PushBridge-server/models"
)

func sendBridgeSettingsToUser(user models.User) error {
	// get the bridges
	// TODO: what if user.Country is null?
	bridgeSettings := models.BridgeSettingsResponseFragment{
		Settings: json.RawMessage(getBridgeSettings(user.Country)),
	}

	var wg sync.WaitGroup
	wg.Add(1)

	bridgeSettingsString, err := json.Marshal(bridgeSettings)
	if err != nil {
		return err
	}

	payload := map[string]string{
		"payload": string(bridgeSettingsString),
	}

	// check length of payload
	var length int
	for _, value := range payload {
		length += len(value)
	}
	if length > 2800 {
		// TODO: Support sending slices of a large file through FCM
		return errPacketTooLarge
	}

	fcmSender.SendTo(payload, user.FCMToken, &wg)

	wg.Wait()
	return nil
}

func getBridgeSettings(country models.Country) string {
	return country.BridgeSetting
}
