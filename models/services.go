package appschema

// face scanner service stream structs
type Quantitative struct {
	Percentage  float64 `json:"percentage"`
	Coordinates any     `json:"coordinates"`
}

type Qualitative struct {
	IsPresent   bool `json:"is_present"`
	Coordinates any  `json:"coordinates"`
}

type FaceScannerResponse struct {
	Data FaceScanData `json:"data"`
}

type FaceScanData struct {
	Qualitative  []map[string]Qualitative  `json:"qualitative"`
	Quantitative []map[string]Quantitative `json:"quantitative"`
}

// user service structs
type GetUserData struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Gender     string `json:"gender"`
	Email      string `json:"email"`
	DOB        string `json:"dob"`
	Type       int    `json:"type"`
	SkinType   string `json:"skin_type"`
	IsVerified bool   `json:"is_active"`
	CreatedAt  string `json:"created_at"`
}
