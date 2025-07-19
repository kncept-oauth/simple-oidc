package users

import (
	"testing"

	"github.com/google/uuid"
)

func TestEncodeAndCompareNone(t *testing.T) {
	salt := GenerateSalt(EtNone)
	password := uuid.NewString()
	encodeAndCompareType(t, salt, password)
}
func TestEncodeAndCompareMD5(t *testing.T) {
	salt := GenerateSalt(EtMd5)
	password := uuid.NewString()
	encodeAndCompareType(t, salt, password)
}
func TestEncodeAndCompareSha512(t *testing.T) {
	salt := GenerateSalt(EtSha512)
	password := uuid.NewString()
	encodeAndCompareType(t, salt, password)
}
func TestEncodeAndCompareBcrypt(t *testing.T) {
	salt := GenerateSalt(EtBcrypt)
	password := uuid.NewString() // max 72 bytes length(!)
	encodeAndCompareType(t, salt, password)
}

func encodeAndCompareType(t *testing.T, salt, password string) {
	encoded, err := EncodePassword(salt, password)
	if err != nil {
		t.Fatalf("error encoding password: %v", err)
	}
	reEncoded, err := EncodePassword(password, salt)
	if err != nil {
		t.Fatalf("error encoding password: %v", err)
	}
	if encoded == reEncoded {
		t.Fatalf("must not be able to generate the same password with salt/password reversed")
	}

	pwMatches := ComparePassword(salt, password, reEncoded)
	if pwMatches {
		t.Fatalf("must not be able to match incorrect hash ")
	}

	pwMatches = ComparePassword(salt, password, encoded)
	if !pwMatches {
		t.Fatalf("password should have matched")
	}

	pwMatches = ComparePassword(salt, password, uuid.NewString())
	if pwMatches {
		t.Fatalf("must not be able to match random hash ")
	}
}

func TestSalt(t *testing.T) {
	salt := GenerateSalt()
	encodingType := GetEncodingType(salt)
	if encodingType != EtBcrypt {
		t.Fatalf("Wrong Encoding Type: %v", encodingType)
	}

	salt = GenerateSalt(EtSha512)
	encodingType = GetEncodingType(salt)
	if encodingType != EtSha512 {
		t.Fatalf("Wrong Encoding Type: %v", encodingType)
	}

	salt = GenerateSalt(EtBcrypt)
	encodingType = GetEncodingType(salt)
	if encodingType != EtBcrypt {
		t.Fatalf("Wrong Encoding Type: %v", encodingType)
	}
}
