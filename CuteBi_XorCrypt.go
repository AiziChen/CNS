// CuteBi_XorCrypt.go
package main

import (
	"encoding/base64"
	"errors"
)

var CuteBi_XorCrypt_password []byte = nil

/* 一个简单的异或加密 */
func CuteBi_XorCrypt(data []byte, passwordSub int) int {
	dataLen := len(data)
	if dataLen <= 0 {
		return passwordSub
	}
	passLen := len(CuteBi_XorCrypt_password)
	pi := passwordSub
	for dataSub := 0; dataSub < dataLen; dataSub++ {
		pi = (dataSub + passwordSub) % passLen
		data[dataSub] ^= CuteBi_XorCrypt_password[pi] | byte(pi)
	}

	return pi + 1
}

func CuteBi_decrypt_host(host []byte) ([]byte, error) {
	hostDec := make([]byte, len(host))
	n, err := base64.StdEncoding.Decode(hostDec, host)
	if err != nil {
		return nil, err
	}
	CuteBi_XorCrypt(hostDec, 0)
	if hostDec[n-1] != 0 {
		return nil, errors.New("host decrypt failed")
	}

	return hostDec[:n-1], nil
}
