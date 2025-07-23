package deck

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEncyptCard(t *testing.T) {
	key := []byte("foobarbaz")
	card := Card {
		Suit: Spades,
		Value: 1,
	}

	encryptedOutput, err := EncryptCard(key, card)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(encryptedOutput)
	decryptedOutput, err := DecryptCard(key, encryptedOutput)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(decryptedOutput)
	if !reflect.DeepEqual(card, decryptedOutput) {
		t.Errorf("got %+v but want %+v", decryptedOutput, card)
	}
}
