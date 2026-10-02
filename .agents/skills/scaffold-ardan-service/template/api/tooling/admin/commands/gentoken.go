package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tmplorg/tmplsvc/app/sdk/auth"
	"github.com/tmplorg/tmplsvc/business/domain/userbus"
	"github.com/tmplorg/tmplsvc/business/domain/userbus/stores/userdb"
	"github.com/tmplorg/tmplsvc/business/sdk/sqldb"
	"github.com/tmplorg/tmplsvc/business/types/role"
	"github.com/tmplorg/tmplsvc/foundation/keystore"
	"github.com/tmplorg/tmplsvc/foundation/logger"
)

// GenToken generates a JWT for the specified user.
func GenToken(log *logger.Logger, dbConfig sqldb.Config, keyPath string, issuer string, userID uuid.UUID, kid string) error {
	if kid == "" {
		fmt.Println("help: gentoken <user_id> <kid>")
		return ErrHelp
	}

	db, err := sqldb.Open(dbConfig)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userBus := userbus.NewBusiness(log, nil, userdb.NewStore(log, db))

	usr, err := userBus.QueryByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("retrieve user: %w", err)
	}

	ks := keystore.New()

	n, err := ks.LoadByFileSystem(os.DirFS(keyPath))
	if err != nil {
		return fmt.Errorf("loading keys by fs: %w", err)
	}

	if n == 0 {
		return errors.New("no keys exist")
	}

	authCfg := auth.Config{
		Log:       log,
		UserBus:   userBus,
		KeyLookup: ks,
		Issuer:    issuer,
	}

	ath := auth.New(authCfg)

	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := auth.Claims{
		Subject:   usr.ID.String(),
		Issuer:    ath.Issuer(),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Roles:     role.ParseToString(usr.Roles),
	}

	// This will generate a JWT with the claims embedded in them. The service
	// validates these claims with the public key that matches the kid.
	token, err := ath.GenerateToken(kid, claims)
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	fmt.Printf("-----BEGIN TOKEN-----\n%s\n-----END TOKEN-----\n", token)
	return nil
}
