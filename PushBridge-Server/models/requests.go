package models

import "encoding/json"

type UpdateBridgesManuallyRequest struct {
	Country  string          `json:"country"`
	Settings json.RawMessage `json:"settings"` // json.RawMessage type is required to get around the "cannot unmarshal array into ... of type string" error
}
