package anticaptcha

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Constants for the AntiCaptcha API
const (
	apiBaseURL     = "https://api.anti-captcha.com"
	checkInterval  = 2 * time.Second
	defaultTimeout = 60 * time.Second
)

// Default logger for the package
var defaultLogger = log.New(os.Stdout, "AntiCaptcha: ", log.LstdFlags)

// Client represents an AntiCaptcha API client
type Client struct {
	APIKey     string
	HTTPClient *http.Client
	Logger     *log.Logger
}

// NewClient creates a new AntiCaptcha API client with a logger.
// If no logger is provided, it uses the default logger.
func NewClient(apiKey string, logger *log.Logger) *Client {
	if logger == nil {
		logger = defaultLogger
	}

	return &Client{
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: defaultTimeout},
		Logger:     logger,
	}
}

// makeRequest sends a request to the AntiCaptcha API and decodes the response
func (c *Client) makeRequest(ctx context.Context, endpoint string, body interface{}, response interface{}) error {
	// Prepare URL
	u, err := url.Parse(apiBaseURL + endpoint)
	if err != nil {
		c.Logger.Printf("Error parsing URL: %v\n", err)
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Marshal the body to JSON
	b, err := json.Marshal(body)
	if err != nil {
		c.Logger.Printf("Error marshaling request body: %v\n", err)
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create a new HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		c.Logger.Printf("Error creating HTTP request: %v\n", err)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Log the request being sent
	c.Logger.Printf("Sending request to %s with body: %v\n", u.String(), len(string(b)))

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Printf("Request failed: %v\n", err)
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.Logger.Printf("Error closing response body: %v\n", cerr)
		}
	}()

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.Logger.Printf("Received non-2xx status code: %d\n", resp.StatusCode)
		return fmt.Errorf("non-2xx status code: %d", resp.StatusCode)
	}

	// Decode the response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.Logger.Printf("Error decoding response: %v\n", err)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Log the received response
	c.Logger.Printf("Received response: %v\n", response)

	return nil
}

// createTaskImage creates an image-to-text task on the AntiCaptcha API
func (c *Client) createTaskImage(ctx context.Context, imgString string) (float64, error) {
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type": "ImageToTextTask",
			"body": imgString,
		},
	}

	c.Logger.Println("Creating task for image captcha...")

	var response map[string]interface{}
	err := c.makeRequest(ctx, "/createTask", body, &response)
	if err != nil {
		c.Logger.Printf("Failed to create task: %v\n", err)
		return 0, fmt.Errorf("failed to create task: %w", err)
	}

	// Check for API errors
	if errMsg, ok := response["errorId"]; ok && errMsg.(float64) != 0 {
		c.Logger.Printf("API error creating task: %s\n", response["errorDescription"].(string))
		return 0, errors.New(response["errorDescription"].(string))
	}

	// Type assertion to float64
	taskID, ok := response["taskId"].(float64)
	if !ok {
		c.Logger.Println("Failed to retrieve taskId from response")
		return 0, errors.New("failed to retrieve taskId from response")
	}

	c.Logger.Printf("Task created successfully with ID: %f\n", taskID)

	return taskID, nil
}

// getTaskResult checks the result of a given task
func (c *Client) getTaskResult(ctx context.Context, taskID float64) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"taskId":    taskID,
	}

	c.Logger.Printf("Checking result for task ID: %f\n", taskID)

	var response map[string]interface{}
	err := c.makeRequest(ctx, "/getTaskResult", body, &response)
	if err != nil {
		c.Logger.Printf("Failed to get task result: %v\n", err)
		return nil, fmt.Errorf("failed to get task result: %w", err)
	}

	return response, nil
}

// SendImage sends an image captcha to the AntiCaptcha API and waits for the solution
func (c *Client) SendImage(imgString string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// Create the task and get the task ID
	taskID, err := c.createTaskImage(ctx, imgString)
	if err != nil {
		c.Logger.Printf("Error sending image: %v\n", err)
		return "", fmt.Errorf("failed to send image: %w", err)
	}

	// Poll for the task result until it's ready
	for {
		response, err := c.getTaskResult(ctx, taskID)
		if err != nil {
			c.Logger.Printf("Error getting task result: %v\n", err)
			return "", fmt.Errorf("failed to get task result: %w", err)
		}

		if status, ok := response["status"].(string); ok && status == "ready" {
			c.Logger.Printf("Task ID %f is ready with solution.\n", taskID)
			solution, ok := response["solution"].(map[string]interface{})
			if !ok {
				c.Logger.Println("Invalid solution format in response")
				return "", errors.New("invalid solution format in response")
			}

			text, ok := solution["text"].(string)
			if !ok {
				c.Logger.Println("Text not found in solution")
				return "", errors.New("text not found in solution")
			}

			c.Logger.Printf("Captcha solved successfully: %s\n", text)
			return text, nil
		}

		c.Logger.Printf("Task ID %f is still processing...\n", taskID)
		time.Sleep(checkInterval)
	}
}
