package axiom

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
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
	go axiomServer.RunServer()

	// Run concurrently with some sanity
	wg := sync.WaitGroup{}
	sem := make(chan bool, 200)

	// Loop through our test data
	ls, err := ioutil.ReadDir("../json/")
	assert.Nil(t, err)
	for _, f := range ls {
		var k esi.GetKillmailsKillmailIdKillmailHashOk
		j, err := ioutil.ReadFile("../json/" + f.Name())
		assert.Nil(t, err)
		json.Unmarshal(j, &k)
		sem <- true
		go func(k esi.GetKillmailsKillmailIdKillmailHashOk, f os.FileInfo) {
			// Deal with concurrency limits
			wg.Add(1)
			defer func() { <-sem; wg.Done() }()

			// Do the call
			ret, err := callHTTP(&k)
			if err != nil && strings.Contains(err.Error(), "Abyssal") {
				os.Remove("../json/" + f.Name())
				log.Println("removing abyssal fitted ../json/" + f.Name())
			} else {
				assert.Nil(t, err)
			}
			//fmt.Println(ret)
			assert.NotEqual(t, ret, "")
		}(k, f)
	}
	wg.Wait()
}

var client http.Client

func callHTTP(km *esi.GetKillmailsKillmailIdKillmailHashOk) (string, error) {
	json, err := json.Marshal(km)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "http://localhost:3005/killmail", bytes.NewBuffer(json))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
