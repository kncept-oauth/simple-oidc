package users

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	GetUser(id string) (*OidcUser, error)
	SaveUser(user *OidcUser) error
}

type OidcUser struct {
	Id              string
	EncodedSalt     string
	EncodedPassword string
}

type EncodingType string

const (
	EtNone   EncodingType = "EtNone"
	EtMd5    EncodingType = "EtMd5"
	EtSha512 EncodingType = "EtSha512"
	EtBcrypt EncodingType = "EtBcrypt"
)

func (obj EncodingType) String() string {
	return string(obj)
}
func ToEncodingType(s string) EncodingType {
	if s == EtNone.String() {
		return EtNone
	}
	if s == EtMd5.String() {
		return EtMd5
	}
	if s == EtSha512.String() {
		return EtSha512
	}
	if s == EtBcrypt.String() {
		return EtBcrypt
	}
	return EtNone
}

func GenerateSalt(encodingVersions ...EncodingType) string {
	if len(encodingVersions) > 1 {
		panic("must supply at most one encoding version")
	}
	if len(encodingVersions) == 0 {
		encodingVersions = []EncodingType{EtBcrypt} // default to latest
	}
	return fmt.Sprintf("%v:%v", encodingVersions[0], uuid.NewString())
}
func GetEncodingType(salt string) EncodingType {
	idx := strings.Index(salt, ":")
	if idx == -1 {
		return ToEncodingType(salt)
	}
	return ToEncodingType(salt[0:idx])
}

func EncodePassword(salt, password string, encodingVersions ...EncodingType) (string, error) {
	password = fmt.Sprintf("%v%v", salt, password)
	if len(encodingVersions) > 1 {
		return "", fmt.Errorf("must supply at most one encoding version")
	}
	if len(encodingVersions) == 0 {
		encodingVersions = []EncodingType{EtBcrypt} // default to latest
	}
	switch encodingVersions[0] {
	case EtNone:
		return password, nil
	case EtMd5:
		return hex.EncodeToString(md5.New().Sum([]byte(password))), nil
	case EtSha512:
		return hex.EncodeToString(sha512.New().Sum([]byte(password))), nil
	case EtBcrypt:
		data, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		return base64.StdEncoding.EncodeToString(data), err
	default:
		return "", fmt.Errorf("unsupported Encoding version %v", encodingVersions[0])
	}
}

func ComparePassword(salt, password string, hash string, encodingVersions ...EncodingType) bool {
	if len(encodingVersions) > 1 {
		return false
	}
	if len(encodingVersions) == 0 {
		encodingVersions = []EncodingType{EtBcrypt} // default to latest
	}
	switch encodingVersions[0] {
	case EtBcrypt:
		password = fmt.Sprintf("%v%v", salt, password)
		encoded, err := base64.StdEncoding.DecodeString(hash)
		if err != nil {
			return false
		}
		err = bcrypt.CompareHashAndPassword(encoded, []byte(password))
		return err == nil
	default:
		expected, err := EncodePassword(salt, password, encodingVersions[0])
		return err == nil && expected == hash
	}
}
