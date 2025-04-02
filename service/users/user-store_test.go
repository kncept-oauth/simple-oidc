package users

import (
	"testing"

	"github.com/google/uuid"
)

func TestEncodeAndCompareNone(t *testing.T) {
	salt := uuid.NewString()
	password := uuid.NewString()
	encodeAndCompareType(t, salt, password, EtNone)
}
func TestEncodeAndCompareMD5(t *testing.T) {
	salt := uuid.NewString()
	password := uuid.NewString()
	encodeAndCompareType(t, salt, password, EtMd5)
}
func TestEncodeAndCompareSha512(t *testing.T) {
	salt := uuid.NewString()
	password := uuid.NewString()
	encodeAndCompareType(t, salt, password, EtSha512)
}
func TestEncodeAndCompareBcrypt(t *testing.T) {
	salt := uuid.NewString()
	password := uuid.NewString()
	encodeAndCompareType(t, salt, password, EtBcrypt)
}

func encodeAndCompareType(t *testing.T, salt, password string, encodingType EncodingType) {
	encoded, err := EncodePassword(salt, password, encodingType)
	if err != nil {
		t.Fatalf("error encoding password: %v", err)
	}
	reEncoded, err := EncodePassword(password, salt, encodingType)
	if err != nil {
		t.Fatalf("error encoding password: %v", err)
	}
	if encoded == reEncoded {
		t.Fatalf("must not be able to generate the same password with salt/password reversed")
	}

	pwMatches := ComparePassword(salt, password, reEncoded, encodingType)
	if pwMatches {
		t.Fatalf("must not be able to match incorrect hash ")
	}

	pwMatches = ComparePassword(salt, password, encoded, encodingType)
	if !pwMatches {
		t.Fatalf("password should have matched")
	}

	pwMatches = ComparePassword(salt, password, uuid.NewString(), encodingType)
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
