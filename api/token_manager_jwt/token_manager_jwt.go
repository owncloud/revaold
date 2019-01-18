package token_manager_jwt

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/zap"
	"github.com/owncloud/revaold/api"
	"go.uber.org/zap"
)

func New(signSecret string, publicKey *rsa.PublicKey) api.TokenManager {
	return &tokenManager{signSecret: signSecret, publicKey: publicKey}
}

type tokenManager struct {
	signSecret string
	publicKey  *rsa.PublicKey
}

func (tm *tokenManager) ForgeUserToken(ctx context.Context, user *api.User) (string, error) {
	l := ctx_zap.Extract(ctx)
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	claims := token.Claims.(jwt.MapClaims)
	claims["account_id"] = user.AccountId
	claims["display_name"] = user.DisplayName
	claims["groups"] = user.Groups
	claims["exp"] = time.Now().Add(time.Second * time.Duration(3600))
	tokenString, err := token.SignedString([]byte(tm.signSecret))
	if err != nil {
		l.Error("", zap.Error(err))
		return "", err
	}
	return tokenString, nil
}

func (tm *tokenManager) DismantleUserToken(ctx context.Context, token string) (*api.User, error) {
	l := ctx_zap.Extract(ctx)
	rawToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		l.Debug("parsed", zap.Any("token", token))
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			return tm.publicKey, nil
		} else if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			return []byte(tm.signSecret), nil
		} else {
			return nil, errors.New("unsupported token method")
		}
	})
	if err != nil {
		l.Error("invalid token", zap.Error(err), zap.String("token", token))
		return nil, err
	}
	if !rawToken.Valid {
		l.Error("invalid token", zap.Error(err), zap.String("token", token))
		return nil, err

	}

	claims := rawToken.Claims.(jwt.MapClaims)
	var accountID string
	var displayName string
	groups := []string{}

	if kopanoIdentity, ok := claims["kc.identity"].(map[string]interface{}); ok {
		accountID, ok = kopanoIdentity["kc.i.un"].(string)
		if !ok {
			return nil, errors.New("kc.identity kc.i.un claim is not a string")
		}

		displayName, _ = kopanoIdentity["kc.i.dn"].(string) // no displayname is not an error

		// FIXME ... fetch groups from userInfo? LDAP?
		// note that OIDC explicitly lists groups as an example for a claim that is not in the spec: https://openid.net/specs/openid-connect-core-1_0.html#ClaimsParameter

	} else {

		accountID, ok = claims["account_id"].(string)
		if !ok {
			return nil, errors.New("account_id claim is not a string")
		}

		displayName, _ = claims["display_name"].(string) // no displayname is not an error

		rawGroups, ok := claims["groups"].([]interface{})
		if !ok {
			return nil, errors.New("groups claim is not a []interface{}")
		}
		for _, g := range rawGroups {
			group, ok := g.(string)
			if !ok {
				err := errors.New(fmt.Sprintf("group %+v can not be casted to string", g))
				l.Error("", zap.Error(err))
				return nil, err
			}
			groups = append(groups, group)
		}
	}

	user := &api.User{
		AccountId:   accountID,
		Groups:      groups,
		DisplayName: displayName,
	}
	return user, nil
}

func (tm *tokenManager) ForgePublicLinkToken(ctx context.Context, pl *api.PublicLink) (string, error) {
	l := ctx_zap.Extract(ctx)
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	claims := token.Claims.(jwt.MapClaims)
	claims["token"] = pl.Token
	claims["owner"] = pl.OwnerId
	claims["id"] = pl.Id
	claims["path"] = pl.Path
	claims["protected"] = pl.Protected
	claims["expires"] = pl.Expires
	claims["read_only"] = pl.ReadOnly
	claims["drop_only"] = pl.DropOnly
	claims["mtime"] = pl.Mtime
	claims["item_type"] = pl.ItemType
	claims["share_name"] = pl.Name
	claims["exp"] = time.Now().Add(time.Second * time.Duration(3600))
	tokenString, err := token.SignedString([]byte(tm.signSecret))
	if err != nil {
		l.Error("", zap.Error(err))
		return "", err
	}
	return tokenString, nil
}

func (tm *tokenManager) DismantlePublicLinkToken(ctx context.Context, token string) (*api.PublicLink, error) {
	l := ctx_zap.Extract(ctx)
	rawToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(tm.signSecret), nil
	})
	if err != nil {
		l.Error("invalid token", zap.Error(err), zap.String("token", token))
		return nil, err
	}
	if !rawToken.Valid {
		l.Error("invalid token", zap.Error(err), zap.String("token", token))
		return nil, err

	}

	/*
		"exp": "2018-07-24T10:11:11.827901148+02:00",
		"expires": 0,
		"id": "103",
		"item_type": 0,
		"mtime": 1532362779,
		"owner": "gonzalhu",
		"path": "oldhome:22510091102060544",
		"protected": false,
		"read_only": true,
		"token": "fgDsc2WD8F2qNfH"
	*/
	claims := rawToken.Claims.(jwt.MapClaims)
	token, ok := claims["token"].(string)
	if !ok {
		return nil, errors.New("token claim is not a string")
	}
	owner, ok := claims["owner"].(string)
	if !ok {
		return nil, errors.New("owner claim is not a string")
	}
	readOnly, ok := claims["read_only"].(bool)
	if !ok {
		return nil, errors.New("read_only claim is not a bool")
	}
	dropOnly, ok := claims["drop_only"].(bool)
	if !ok {
		return nil, errors.New("drop_only claim is not a bool")
	}
	path, ok := claims["path"].(string)
	if !ok {
		return nil, errors.New("path claim is not a string")
	}
	protected, ok := claims["protected"].(bool)
	if !ok {
		return nil, errors.New("protected claim is not a bool")
	}
	mtime, ok := claims["mtime"].(float64)
	if !ok {
		return nil, errors.New("mtime claim is not a float64")
	}
	itemType, ok := claims["item_type"].(float64)
	if !ok {
		return nil, errors.New("item_type claim is not a float64")
	}
	shareName, ok := claims["share_name"].(string)
	if !ok {
		return nil, errors.New("share_name claim is not a string")
	}

	pl := &api.PublicLink{
		Token:     token,
		OwnerId:   owner,
		ReadOnly:  readOnly,
		Path:      path,
		Protected: protected,
		Mtime:     uint64(mtime),
		ItemType:  api.PublicLink_ItemType(itemType),
		Name:      shareName,
		DropOnly:  dropOnly,
	}
	return pl, nil
}
