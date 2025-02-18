package tokens

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Subject   string    `json:"username"`
	Audience  string    `json:"audience"`
	Issuer    string    `json:"issuer"`
	NotBefore time.Time `json:"not_before"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (p Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: p.ExpiredAt}, nil
}

func (p Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: p.IssuedAt}, nil
}

func (p Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: p.NotBefore}, nil
}

func (p Payload) GetIssuer() (string, error) {
	return p.Issuer, nil
}

func (p Payload) GetSubject() (string, error) {
	return p.Subject, nil
}

func (p Payload) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{p.Audience}, nil
}

type PayloadCreationParams struct {
	Subject   string
	Audience  string
	Issuer    string
	NotBefore time.Time
	Duration  time.Duration
}

func NewPayload(params PayloadCreationParams) (*Payload, error) {
	tokenId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return &Payload{
		ID:        tokenId,
		Subject:   params.Subject,
		Audience:  params.Audience,
		Issuer:    params.Issuer,
		NotBefore: params.NotBefore,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(params.Duration),
	}, nil
}
