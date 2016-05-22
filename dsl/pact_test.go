package dsl

import (
	"errors"
	"testing"
)

func simplePact() (pact *PactMock) {
	pact = &PactMock{}
	pact.
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{})
	return
}

func providerStatesPact() (pact *PactMock) {
	pact = &PactMock{}
	pact.
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{})
	return
}

func TestPactDSL(t *testing.T) {
	pact := &PactMock{}
	pact.
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{})
}

func TestPactVerify_NoState(t *testing.T) {
	pact := simplePact()
	err := pact.Verify()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestPactVerify_NoState_Fail(t *testing.T) {
	pact := simplePact()
	pact.VerifyResponse = errors.New("Pact failure!")
	err := pact.Verify()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPactVerify_State(t *testing.T) {
	pact := providerStatesPact()
	err := pact.Verify()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestPactVerify_State_Fail(t *testing.T) {
	pact := providerStatesPact()
	pact.VerifyResponse = errors.New("Pact failure!")
	err := pact.Verify()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}
