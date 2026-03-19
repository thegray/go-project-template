package crypto

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct {
	Cost int
}

func NewBcryptHasher(cost int) BcryptHasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return BcryptHasher{Cost: cost}
}

func (h BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.Cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
