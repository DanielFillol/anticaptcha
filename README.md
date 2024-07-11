# AntiCaptcha Go Client Library

Welcome to the AntiCaptcha Go Client Library! This library provides an easy-to-use interface to interact with the AntiCaptcha API, enabling you to solve image-based CAPTCHAs and HCaptcha challenges programmatically.

## Features

- Create CAPTCHA tasks: Initiate tasks to solve image CAPTCHAs.
- Poll for results: Check and retrieve the solution for a CAPTCHA task.
- Configurable logging: Log activities and errors using a custom or default logger.
- HTTP client integration: Leverage Go's http.Client for request handling with configurable timeouts.

## Installation

To use this library, ensure you have Go installed. Then, you can get the package via:

```go
go get github.com/DanielFillol/anticaptcha
```

## Usage
Import the package
```go
import "github.com/DanielFillol/anticaptcha"
```

## Creating a Client
To interact with the AntiCaptcha API, create a Client instance. You'll need an API key from AntiCaptcha.

```go
package main

import (
    "log"
    "github.com/DanielFillol/anticaptcha"
)

func main() {
    apiKey := "your_api_key_here"
    logger := log.New(os.Stdout, "AntiCaptcha: ", log.LstdFlags) // optional custom logger
    client := anticaptcha.NewClient(apiKey, logger)

    // Use the client to send images and get results...
}
```
## Sending an Image CAPTCHA
To send an image CAPTCHA to the AntiCaptcha service and get the solution:
```go
import (
    "fmt"
    "log"
    "os"
    "github.com/DanielFillol/anticaptcha"
)

func main() {
    apiKey := "your_api_key_here"
    client := anticaptcha.NewClient(apiKey, nil) // Using default logger

    imgString := "base64_encoded_image_data_here"
    solution, err := client.SendImage(imgString)
    if err != nil {
        log.Fatalf("Failed to solve CAPTCHA: %v", err)
    }

    fmt.Printf("CAPTCHA Solution: %s\n", solution)
}
```
## Sending an Image CAPTCHA
To send an HCaptcha challenge to the AntiCaptcha service and get the solution:
```go
package main

import (
    "fmt"
    "log"
    "github.com/DanielFillol/anticaptcha"
)

func main() {
    apiKey := "your_api_key_here"
    client := anticaptcha.NewClient(apiKey, nil) // Using default logger

    hCaptcha := anticaptcha.NewHCaptchaProxyless(client)
    hCaptcha.SetWebsiteURL("https://website.com")
    hCaptcha.SetWebsiteKey("SITE_KEY")
    hCaptcha.SetIsInvisible(true)  // Optional: Set if HCaptcha is invisible
    hCaptcha.SetIsEnterprise(true) // Optional: Set if HCaptcha is enterprise
    hCaptcha.SetEnterprisePayload(map[string]interface{}{
        "rqdata": "rq data value from target website",
        "sentry": true,
    }) // Optional: Set additional enterprise payload
    hCaptcha.SetSoftID(0) // Optional: Set SoftID

    gResponse, err := hCaptcha.SolveAndReturnSolution()
    if err != nil {
        log.Fatalf("Failed to solve HCaptcha: %v", err)
    }

    fmt.Printf("g-response: %s\n", gResponse)
    fmt.Printf("user-agent: %s\n", hCaptcha.UserAgent)
    fmt.Printf("respkey: %s\n", hCaptcha.RespKey)
}

```

## Polling for Task Results
The SendImage and SolveAndReturnSolution methods automatically handle polling for the task result. However, if you want to manually poll for results:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"github.com/DanielFillol/anticaptcha"
)

func main() {
	apiKey := "your_api_key_here"
	client := anticaptcha.NewClient(apiKey, nil) // Using default logger

	imgString := "base64_encoded_image_data_here"
	ctx, cancel := context.WithTimeout(context.Background(), anticaptcha.DefaultTimeout)
	defer cancel()

	taskID, err := client.CreateTaskImage(ctx, imgString)
	if err != nil {
		log.Fatalf("Failed to create task: %v", err)
	}

	for {
		result, err := client.GetTaskResult(ctx, taskID)
		if err != nil {
			log.Printf("Error checking task result: %v", err)
			continue
		}

		if result["status"] == "ready" {
			solution := result["solution"].(map[string]interface{})["text"].(string)
			fmt.Printf("CAPTCHA Solved: %s\n", solution)
			break
		}

		log.Println("Waiting for solution...")
		time.Sleep(anticaptcha.CheckInterval)
	}
}

```
## Logging
The client supports logging to help you track API requests and responses. You can either use the default logger or provide your own. Log messages include details about requests, responses, and errors.

## Custom Logger
To use a custom logger, pass a *log.Logger instance when creating the client:
```go
package main

import (
	"log"
	"os"
	"github.com/DanielFillol/anticaptcha"
)

func main() {
	apiKey := "your_api_key_here"
	customLogger := log.New(os.Stdout, "CustomAntiCaptcha: ", log.LstdFlags)
	client := anticaptcha.NewClient(apiKey, customLogger)

	// Use the client with the custom logger...
}
```
If you pass nil, the default logger is used.

## Error Handling
The library returns detailed error messages to help you debug issues with API requests or responses. Ensure you handle these errors appropriately in your application.

## Configuration
### Constants
- apiBaseURL: The base URL for the AntiCaptcha API.
- checkInterval: The interval between checks when polling for task results.
- defaultTimeout: The default timeout for HTTP requests.
These constants can be adjusted as per your requirements.

## Contributing
We welcome contributions to improve this library. Feel free to submit issues or pull requests on the GitHub repository.

Feel free to explore and integrate the AntiCaptcha Go Client Library into your projects. If you encounter any issues or have suggestions, don't hesitate to contribute or reach out!

For more details on the AntiCaptcha API, refer to the official [documentation](https://anti-captcha.com/pt/apidoc).
