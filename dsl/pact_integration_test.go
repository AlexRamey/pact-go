package dsl

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"testing"
)

func Test_NativeMockServer(t *testing.T) {
	os.RemoveAll(pactDir)

	matcher := map[string]interface{}{
		"user": map[string]interface{}{
			"phone":     Regex("\\d+", 12345678),
			"name":      Regex("\\s+", "someusername"),
			"address":   Regex("\\s+", "some address"),
			"plaintext": "plaintext",
		},
		"pass": Regex("\\d+", 1234),
	}

	pact := Pact{
		Port:     6666,
		Consumer: "billy",
		Provider: "bobby",
		LogLevel: "DEBUG",
		LogDir:   logDir,
		PactDir:  pactDir,
	}
	defer pact.Teardown()

	pact.AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{
			Path:   "/foobar",
			Method: "POST",
			Body:   PactBodyBuilder(matcher),
		}).
		WillRespondWith(Response{
			Status: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		})

	err := pact.Verify(func() error {
		request := `{
			"pass": 1234,
			"user": {
				"address": "some address",
				"name": "someusername",
				"phone": 12345678,
				"plaintext": "plaintext"
			}
		}`

		http.Post(fmt.Sprintf("http://localhost:%d/foobar", pact.ServerPort), "application/json", bytes.NewReader([]byte(request)))
		return nil
	})

	if err != nil {
		t.Fatal("err: ", err)
	}
	pact.WritePact()
	if err != nil {
		t.Fatal("error:", err)
	}

	pact2 := Pact{
		Port:     6666,
		Consumer: "billy",
		Provider: "bobby",
		LogLevel: "DEBUG",
		LogDir:   logDir,
		PactDir:  pactDir,
	}
	defer pact2.Teardown()

	pact2.AddInteraction().
		Given("A user").
		UponReceiving("A request for user A").
		WithRequest(Request{
			Path:   "/user",
			Method: "GET",
		}).
		WillRespondWith(Response{
			Status: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		})

	err = pact2.Verify(func() error {
		http.Get(fmt.Sprintf("http://localhost:%d/user", pact2.ServerPort))
		return nil
	})

	if err != nil {
		t.Fatal("error:", err)
	}

	err = pact2.WritePact()
	if err != nil {
		t.Fatal("error:", err)
	}

}

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../pacts", dir)
var logDir = fmt.Sprintf("%s/../logs", dir)

// func TestPact_Integration(t *testing.T) {
// 	// Enable when running E2E/integration tests before a release
// 	if os.Getenv("PACT_INTEGRATED_TESTS") != "" {
//
// 		// Setup Provider API for verification (later...)
// 		providerPort := setupProviderAPI()
// 		pactDaemonPort := 6666
//
// 		// Create Pact connecting to local Daemon
// 		pact := Pact{
// 			Port:     pactDaemonPort,
// 			Consumer: "billy",
// 			Provider: "bobby",
// 			LogLevel: "ERROR",
// 			LogDir:   logDir,
// 			PactDir:  pactDir,
// 		}
// 		defer pact.Teardown()
//
// 		// Pass in test case
// 		var test = func() error {
// 			_, err := http.Get(fmt.Sprintf("http://localhost:%d/foobar", pact.Server.Port))
// 			if err != nil {
// 				t.Fatalf("Error sending request: %v", err)
// 			}
// 			_, err = http.Get(fmt.Sprintf("http://localhost:%d/bazbat", pact.Server.Port))
// 			if err != nil {
// 				t.Fatalf("Error sending request: %v", err)
// 			}
//
// 			return err
// 		}
//
// 		// Setup a complex interaction
// 		jumper := Like(`"jumper"`)
// 		shirt := Like(`"shirt"`)
// 		tag := EachLike(fmt.Sprintf(`[%s, %s]`, jumper, shirt), 2)
// 		size := Like(10)
// 		colour := Term("red", "red|green|blue")
//
// 		body :=
// 			formatJSON(
// 				EachLike(
// 					EachLike(
// 						fmt.Sprintf(
// 							`{
// 						"size": %s,
// 						"colour": %s,
// 						"tag": %s
// 					}`, size, colour, tag),
// 						1),
// 					1))
//
// 		// Set up our interactions. Note we have multiple in this test case!
// 		pact.
// 			AddInteraction().
// 			Given("Some state").
// 			UponReceiving("Some name for the test").
// 			WithRequest(Request{
// 				Method: "GET",
// 				Path:   "/foobar",
// 			}).
// 			WillRespondWith(Response{
// 				Status: 200,
// 				Headers: map[string]string{
// 					"Content-Type": "application/json",
// 				},
// 			})
// 		pact.
// 			AddInteraction().
// 			Given("Some state2").
// 			UponReceiving("Some name for the test").
// 			WithRequest(Request{
// 				Method: "GET",
// 				Path:   "/bazbat",
// 			}).
// 			WillRespondWith(Response{
// 				Status: 200,
// 				Body:   body,
// 			})
//
// 		// Verify Collaboration Test interactionns (Consumer sid)
// 		err := pact.Verify(test)
// 		if err != nil {
// 			t.Fatalf("Error on Verify: %v", err)
// 		}
//
// 		// Write pact to file `<pact-go>/pacts/my_consumer-my_provider.json`
// 		pact.WritePact()
//
// 		// Publish the Pacts...
// 		p := Publisher{}
// 		brokerHost := os.Getenv("PACT_BROKER_HOST")
// 		err = p.Publish(types.PublishRequest{
// 			PactURLs:        []string{"../pacts/billy-bobby.json"},
// 			PactBroker:      brokerHost,
// 			ConsumerVersion: "1.0.0",
// 			Tags:            []string{"latest", "sit4"},
// 			BrokerUsername:  os.Getenv("PACT_BROKER_USERNAME"),
// 			BrokerPassword:  os.Getenv("PACT_BROKER_PASSWORD"),
// 		})
//
// 		if err != nil {
// 			t.Fatalf("Error: %v", err)
// 		}
//
// 		// Verify the Provider - local Pact Files
// 		err = pact.VerifyProvider(types.VerifyRequest{
// 			ProviderBaseURL:        fmt.Sprintf("http://localhost:%d", providerPort),
// 			PactURLs:               []string{"./pacts/billy-bobby.json"},
// 			ProviderStatesURL:      fmt.Sprintf("http://localhost:%d/states", providerPort),
// 			ProviderStatesSetupURL: fmt.Sprintf("http://localhost:%d/setup", providerPort),
// 		})
//
// 		if err != nil {
// 			t.Fatal("Error:", err)
// 		}
//
// 		// Verify the Provider - Specific Published Pacts
// 		err = pact.VerifyProvider(types.VerifyRequest{
// 			ProviderBaseURL:        fmt.Sprintf("http://localhost:%d", providerPort),
// 			PactURLs:               []string{fmt.Sprintf("%s/pacts/provider/bobby/consumer/billy/latest/sit4", brokerHost)},
// 			ProviderStatesURL:      fmt.Sprintf("http://localhost:%d/states", providerPort),
// 			ProviderStatesSetupURL: fmt.Sprintf("http://localhost:%d/setup", providerPort),
// 			BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
// 			BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
// 		})
//
// 		if err != nil {
// 			t.Fatal("Error:", err)
// 		}
//
// 		// Verify the Provider - Latest Published Pacts for any known consumers
// 		err = pact.VerifyProvider(types.VerifyRequest{
// 			ProviderBaseURL:        fmt.Sprintf("http://localhost:%d", providerPort),
// 			BrokerURL:              brokerHost,
// 			ProviderStatesURL:      fmt.Sprintf("http://localhost:%d/states", providerPort),
// 			ProviderStatesSetupURL: fmt.Sprintf("http://localhost:%d/setup", providerPort),
// 			BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
// 			BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
// 		})
//
// 		if err != nil {
// 			t.Fatal("Error:", err)
// 		}
//
// 		// Verify the Provider - Tag-based Published Pacts for any known consumers
// 		err = pact.VerifyProvider(types.VerifyRequest{
// 			ProviderBaseURL:        fmt.Sprintf("http://localhost:%d", providerPort),
// 			BrokerURL:              brokerHost,
// 			Tags:                   []string{"latest", "sit4"},
// 			ProviderStatesURL:      fmt.Sprintf("http://localhost:%d/states", providerPort),
// 			ProviderStatesSetupURL: fmt.Sprintf("http://localhost:%d/setup", providerPort),
// 			BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
// 			BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
// 		})
//
// 		if err != nil {
// 			t.Fatal("Error:", err)
// 		}
// 	}
// }
//
// // Used as the Provider in the verification E2E steps
// func setupProviderAPI() int {
// 	port, _ := utils.GetFreePort()
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/setup", func(w http.ResponseWriter, req *http.Request) {
// 		log.Println("[DEBUG] provider API: states setup")
// 		w.Header().Add("Content-Type", "application/json")
// 	})
// 	mux.HandleFunc("/states", func(w http.ResponseWriter, req *http.Request) {
// 		log.Println("[DEBUG] provider API: states")
// 		fmt.Fprintf(w, `{"billy": ["Some state", "Some state2"]}`)
// 		w.Header().Add("Content-Type", "application/json")
// 	})
// 	mux.HandleFunc("/foobar", func(w http.ResponseWriter, req *http.Request) {
// 		log.Println("[DEBUG] provider API: /foobar")
// 		w.Header().Add("Content-Type", "application/json")
// 	})
// 	mux.HandleFunc("/bazbat", func(w http.ResponseWriter, req *http.Request) {
// 		log.Println("[DEBUG] provider API: /bazbat")
// 		w.Header().Add("Content-Type", "application/json")
// 		fmt.Fprintf(w, `
// 			[
// 			  [
// 			    {
// 			      "size": 10,
// 			      "colour": "red",
// 			      "tag": [
// 			        [
// 			          "jumper",
// 			          "shirt"
// 			        ],
// 			        [
// 			          "jumper",
// 			          "shirt"
// 			        ]
// 			      ]
// 			    }
// 			  ]
// 			]`)
// 	})
//
// 	go http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
// 	return port
// }
