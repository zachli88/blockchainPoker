package deck

import (
	"bytes"
	"encoding/gob"
)

func EncryptCard(key []byte, card Card) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(card); err != nil {
		return nil, err
	}
	payload := buf.Bytes()
	
	return Encrypt(key, payload)
}

func DecryptCard(key []byte, encCard []byte) (Card, error) {
	card := Card{}
	decryptedOutput, err := Encrypt(key, encCard)
	if err != nil {
		return card, err
	}
	if err := gob.NewDecoder(bytes.NewReader(decryptedOutput)).Decode(&card); err != nil {
		return card, err
	}
	return card, nil
}

func Encrypt(key, payload []byte) ([]byte, error) {
	encryptedOutput := make([]byte, len(payload))
	for i := 0; i < len(payload); i++ {
		encryptedOutput[i] = payload[i] ^ key[i % len(key)]
	}
	return encryptedOutput, nil
}
