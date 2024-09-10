package tokens

import (
	"time"

	"github.com/8thgencore/microservice-auth/internal/model"
)

// TokenOperations is the interface for token functions.
type TokenOperations interface {
	Generate(user model.User, secretKey []byte, duration time.Duration) (string, error)
	Verify(tokenStr string, secretKey []byte) (*model.UserClaims, error)
}
