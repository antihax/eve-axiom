package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"sync"
	"time"

	"github.com/antihax/goesi"
	"github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"
)

var (
	attrChan   chan esi.GetDogmaAttributesAttributeIdOk
	effectChan chan esi.GetDogmaEffectsEffectIdOk
	typeChan   chan esi.GetUniverseTypesTypeIdOk
	esiCli     *goesi.APIClient
	wg         sync.WaitGroup
	ag         sync.WaitGroup
	tg         sync.WaitGroup
	eg         sync.WaitGroup
)

func main() {
	fmt.Printf("package dogma\n\n")
	attrChan = make(chan esi.GetDogmaAttributesAttributeIdOk, 5000)
	typeChan = make(chan esi.GetUniverseTypesTypeIdOk, 50000)
	effectChan = make(chan esi.GetDogmaEffectsEffectIdOk, 50000)

	esiCli = goesi.NewAPIClient(
		&http.Client{
			Transport: &ApiTransport{
				next: &http.Transport{
					MaxIdleConns: 200,
					DialContext: (&net.Dialer{
						Timeout:   300 * time.Second,
						KeepAlive: 5 * 60 * time.Second,
						DualStack: true,
					}).DialContext,
					IdleConnTimeout:       5 * 60 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ResponseHeaderTimeout: 60 * time.Second,
					ExpectContinueTimeout: 0,
					MaxIdleConnsPerHost:   20,
				},
			},
		},
		"eve-axiom: shhh it's almost over",
	)

	effects, _, err := esiCli.ESI.DogmaApi.GetDogmaEffects(context.Background(), nil)
	if err != nil {
		log.Fatalln(err)
	}

	go collectAllEffects()
	for _, a := range effects {
		wg.Add(1)
		go getEffect(a)
	}

	attributes, _, err := esiCli.ESI.DogmaApi.GetDogmaAttributes(context.Background(), nil)
	if err != nil {
		log.Fatalln(err)
	}

	go collectAllAttributes()
	for _, a := range attributes {
		wg.Add(1)
		go getAttribute(a)
	}

	page := int32(1)
	types := []int32{}
	for {
		t, r, err := esiCli.ESI.UniverseApi.GetUniverseTypes(context.Background(),
			&esi.GetUniverseTypesOpts{
				Page: optional.NewInt32(page),
			})
		if err != nil {
			log.Fatalln(err)
		}
		types = append(types, t...)

		if r.Header.Get("x-pages") == strconv.Itoa(int(page)) {
			break
		}
		page++
	}

	go collectAllTypes()
	for _, a := range types {
		wg.Add(1)
		go getType(a)
	}

	// wait for this to finish
	wg.Wait()

	// were done with this,
	close(attrChan)
	close(typeChan)
	close(effectChan)

	ag.Wait()
	tg.Wait()
	eg.Wait()
}

func collectAllTypes() {
	tg.Add(1)
	types := make(map[int32]string)
	typeAttrs := make(map[int32][]int32)
	for t := range typeChan {
		types[t.TypeId] = t.Name
		for _, att := range t.DogmaAttributes {
			typeAttrs[t.TypeId] = append(typeAttrs[t.TypeId], att.AttributeId)
		}
	}

	// Dump our structure to a go file
	fmt.Printf("var typeMap = %#v\n", types)
	fmt.Printf("var typeAttributeMap = %#v\n", typeAttrs)
	tg.Done()
}

func collectAllEffects() {
	eg.Add(1)
	effects := make(map[int32]string)
	effectAttrs := make(map[int32]bool)
	for t := range effectChan {
		effects[t.EffectId] = t.Name
		effectAttrs[t.EffectId] = t.IsAssistance
	}

	// Dump our structure to a go file
	fmt.Printf("var effectMap = %#v\n", effects)
	fmt.Printf("var effectIsAssistanceMap = %#v\n", effectAttrs)
	eg.Done()
}

func collectAllAttributes() {
	ag.Add(1)
	atts := make(map[int32]string)
	attDisplay := make(map[int32]string)
	for att := range attrChan {
		atts[att.AttributeId] = att.Name
		attDisplay[att.AttributeId] = att.DisplayName
	}

	// Dump our structure to a go file
	fmt.Printf("var attributeMap = %#v\n", atts)
	fmt.Printf("var attributeDisplayMap = %#v\n", attDisplay)
	ag.Done()
}
func getEffect(a int32) {
	defer wg.Done()
	t, _, err := esiCli.ESI.DogmaApi.GetDogmaEffectsEffectId(context.Background(), a, nil)
	if err != nil {
		log.Fatalln(err)
	}

	effectChan <- t
}

func getType(a int32) {
	defer wg.Done()
	t, _, err := esiCli.ESI.UniverseApi.GetUniverseTypesTypeId(context.Background(), a, nil)
	if err != nil {
		log.Fatalln(err)
	}

	typeChan <- t
}

func getAttribute(a int32) {
	defer wg.Done()
	att, _, err := esiCli.ESI.DogmaApi.GetDogmaAttributesAttributeId(context.Background(), a, nil)
	if err != nil {
		log.Fatalln(err)
	}

	attrChan <- att
}

var apiTransportLimiter chan bool

func init() {
	// concurrency limiter
	// 100 concurrent requests should fill 1 connection
	apiTransportLimiter = make(chan bool, 100)
}

// ApiTransport custom transport to chain into the HTTPClient to gather statistics.
type ApiTransport struct {
	next *http.Transport
}

// RoundTrip wraps http.DefaultTransport.RoundTrip to provide stats and handle error rates.
func (t *ApiTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Limit concurrency
	apiTransportLimiter <- true

	// Free the worker
	defer func() { <-apiTransportLimiter }()

	// Loop until success
	tries := 0
	for {

		// Tickup retry counter
		tries++

		// Run the request and time the response
		res, triperr := t.next.RoundTrip(req)

		// We got a response
		if res != nil {

			// Get the ESI error information
			resetS := res.Header.Get("x-esi-error-limit-reset")
			tokensS := res.Header.Get("x-esi-error-limit-remain")

			// Log any errors
			if res.StatusCode >= 400 {
				log.Printf("St: %d Res: %s Tok: %s - %s\n", res.StatusCode, resetS, tokensS, req.URL)
			}

			// If we cannot decode this is likely from another source.
			esiRateLimiter := true
			reset, err := strconv.ParseFloat(resetS, 64)
			if err != nil {
				esiRateLimiter = false
			}
			tokens, err := strconv.ParseFloat(tokensS, 64)
			if err != nil {
				esiRateLimiter = false
			}

			// Backoff
			if res.StatusCode == 420 { // Something went wrong
				time.Sleep(time.Duration(reset) * time.Second)
			} else if esiRateLimiter { // Sleep based on error rate.
				percentRemain := 1 - (tokens / 100)
				duration := reset * percentRemain
				time.Sleep(time.Second * time.Duration(duration))
			} else if !esiRateLimiter { // Not an ESI error
				time.Sleep(time.Second * time.Duration(tries))
			}

			// Get out for "our bad" statuses
			if res.StatusCode >= 400 && res.StatusCode < 420 {
				if res.StatusCode != 403 {
					log.Printf("Giving up %d %s\n", res.StatusCode, req.URL)
				}
				return res, triperr
			}
			if res.StatusCode >= 200 && res.StatusCode < 400 {
				return res, triperr
			}
		}

		if tries > 10 {
			log.Printf("Too many tries\n")
			return res, triperr
		}
	}
}
