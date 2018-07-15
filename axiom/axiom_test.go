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

func TestAxiom(t *testing.T) {
	// Setup and run the axiom server
	axiomServer := NewAxiom()
	axiomServer.RunServer()

	// Create a new API client
	a, err := NewAxiomAPIClient("localhost:3003")
	assert.Nil(t, err)

	// Loop through our test data
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
			ret, err := a.GetKillmailAttributes(&k)
			if err != nil && strings.Contains(err.Error(), "Abyssal") {
				os.Remove("../json/" + f.Name())
				log.Println("removing abyssal fitted ../json/" + f.Name())
			} else {
				assert.Nil(t, err)
			}
			assert.NotNil(t, ret)
		}(k, f)
	}
	wg.Wait()
}
