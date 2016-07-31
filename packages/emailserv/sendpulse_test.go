// sendpulse_test
package main

import (
	"fmt"
	"testing"
)

func TestSendPulse(t *testing.T) {
	ec := NewEmailClient(`YOUR API_ID`,
		`YOUR API_SECRET`,
		&Email{`FROM NAME`, `FROM EMAIL`})
	/*	err := ec.GetToken()
		if err != nil {
			t.Error( err )
		} else {*/
	err := ec.SendEmail("<p>Тестовое сообщение</p>", "Тестовое сообщение", "Тест",
		[]*Email{
			&Email{`John Doe`, `johndoe@gmail.com`}})
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(`Send OK`)
	}
	//	}
}
