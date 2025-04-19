package users

import "errors"

var ErrUserExists = errors.New("user already exists")

type UserService struct {
	UserStore UserStore
}

func (obj UserService) AttemptUserRegistration(username, password string) (*OidcUser, error) {
	user, err := obj.UserStore.GetUser(username)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, ErrUserExists
	}

	salt := GenerateSalt()
	encodedPassword, err := EncodePassword(salt, password)
	if err != nil {
		return nil, err
	}
	user = &OidcUser{
		Id:              username,
		Salt:            salt,
		EncodedPassword: encodedPassword,
	}
	err = obj.UserStore.SaveUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (obj UserService) AttemptUserLogin(username, password string) (*OidcUser, error) {
	user, err := obj.UserStore.GetUser(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	match := ComparePassword(user.Salt, password, user.EncodedPassword)
	if match {
		return user, nil
	}
	return nil, nil
}
