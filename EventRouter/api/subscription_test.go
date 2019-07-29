package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/api/models"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func init() {
	log.SetLevel(log.D)
}

func TestSubscribe(t *testing.T) {
	server := NewSubscriptionApi(":8070")
	server.Start()

	time.Sleep(1 * time.Second)

	subscription := &models.Subscription{
		Callback: "url",
	}

	// write
	content, err := json.Marshal(subscription)
	if err != nil {
		t.Fatalf("marsheling failed: %v", err)
	}

	fmt.Printf("request: %s\n", content)

	response, err := http.Post("http://localhost:8070/subscribe", "application/json", bytes.NewBuffer(content))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	fmt.Printf("response: %s\n", body)

	var result models.Subscription
	err = json.Unmarshal(body, &result)
	if err != nil {
		t.Fatalf("unmarsheling failed: %v", err)
	}

	assert.Equal(t, subscription.Callback, result.Callback)
	assert.NotNil(t, result.Id)
}

// test invalid request
// test parsing error
// test method not allowed
// test url not found
