package notification

import (
	"context"
	"encoding/json"
	"fmt"
	novugo "github.com/novuhq/novu-go"
	"github.com/novuhq/novu-go/models/components"
	"log"
	"net/http"
	"strings"
)

type NovuClient struct {
	Client *novugo.Novu
	APIKey string
	APIURL string // Ví dụ: "https://api.novu.co/v1"
}

func NewNovuClient(apiKey string) *NovuClient {
	client := novugo.New(
		novugo.WithSecurity(apiKey),
	)
	return &NovuClient{
		Client: client,
		APIKey: apiKey,
		APIURL: "https://api.novu.co/v1", // Có thể cấu hình trong .env
	}
}

func (n *NovuClient) SendNotification(subscriberID uint, workflowID, message string) error {
	ctx := context.Background()
	to := components.CreateToStr(fmt.Sprintf("%d", subscriberID))
	payload := map[string]interface{}{"message": message}
	triggerRequest := components.TriggerEventRequestDto{
		WorkflowID: workflowID,
		To:         to,
		Payload:    payload,
	}

	_, err := n.Client.Trigger(ctx, triggerRequest, nil)
	if err != nil {
		log.Printf("Error sending notification: %v", err)
		return err
	}
	return nil
}

func (n *NovuClient) PublishToTopic(topicKey, workflowID, message string) error {
	url := fmt.Sprintf("%s/events/trigger", n.APIURL)
	payload := map[string]interface{}{
		"name":     workflowID,
		"topicKey": topicKey,
		"payload": map[string]interface{}{
			"message": message,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("ApiKey %s", n.APIKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request to Novu: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("Unexpected status code: %d", resp.StatusCode)
		return fmt.Errorf("failed to publish to topic, status: %d", resp.StatusCode)
	}

	return nil
}
