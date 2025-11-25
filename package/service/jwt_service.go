package service

import (
	"fmt"
	"time"

	"github.com/mini-maxit/backend/internal/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	AccessTokenDuration  = time.Hour * 1      // 1 hour
	RefreshTokenDuration = time.Hour * 24 * 7 // 7 days
	TokenTypeBearer      = "Bearer"
	TokenTypeAccess      = "access"
	TokenTypeRefresh     = "refresh"
)

type JWTService interface {
	GenerateTokens(db database.Database, userId int64) (*schemas.JWTTokens, error)
	RefreshTokens(db database.Database, refreshToken string) (*schemas.JWTTokens, error)
	AuthenticateToken(db database.Database, tokenString string) (*schemas.ValidateTokenResponse, error)
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
		return nil, errors.ErrInvalidToken
	}

	if !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.ErrInvalidToken
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, errors.ErrTokenExpired
		}
	}

	return &schemas.JWTClaims{
		UserID:   int64(claims["user_id"].(float64)),
		Username: claims["username"].(string),
		Role:     claims["role"].(string),
		TokenID:  claims["token_id"].(string),
		Type:     claims["type"].(string),
	}, nil
}

// GenerateTokens creates both access and refresh tokens for a user
func (j *jwtService) GenerateTokens(db database.Database, userId int64) (*schemas.JWTTokens, error) {
	user, err := j.userRepository.Get(db, userId)
	if err != nil {
		j.logger.Errorf("Error getting user by id: %v", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrTokenUserNotFound
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
	}, nil
}

// validateAccessToken validates an access token
func (j *jwtService) validateAccessToken(tokenString string) (*schemas.JWTClaims, error) {
	claims, err := j.parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	j.logger.Debugf("Parsed claims: %+v", claims)

	if claims.Type != TokenTypeAccess {
		return nil, errors.ErrInvalidTokenType
	}

	return claims, nil
}

// validateRefreshToken validates a refresh token
func (j *jwtService) validateRefreshToken(tokenString string) (*schemas.JWTClaims, error) {
	claims, err := j.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeRefresh {
		return nil, errors.ErrInvalidTokenType
	}

	return claims, nil
}

// RefreshTokens generates new tokens using a valid refresh token
func (j *jwtService) RefreshTokens(db database.Database, refreshToken string) (*schemas.JWTTokens, error) {
	claims, err := j.validateRefreshToken(refreshToken)
	if err != nil {
		j.logger.Errorf("Error validating refresh token: %v", err)
		return nil, err
	}

	return j.GenerateTokens(db, claims.UserID)
}

// AuthenticateToken validates a token and returns user information
func (j *jwtService) AuthenticateToken(db database.Database, tokenString string) (*schemas.ValidateTokenResponse, error) {
	claims, err := j.validateAccessToken(tokenString)
	if err != nil {
		j.logger.Errorf("Error validating access token: %v", err)
		return &schemas.ValidateTokenResponse{Valid: false, User: InvalidUser}, err
	}

	user, err := j.userRepository.Get(db, claims.UserID)
	if err != nil {
		j.logger.Errorf("Error getting user by id: %v", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &schemas.ValidateTokenResponse{Valid: false, User: InvalidUser}, errors.ErrTokenUserNotFound
		}
		return &schemas.ValidateTokenResponse{Valid: false, User: InvalidUser}, err
	}
	if user.Role != types.UserRole(claims.Role) {
		return &schemas.ValidateTokenResponse{Valid: false, User: InvalidUser}, errors.ErrNotAuthorized
	}

	return &schemas.ValidateTokenResponse{Valid: true, User: *UserToSchema(user)}, nil
}
