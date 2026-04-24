// Package stickers manages the FIFA 2026 sticker catalog.
package stickers

import "time"

// Sticker represents a single FIFA 2026 Panini sticker.
type Sticker struct {
	ID            string     `json:"id"`
	CountryCode   string     `json:"countryCode"`
	StickerNumber int        `json:"stickerNumber"`
	Type          string     `json:"type"`
	PlayerName    string     `json:"playerName,omitempty"`
	DOB           *time.Time `json:"dob,omitempty"`
	Height        *float32   `json:"height,omitempty"`
	Weight        *float32   `json:"weight,omitempty"`
	Club          string     `json:"club,omitempty"`
	ClubCountry   string     `json:"clubCountry,omitempty"`
	Position      string     `json:"position,omitempty"`
	NationalTeam  string     `json:"nationalTeam"`
	ImageURL      string     `json:"imageUrl,omitempty"`
}
