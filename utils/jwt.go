package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gitee.com/palagend/fx/config"
	"gitee.com/palagend/fx/models"
)

var (
	jwtSecret        = []byte("your-secret-key-change-in-production")
	accessTokenTTL   = 15 * time.Minute
	refreshTokenTTL  = 7 * 24 * time.Hour
)

type TokenClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func GenerateTokenPair(userID uint, username string) (*TokenPair, error) {
	accessToken, err := generateAccessToken(userID, username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken(userID, username)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
	}, nil
}

func generateAccessToken(userID uint, username string) (string, error) {
	claims := TokenClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func generateRefreshToken(userID uint, username string) (string, error) {
	claims := TokenClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	db := config.GetDB()
	refreshTokenModel := models.RefreshToken{
		UserID:    userID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(refreshTokenTTL),
	}

	if err := db.Create(&refreshTokenModel).Error; err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := ParseToken(refreshTokenString)
	if err != nil {
		return nil, errors.New("无效的刷新令牌")
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("无效的令牌类型")
	}

	db := config.GetDB()
	var refreshToken models.RefreshToken
	if err := db.Where("token = ?", refreshTokenString).First(&refreshToken).Error; err != nil {
		return nil, errors.New("刷新令牌不存在或已过期")
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		db.Delete(&refreshToken)
		return nil, errors.New("刷新令牌已过期")
	}

	db.Delete(&refreshToken)

	return GenerateTokenPair(claims.UserID, claims.Username)
}

func RevokeRefreshToken(refreshTokenString string) error {
	db := config.GetDB()
	return db.Where("token = ?", refreshTokenString).Delete(&models.RefreshToken{}).Error
}

func RevokeAllUserRefreshTokens(userID uint) error {
	db := config.GetDB()
	return db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}
