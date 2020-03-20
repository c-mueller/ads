package ads

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/idna"
	"testing"
)

var values = map[string]string{
	"ɢoogle.com":          "xn--oogle-wmc.com",
	"müller.c-mueller.de": "xn--mller-kva.c-mueller.de",
	"mähl.c-mueller.de":   "xn--mhl-qla.c-mueller.de",
	"💩.krnl.eu":          "xn--ls8h.krnl.eu",
	"c-mueller.de":        "c-mueller.de",
}

func TestIDNADecode(t *testing.T) {
	for k, v := range values {
		result, err := idna.ToASCII(k)
		assert.NoError(t, err)
		t.Log(result)
		assert.Equal(t, v, result)
	}
}
