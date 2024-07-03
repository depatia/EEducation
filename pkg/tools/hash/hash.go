package hash

import "golang.org/x/crypto/bcrypt"

func HashPass(pass string) ([]byte, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 5)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func CheckPass(pass string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(pass))

	return err == nil
}
