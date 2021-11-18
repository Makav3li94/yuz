package wallet

import (
	"github.com/mr-tron/base58"
	"log"
)
//Base58Encode is a part of algorithm for creating address (3)
//base58 is an algorithm invited by bitcoin
//its like base64 but some characters removed. (O 0 1 I) for not confusing !
func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)

	return []byte(encode)
}
//Base58Decode is a part of verifying address and transactions algorithm
//to get checksum again
func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))

	if err != nil {
		log.Panic(err)
	}
	return decode
}
