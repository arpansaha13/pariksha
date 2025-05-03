package utils

import (
	"crypto/rand"
	"math/big"
)

func GenerateOTP(length int16) (string, error) {
	digits := "0123456789"
	otp := ""

	var i int16
	for i = 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp += string(digits[num.Int64()])
	}

	return otp, nil
}
