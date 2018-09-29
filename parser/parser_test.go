package parser_test

import (
	"testing"

	. "github.com/Eric-GreenComb/parse-eth-balance/parser"
)

func TestTest(t *testing.T) {
	resp, err := Call("bla", nil)
	if err != nil {
		t.Errorf("Error - %v", err)
	}
	t.Logf("Resp - %v", resp)
	//t.Fail()
}
