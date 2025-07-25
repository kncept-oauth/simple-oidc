// Code generated by ogen, DO NOT EDIT.

package api

import (
	"math/bits"
	"strconv"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/validate"
)

// Encode encodes AuthorizeGetBadRequestApplicationJSON as json.
func (s AuthorizeGetBadRequestApplicationJSON) Encode(e *jx.Encoder) {
	unwrapped := string(s)

	e.Str(unwrapped)
}

// Decode decodes AuthorizeGetBadRequestApplicationJSON from json.
func (s *AuthorizeGetBadRequestApplicationJSON) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode AuthorizeGetBadRequestApplicationJSON to nil")
	}
	var unwrapped string
	if err := func() error {
		v, err := d.Str()
		unwrapped = string(v)
		if err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return errors.Wrap(err, "alias")
	}
	*s = AuthorizeGetBadRequestApplicationJSON(unwrapped)
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s AuthorizeGetBadRequestApplicationJSON) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *AuthorizeGetBadRequestApplicationJSON) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *JWKResponse) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *JWKResponse) encodeFields(e *jx.Encoder) {
	{
		if s.Kty.Set {
			e.FieldStart("kty")
			s.Kty.Encode(e)
		}
	}
	{
		if s.Use.Set {
			e.FieldStart("use")
			s.Use.Encode(e)
		}
	}
	{
		if s.KeyOps.Set {
			e.FieldStart("key_ops")
			s.KeyOps.Encode(e)
		}
	}
	{
		if s.Alg.Set {
			e.FieldStart("alg")
			s.Alg.Encode(e)
		}
	}
	{
		if s.Kid.Set {
			e.FieldStart("kid")
			s.Kid.Encode(e)
		}
	}
	{
		if s.X5u.Set {
			e.FieldStart("x5u")
			s.X5u.Encode(e)
		}
	}
	{
		if s.X5c.Set {
			e.FieldStart("x5c")
			s.X5c.Encode(e)
		}
	}
	{
		if s.X5t.Set {
			e.FieldStart("x5t")
			s.X5t.Encode(e)
		}
	}
	{
		if s.N.Set {
			e.FieldStart("n")
			s.N.Encode(e)
		}
	}
	{
		if s.E.Set {
			e.FieldStart("e")
			s.E.Encode(e)
		}
	}
}

var jsonFieldsNameOfJWKResponse = [10]string{
	0: "kty",
	1: "use",
	2: "key_ops",
	3: "alg",
	4: "kid",
	5: "x5u",
	6: "x5c",
	7: "x5t",
	8: "n",
	9: "e",
}

// Decode decodes JWKResponse from json.
func (s *JWKResponse) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode JWKResponse to nil")
	}

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "kty":
			if err := func() error {
				s.Kty.Reset()
				if err := s.Kty.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"kty\"")
			}
		case "use":
			if err := func() error {
				s.Use.Reset()
				if err := s.Use.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"use\"")
			}
		case "key_ops":
			if err := func() error {
				s.KeyOps.Reset()
				if err := s.KeyOps.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"key_ops\"")
			}
		case "alg":
			if err := func() error {
				s.Alg.Reset()
				if err := s.Alg.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"alg\"")
			}
		case "kid":
			if err := func() error {
				s.Kid.Reset()
				if err := s.Kid.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"kid\"")
			}
		case "x5u":
			if err := func() error {
				s.X5u.Reset()
				if err := s.X5u.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"x5u\"")
			}
		case "x5c":
			if err := func() error {
				s.X5c.Reset()
				if err := s.X5c.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"x5c\"")
			}
		case "x5t":
			if err := func() error {
				s.X5t.Reset()
				if err := s.X5t.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"x5t\"")
			}
		case "n":
			if err := func() error {
				s.N.Reset()
				if err := s.N.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"n\"")
			}
		case "e":
			if err := func() error {
				s.E.Reset()
				if err := s.E.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"e\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode JWKResponse")
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *JWKResponse) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *JWKResponse) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *JWKSetResponse) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *JWKSetResponse) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("keys")
		e.ArrStart()
		for _, elem := range s.Keys {
			elem.Encode(e)
		}
		e.ArrEnd()
	}
}

var jsonFieldsNameOfJWKSetResponse = [1]string{
	0: "keys",
}

// Decode decodes JWKSetResponse from json.
func (s *JWKSetResponse) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode JWKSetResponse to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "keys":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				s.Keys = make([]JWKResponse, 0)
				if err := d.Arr(func(d *jx.Decoder) error {
					var elem JWKResponse
					if err := elem.Decode(d); err != nil {
						return err
					}
					s.Keys = append(s.Keys, elem)
					return nil
				}); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"keys\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode JWKSetResponse")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000001,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfJWKSetResponse) {
					name = jsonFieldsNameOfJWKSetResponse[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *JWKSetResponse) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *JWKSetResponse) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *LoginTokens) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *LoginTokens) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("access_token")
		e.Str(s.AccessToken)
	}
	{
		e.FieldStart("token_type")
		e.Str(s.TokenType)
	}
	{
		e.FieldStart("expires_in")
		e.Float64(s.ExpiresIn)
	}
	{
		e.FieldStart("id_token")
		e.Str(s.IDToken)
	}
	{
		e.FieldStart("refresh_token")
		e.Str(s.RefreshToken)
	}
}

var jsonFieldsNameOfLoginTokens = [5]string{
	0: "access_token",
	1: "token_type",
	2: "expires_in",
	3: "id_token",
	4: "refresh_token",
}

// Decode decodes LoginTokens from json.
func (s *LoginTokens) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode LoginTokens to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "access_token":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				v, err := d.Str()
				s.AccessToken = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"access_token\"")
			}
		case "token_type":
			requiredBitSet[0] |= 1 << 1
			if err := func() error {
				v, err := d.Str()
				s.TokenType = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"token_type\"")
			}
		case "expires_in":
			requiredBitSet[0] |= 1 << 2
			if err := func() error {
				v, err := d.Float64()
				s.ExpiresIn = float64(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"expires_in\"")
			}
		case "id_token":
			requiredBitSet[0] |= 1 << 3
			if err := func() error {
				v, err := d.Str()
				s.IDToken = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"id_token\"")
			}
		case "refresh_token":
			requiredBitSet[0] |= 1 << 4
			if err := func() error {
				v, err := d.Str()
				s.RefreshToken = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"refresh_token\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode LoginTokens")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00011111,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfLoginTokens) {
					name = jsonFieldsNameOfLoginTokens[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *LoginTokens) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *LoginTokens) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *OpenIDProviderMetadataResponse) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *OpenIDProviderMetadataResponse) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("issuer")
		e.Str(s.Issuer)
	}
	{
		e.FieldStart("authorization_endpoint")
		e.Str(s.AuthorizationEndpoint)
	}
	{
		e.FieldStart("token_endpoint")
		e.Str(s.TokenEndpoint)
	}
	{
		e.FieldStart("jwks_uri")
		e.Str(s.JwksURI)
	}
}

var jsonFieldsNameOfOpenIDProviderMetadataResponse = [4]string{
	0: "issuer",
	1: "authorization_endpoint",
	2: "token_endpoint",
	3: "jwks_uri",
}

// Decode decodes OpenIDProviderMetadataResponse from json.
func (s *OpenIDProviderMetadataResponse) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode OpenIDProviderMetadataResponse to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "issuer":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				v, err := d.Str()
				s.Issuer = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"issuer\"")
			}
		case "authorization_endpoint":
			requiredBitSet[0] |= 1 << 1
			if err := func() error {
				v, err := d.Str()
				s.AuthorizationEndpoint = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"authorization_endpoint\"")
			}
		case "token_endpoint":
			requiredBitSet[0] |= 1 << 2
			if err := func() error {
				v, err := d.Str()
				s.TokenEndpoint = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"token_endpoint\"")
			}
		case "jwks_uri":
			requiredBitSet[0] |= 1 << 3
			if err := func() error {
				v, err := d.Str()
				s.JwksURI = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"jwks_uri\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode OpenIDProviderMetadataResponse")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00001111,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfOpenIDProviderMetadataResponse) {
					name = jsonFieldsNameOfOpenIDProviderMetadataResponse[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *OpenIDProviderMetadataResponse) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *OpenIDProviderMetadataResponse) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes string as json.
func (o OptString) Encode(e *jx.Encoder) {
	if !o.Set {
		return
	}
	e.Str(string(o.Value))
}

// Decode decodes string from json.
func (o *OptString) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptString to nil")
	}
	o.Set = true
	v, err := d.Str()
	if err != nil {
		return err
	}
	o.Value = string(v)
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s OptString) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *OptString) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes TokenPostApplicationJSON as json.
func (s *TokenPostApplicationJSON) Encode(e *jx.Encoder) {
	unwrapped := (*TokenRequestBody)(s)

	unwrapped.Encode(e)
}

// Decode decodes TokenPostApplicationJSON from json.
func (s *TokenPostApplicationJSON) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode TokenPostApplicationJSON to nil")
	}
	var unwrapped TokenRequestBody
	if err := func() error {
		if err := unwrapped.Decode(d); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return errors.Wrap(err, "alias")
	}
	*s = TokenPostApplicationJSON(unwrapped)
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *TokenPostApplicationJSON) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *TokenPostApplicationJSON) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes TokenPostApplicationXWwwFormUrlencoded as json.
func (s *TokenPostApplicationXWwwFormUrlencoded) Encode(e *jx.Encoder) {
	unwrapped := (*TokenRequestBody)(s)

	unwrapped.Encode(e)
}

// Decode decodes TokenPostApplicationXWwwFormUrlencoded from json.
func (s *TokenPostApplicationXWwwFormUrlencoded) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode TokenPostApplicationXWwwFormUrlencoded to nil")
	}
	var unwrapped TokenRequestBody
	if err := func() error {
		if err := unwrapped.Decode(d); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return errors.Wrap(err, "alias")
	}
	*s = TokenPostApplicationXWwwFormUrlencoded(unwrapped)
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *TokenPostApplicationXWwwFormUrlencoded) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *TokenPostApplicationXWwwFormUrlencoded) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes TokenPostBadRequestApplicationJSON as json.
func (s TokenPostBadRequestApplicationJSON) Encode(e *jx.Encoder) {
	unwrapped := string(s)

	e.Str(unwrapped)
}

// Decode decodes TokenPostBadRequestApplicationJSON from json.
func (s *TokenPostBadRequestApplicationJSON) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode TokenPostBadRequestApplicationJSON to nil")
	}
	var unwrapped string
	if err := func() error {
		v, err := d.Str()
		unwrapped = string(v)
		if err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return errors.Wrap(err, "alias")
	}
	*s = TokenPostBadRequestApplicationJSON(unwrapped)
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s TokenPostBadRequestApplicationJSON) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *TokenPostBadRequestApplicationJSON) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *TokenRequestBody) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *TokenRequestBody) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("code")
		e.Str(s.Code)
	}
	{
		if s.GrantType.Set {
			e.FieldStart("grant_type")
			s.GrantType.Encode(e)
		}
	}
	{
		if s.ClientID.Set {
			e.FieldStart("client_id")
			s.ClientID.Encode(e)
		}
	}
	{
		if s.RedirectURI.Set {
			e.FieldStart("redirect_uri")
			s.RedirectURI.Encode(e)
		}
	}
}

var jsonFieldsNameOfTokenRequestBody = [4]string{
	0: "code",
	1: "grant_type",
	2: "client_id",
	3: "redirect_uri",
}

// Decode decodes TokenRequestBody from json.
func (s *TokenRequestBody) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode TokenRequestBody to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "code":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				v, err := d.Str()
				s.Code = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"code\"")
			}
		case "grant_type":
			if err := func() error {
				s.GrantType.Reset()
				if err := s.GrantType.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"grant_type\"")
			}
		case "client_id":
			if err := func() error {
				s.ClientID.Reset()
				if err := s.ClientID.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"client_id\"")
			}
		case "redirect_uri":
			if err := func() error {
				s.RedirectURI.Reset()
				if err := s.RedirectURI.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"redirect_uri\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode TokenRequestBody")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000001,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfTokenRequestBody) {
					name = jsonFieldsNameOfTokenRequestBody[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *TokenRequestBody) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *TokenRequestBody) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}
