package hash

import "golang.org/x/crypto/bcrypt"

func Hash(text string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckHash(text, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, text)
	return err == nil
}
