package models

type User struct {
	ModelUsingUUID
	FCMToken string `json:"token"` // FCM deviceToken that identifies a device

	// `User` belongs to `Country`, `CountryID` is the foreign key
	Country   Country
	CountryID string
}

type Country struct {
	CountryCode   string `gorm:"primaryKey" json:"code"` // ISO 3166-2 codes (e.g. us)
	BridgeSetting string // for simplicity, we'll just forward the settings JSON as a string for now
	ModelWithoutID
}

//type BridgeSetting struct {
//	Bridges []Bridge `json:"bridges"`
//	Source  string   `json:"source"`
//	Type    string   `json:"type"`
//}
//
//type Bridge struct {
//	String string
//}

// PublicKey is the RSA pubkey for encryption. Not Implemented for now.
//type PublicKey struct {
//	ModelUsingUUID
//	User      User      // belongs to User
//	UserID    uuid.UUID `json:"user_id"`
//	Algorithm int       `json:"algorithm"`
//	IsAuth    int       `json:"is_auth"`
//	KeyBytes  []byte    `json:"key_bytes"`
//}

//type Subscription struct {
//	URL             string   `json:"url"`
//	SubscribedUsers []string `json:"subscribedUsers"`
//}
