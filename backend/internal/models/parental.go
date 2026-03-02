package models

// ParentalStatus represents the parental control state visible to users.
type ParentalStatus struct {
	PinSet              bool   `json:"pin_set"`
	AdultContentEnabled bool   `json:"adult_content_enabled"`
}

// AdminParentalStatus extends ParentalStatus with admin-visible fields.
type AdminParentalStatus struct {
	PinSet               bool     `json:"pin_set"`
	RestrictedGenreCodes []string `json:"restricted_genre_codes"`
}

// SetPinInput is the request body for setting a parental PIN.
type SetPinInput struct {
	Pin string `json:"pin" binding:"required,min=4,max=6,numeric"`
}

// VerifyPinInput is the request body for verifying a parental PIN.
type VerifyPinInput struct {
	Pin string `json:"pin" binding:"required"`
}

// UpdateRestrictedGenresInput is the request body for updating restricted genre codes.
type UpdateRestrictedGenresInput struct {
	Codes []string `json:"codes" binding:"required"`
}

// SetUserAdultContentInput is the request body for admin to set user adult content access.
type SetUserAdultContentInput struct {
	AdultContentEnabled bool `json:"adult_content_enabled"`
}

// UserAdultStatus represents a user's adult content access status (for admin listing).
type UserAdultStatus struct {
	UserID              string `json:"user_id"`
	Username            string `json:"username"`
	DisplayName         string `json:"display_name"`
	Role                string `json:"role"`
	AdultContentEnabled bool   `json:"adult_content_enabled"`
}
