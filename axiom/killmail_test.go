package axiom

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/antihax/goesi/esi"

	"github.com/stretchr/testify/assert"
)

func TestKillmails(t *testing.T) {
	axiom := NewAxiom()

	ls, err := ioutil.ReadDir("../json/")
	assert.Nil(t, err)
	for _, f := range ls {
		var k esi.GetKillmailsKillmailIdKillmailHashOk
		j, err := ioutil.ReadFile("../json/" + f.Name())
		assert.Nil(t, err)
		json.Unmarshal(j, &k)
		err = axiom.GetAttributesFromKillmail(&k)
		assert.Nil(t, err)
	}
}
