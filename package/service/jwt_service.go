package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrInvalidToken      = fmt.Errorf("invalid token")
	ErrTokenExpired      = fmt.Errorf("token expired")
	ErrTokenUserNotFound = fmt.Errorf("token user not found")
	ErrInvalidTokenType  = fmt.Errorf("invalid token type")
)

const (
	AccessTokenDuration  = time.Hour * 1      // 1 hour
	RefreshTokenDuration = time.Hour * 24 * 7 // 7 days
	TokenTypeBearer      = "Bearer"
	TokenTypeAccess      = "access"
	TokenTypeRefresh     = "refresh"
)

type JWTService interface {
	GenerateTokens(tx *gorm.DB, userId int64) (*schemas.JWTTokens, error)
	ValidateAccessToken(tokenString string) (*schemas.JWTClaims, error)
	ValidateRefreshToken(tokenString string) (*schemas.JWTClaims, error)
	RefreshTokens(tx *gorm.DB, refreshToken string) (*schemas.JWTTokens, error)
	ValidateToken(tx *gorm.DB, tokenString string) (schemas.ValidateTokenResponse, error)
}

type jwtService struct {
	userRepository repository.UserRepository
	secretKey      []byte
	logger         *zap.SugaredLogger
}

// NewJWTService creates a new JWT service instance
func NewJWTService(userRepository repository.UserRepository, secretKey string) JWTService {
	log := utils.NewNamedLogger("jwt_service")
	return &jwtService{
		userRepository: userRepository,
		secretKey:      []byte(secretKey),
		logger:         log,
	}
}

// generateToken creates a JWT token with the given claims
func (j *jwtService) generateToken(claims *schemas.JWTClaims, duration time.Duration) (string, error) {
	now := time.Now()
	jwtClaims := jwt.MapClaims{
		"user_id":  claims.UserID,
		"email":    claims.Email,
		"username": claims.Username,
		"role":     claims.Role,
		"token_id": claims.TokenID,
		"type":     claims.Type,
		"iat":      now.Unix(),
		"exp":      now.Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString(j.secretKey)
}

// parseToken parses and validates a JWT token
func (j *jwtService) parseToken(tokenString string) (*schemas.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		j.logger.Errorf("Error parsing token: %v", err)
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, ErrTokenExpired
		}
	}

	return &schemas.JWTClaims{
		UserID:   int64(claims["user_id"].(float64)),
		Email:    claims["email"].(string),
		Username: claims["username"].(string),
		Role:     claims["role"].(string),
		TokenID:  claims["token_id"].(string),
		Type:     claims["type"].(string),
	}, nil
}

// GenerateTokens creates both access and refresh tokens for a user
func (j *jwtService) GenerateTokens(tx *gorm.DB, userId int64) (*schemas.JWTTokens, error) {
	user, err := j.userRepository.Get(tx, userId)
	if err != nil {
		j.logger.Errorf("Error getting user by id: %v", err.Error())
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTokenUserNotFound
		}
		return nil, err
	}

	tokenID := uuid.New().String()
	now := time.Now()

	// Create access token claims
	accessClaims := &schemas.JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     string(user.Role),
		TokenID:  tokenID,
		Type:     TokenTypeAccess,
	}

	// Create refresh token claims
	refreshClaims := &schemas.JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     string(user.Role),
		TokenID:  tokenID,
		Type:     TokenTypeRefresh,
	}

	accessToken, err := j.generateToken(accessClaims, AccessTokenDuration)
	if err != nil {
		j.logger.Errorf("Error generating access token: %v", err)
		return nil, err
	}

	refreshToken, err := j.generateToken(refreshClaims, RefreshTokenDuration)
	if err != nil {
		j.logger.Errorf("Error generating refresh token: %v", err)
		return nil, err
	}

	return &schemas.JWTTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(AccessTokenDuration),
		TokenType:    TokenTypeBearer,
	}, nil
}

// ValidateAccessToken validates an access token
func (j *jwtService) ValidateAccessToken(tokenString string) (*schemas.JWTClaims, error) {
	claims, err := j.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeAccess {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (j *jwtService) ValidateRefreshToken(tokenString string) (*schemas.JWTClaims, error) {
	claims, err := j.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// RefreshTokens generates new tokens using a valid refresh token
func (j *jwtService) RefreshTokens(tx *gorm.DB, refreshToken string) (*schemas.JWTTokens, error) {
	claims, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		j.logger.Errorf("Error validating refresh token: %v", err)
		return nil, err
	}

	return j.GenerateTokens(tx, claims.UserID)
}

// ValidateToken validates a token and returns user information
func (j *jwtService) ValidateToken(tx *gorm.DB, tokenString string) (schemas.ValidateTokenResponse, error) {
	claims, err := j.ValidateAccessToken(tokenString)
	if err != nil {
		j.logger.Errorf("Error validating access token: %v", err)
		return schemas.ValidateTokenResponse{Valid: false, User: InvalidUser}, err
	}

	user, err := j.userRepository.Get(tx, claims.UserID)
	if err != nil {
		j.logger.Errorf("Error getting user by id: %v", err.Error())
		if err == gorm.ErrRecordNotFound {
			return schemas.ValidateTokenResponse{Valid: false, User: InvalidUser}, ErrTokenUserNotFound
		}
		return schemas.ValidateTokenResponse{Valid: false, User: InvalidUser}, err
	}

	currentUser := schemas.User{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		Name:     user.Name,
		Surname:  user.Surname,
	}

	return schemas.ValidateTokenResponse{Valid: true, User: currentUser}, nil
}
