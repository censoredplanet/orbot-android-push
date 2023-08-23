package models

import "encoding/json"

type ServerErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type BridgeSettingResponse struct {
	Country string `json:"country"`
	*BridgeSettingsResponseFragment
}

type BridgeSettingsResponseFragment struct {
	Settings json.RawMessage `json:"settings"`
}

type AllBridgeSettingResponse = map[string]BridgeSettingsResponseFragment
