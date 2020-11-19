package main

import (
	"testing"
)

func TestSubstr(t *testing.T) {
	testStr := "Bramborové knedlíky kulaté"
	res := Substr(&testStr, 10)
	if testStr[0:10] != res{
		t.Fail()
	}
}
