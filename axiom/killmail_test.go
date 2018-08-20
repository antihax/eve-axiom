package axiom

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/antihax/goesi/esi"

	"github.com/stretchr/testify/assert"
)

func TestKillmails(t *testing.T) {
	axiom := NewAxiom()

	ls, err := ioutil.ReadDir("../json/")
	assert.Nil(t, err)
	wg := sync.WaitGroup{}
	for _, f := range ls {
		var k esi.GetKillmailsKillmailIdKillmailHashOk
		j, err := ioutil.ReadFile("../json/" + f.Name())
		assert.Nil(t, err)
		json.Unmarshal(j, &k)
		go func(k esi.GetKillmailsKillmailIdKillmailHashOk, f os.FileInfo) {
			wg.Add(1)
			defer wg.Done()
			r, err := axiom.getAttributesFromKillmail(&k)
			if err != nil && strings.Contains(err.Error(), "Abyssal") {
				os.Remove("../json/" + f.Name())
				log.Println("removing abyssal fitted ../json/" + f.Name())
			} else {
				assert.Nil(t, err)
			}

			b, err := json.MarshalIndent(r, " ", "   ")
			ioutil.WriteFile("../pairedjson/"+f.Name(), b, 0644)

			_, err = json.MarshalIndent(r, " ", "   ")
			if err != nil {
				//	log.Fatalf("%+v\n", r)
			}
			//assert.Nil(t, err)

		}(k, f)
	}
	wg.Done()
	wg.Wait()
}
