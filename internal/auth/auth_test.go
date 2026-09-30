package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-secret"

func signClaims(t *testing.T, method jwt.SigningMethod, key any, claims jwt.RegisteredClaims) string {
	t.Helper()

	ss, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	return ss
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	expiresIn := time.Hour

	tokenString, err := MakeJWT(userID, testSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	claims := jwt.RegisteredClaims{}
	_, err = jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token from MakeJWT(): %v", err)
	}

	if claims.Issuer != "chirpy-access" {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, "chirpy-access")
	}
	if claims.Subject != userID.String() {
		t.Errorf("Subject = %q, want %q", claims.Subject, userID.String())
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatalf("IssuedAt = %v, ExpiresAt = %v, want both set", claims.IssuedAt, claims.ExpiresAt)
	}

	wantExpiry := time.Now().Add(expiresIn)
	if diff := claims.ExpiresAt.Sub(wantExpiry).Abs(); diff > 5*time.Second {
		t.Errorf("ExpiresAt = %v, want within 5s of %v", claims.ExpiresAt, wantExpiry)
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()

	validToken, err := MakeJWT(userID, testSecret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	expiredToken, err := MakeJWT(userID, testSecret, -time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	validClaims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Subject:   userID.String(),
	}

	wrongIssuerClaims := validClaims
	wrongIssuerClaims.Issuer = "someone-else"

	badSubjectClaims := validClaims
	badSubjectClaims.Subject = "not-a-uuid"

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:        "valid token",
			tokenString: validToken,
			tokenSecret: testSecret,
			wantUserID:  userID,
		},
		{
			name:        "expired token",
			tokenString: expiredToken,
			tokenSecret: testSecret,
			wantErr:     true,
			wantErrIs:   jwt.ErrTokenExpired,
		},
		{
			name:        "wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong-secret",
			wantErr:     true,
			wantErrIs:   jwt.ErrTokenSignatureInvalid,
		},
		{
			name:        "malformed token",
			tokenString: "not.a.jwt",
			tokenSecret: testSecret,
			wantErr:     true,
			wantErrIs:   jwt.ErrTokenMalformed,
		},
		{
			name:        "empty token",
			tokenString: "",
			tokenSecret: testSecret,
			wantErr:     true,
			wantErrIs:   jwt.ErrTokenMalformed,
		},
		{
			name:        "unsigned token with none algorithm",
			tokenString: signClaims(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, validClaims),
			tokenSecret: testSecret,
			wantErr:     true,
			wantErrIs:   jwt.ErrTokenSignatureInvalid,
		},
		{
			name:        "unexpected signing method",
			tokenString: signClaims(t, jwt.SigningMethodHS512, []byte(testSecret), validClaims),
			tokenSecret: testSecret,
			wantErr:     true,
			wantErrIs:   jwt.ErrTokenSignatureInvalid,
		},
		{
			name:        "wrong issuer",
			tokenString: signClaims(t, jwt.SigningMethodHS256, []byte(testSecret), wrongIssuerClaims),
			tokenSecret: testSecret,
			wantErr:     true,
			wantErrIs:   jwt.ErrTokenInvalidIssuer,
		},
		{
			name:        "subject is not a uuid",
			tokenString: signClaims(t, jwt.SigningMethodHS256, []byte(testSecret), badSubjectClaims),
			tokenSecret: testSecret,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("ValidateJWT() error = %v, want errors.Is %v", err, tt.wantErrIs)
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() userID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}
