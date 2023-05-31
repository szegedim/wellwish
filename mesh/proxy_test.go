package mesh

import (
	"fmt"
	"net/url"
	"testing"
)

func TestName(t *testing.T) {
	x, _ := url.Parse("https://eper.io:5555/tmp?apikey=abcd")
	fmt.Println(x.Hostname(), x.Port())
	fmt.Println(x.RequestURI())
}
