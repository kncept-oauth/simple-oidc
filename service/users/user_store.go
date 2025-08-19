package users

import (
	"context"
	"crypto/md5"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	GetUser(ctx context.Context, id string) (*OidcUser, error)
	SaveUser(ctx context.Context, user *OidcUser) error
}

type OidcUser struct {
	Id              string `dynamodbav:"id"`
	Salt            string `dynamodbav:"salt"`
	EncodedPassword string `dynamodbav:"pass"`
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

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano())) // Seed the random number generator
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GenerateSalt(encodingVersions ...EncodingType) string {
	if len(encodingVersions) > 1 {
		panic("must supply at most one encoding version")
	}
	if len(encodingVersions) == 0 {
		encodingVersions = []EncodingType{EtBcrypt} // default to latest
	}
	return fmt.Sprintf("%v:%v", encodingVersions[0], generateRandomString(16))
}
func GetEncodingType(salt string) EncodingType {
	idx := strings.Index(salt, ":")
	if idx == -1 {
		return ToEncodingType(salt)
	}
	return ToEncodingType(salt[:idx])
}

func EncodePassword(salt, password string) (string, error) {
	password = fmt.Sprintf("%v%v", salt, password)
	encodingType := GetEncodingType(salt)
	switch encodingType {
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
		return "", fmt.Errorf("unsupported Encoding version %v", encodingType)
	}
}

func ComparePassword(salt, password string, hash string) bool {
	encodingType := GetEncodingType(salt)
	switch encodingType {
	case EtBcrypt:
		password = fmt.Sprintf("%v%v", salt, password)
		encoded, err := base64.StdEncoding.DecodeString(hash)
		if err != nil {
			return false
		}
		err = bcrypt.CompareHashAndPassword(encoded, []byte(password))
		return err == nil
	default:
		expected, err := EncodePassword(salt, password)
		return err == nil && expected == hash
	}
}

func (u *OidcUser) SetPassword(rawPassword string) error {
	salt := GenerateSalt()
	encodedPassword, err := EncodePassword(salt, rawPassword)
	if err != nil {
		return nil
	}
	u.Salt = salt
	u.EncodedPassword = encodedPassword
	return nil
}

func (u *OidcUser) PasswordMatches(rawPassword string) bool {
	return ComparePassword(u.Salt, rawPassword, u.EncodedPassword)
}
