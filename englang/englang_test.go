package englang

import (
	"fmt"
	"testing"
)

func TestEnglang(t *testing.T) {
	var begin string
	var end string
	var var1 string
	var var2 string
	err := ScanfContains("Hello. this is a good text for Moose. Bye!", "this is a %s text for %s.", &begin, &var1, &var2, &end)
	fmt.Println(begin)
	fmt.Println(var1)
	fmt.Println(var2)
	fmt.Println(end)
	if err != nil {
		t.Error(err)
	}
}

func TestEnglangFail(t *testing.T) {
	var begin string
	var end string
	var var1 string
	var var2 string
	err := Scanf("Hello. this is a good text for Moose! Bye!", "this is a %s text for %s.", &begin, &var1, &var2, &end)
	fmt.Println(begin)
	fmt.Println(var1)
	fmt.Println(var2)
	fmt.Println(end)
	if err == nil {
		t.Error("expected failure")
	}
}

func TestEnglangFullMatch(t *testing.T) {
	var var1 string
	var var2 string
	err := Scanf("this is a good text for Moose.", "this is a %s text for %s.", &var1, &var2)
	fmt.Println(var1)
	fmt.Println(var2)
	if err != nil {
		t.Error(err)
	}
}

func TestEnglangEvaluate(t *testing.T) {
	if Evaluate("10.1 multiplied by USD 12") != "USD 121.2" {
		t.Error("evaluation error" + Evaluate("10.1 multiplied by USD 12"))
	}
}
