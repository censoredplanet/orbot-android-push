package models

import "encoding/json"

type UpdateBridgesManuallyRequest struct {
	Country  string          `json:"country"`
	Settings json.RawMessage `json:"settings"` // json.RawMessage type is required to get around the "cannot unmarshal array into ... of type string" error
}

type RegisterFCMRequest struct {
	FCMToken string `json:"token"`
	Country  string `json:"country"`
}

type NotifyFCMByTokenRequest struct {
	UserID string `json:"userID"` // the UUID of the user
}

type NotifyFCMByCountryRequest struct {
	CountryCode string `json:"country" example:"us"` // ISO 3166-2 codes (e.g. us)
}
