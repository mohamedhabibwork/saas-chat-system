# System Integration Guide

## Overview
This guide provides detailed instructions and examples for integrating with our system's various features including file handling, video chat, and bot integration.

## Authentication

### API Authentication
```go
// Example of API authentication
type APIClient struct {
    BaseURL    string
    APIKey     string
    HTTPClient *http.Client
}

func NewAPIClient(baseURL, apiKey string) *APIClient {
    return &APIClient{
        BaseURL:    baseURL,
        APIKey:     apiKey,
        HTTPClient: &http.Client{},
    }
}

func (c *APIClient) Authenticate() error {
    req, err := http.NewRequest("POST", c.BaseURL+"/api/auth", nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.APIKey)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

## File Handling Integration

### Upload File
```go
// Example of file upload
func (c *APIClient) UploadFile(filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil {
        return err
    }

    _, err = io.Copy(part, file)
    if err != nil {
        return err
    }

    writer.Close()

    req, err := http.NewRequest("POST", c.BaseURL+"/api/files/upload", body)
    if err != nil {
        return err
    }

    req.Header.Set("Authorization", "Bearer "+c.APIKey)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

### Download File
```go
// Example of file download
func (c *APIClient) DownloadFile(fileID string, outputPath string) error {
    req, err := http.NewRequest("GET", c.BaseURL+"/api/files/"+fileID, nil)
    if err != nil {
        return err
    }

    req.Header.Set("Authorization", "Bearer "+c.APIKey)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}
```

## Video Chat Integration

### Initialize WebRTC Connection
```javascript
// Example of WebRTC initialization
const webrtc = {
    init: async function(channelId) {
        const peerConnection = new RTCPeerConnection({
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' }
            ]
        });

        // Get user media
        const stream = await navigator.mediaDevices.getUserMedia({
            video: true,
            audio: true
        });

        // Add tracks to peer connection
        stream.getTracks().forEach(track => {
            peerConnection.addTrack(track, stream);
        });

        // Create and send offer
        const offer = await peerConnection.createOffer();
        await peerConnection.setLocalDescription(offer);

        // Send offer to server
        await this.sendOffer(channelId, offer);

        return peerConnection;
    },

    sendOffer: async function(channelId, offer) {
        const response = await fetch('/api/webrtc/offer', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${this.apiKey}`
            },
            body: JSON.stringify({
                channel_id: channelId,
                offer: offer
            })
        });
        return response.json();
    }
};
```

## Bot Integration

### Create Bot
```go
// Example of bot creation
func (c *APIClient) CreateBot(config BotConfig) error {
    data, err := json.Marshal(config)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", c.BaseURL+"/api/bots", bytes.NewBuffer(data))
    if err != nil {
        return err
    }

    req.Header.Set("Authorization", "Bearer "+c.APIKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

### Bot Configuration Example
```json
{
    "name": "Customer Support Bot",
    "description": "Handles customer inquiries",
    "model_type": "gpt-4",
    "model_config": {
        "temperature": 0.7,
        "max_tokens": 150,
        "stop_sequences": ["\n", "Human:", "Assistant:"]
    },
    "capabilities": {
        "file_handling": true,
        "context_awareness": true,
        "multi_language": true
    }
}
```

## WebSocket Integration

### Connect to WebSocket
```javascript
// Example of WebSocket connection
const ws = {
    connect: function() {
        const socket = new WebSocket('ws://your-server/ws');
        
        socket.onopen = () => {
            console.log('Connected to WebSocket');
            this.authenticate(socket);
        };

        socket.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        return socket;
    },

    authenticate: function(socket) {
        socket.send(JSON.stringify({
            type: 'auth',
            token: this.apiKey
        }));
    },

    handleMessage: function(message) {
        switch(message.type) {
            case 'chat_message':
                this.handleChatMessage(message);
                break;
            case 'file_upload':
                this.handleFileUpload(message);
                break;
            case 'webrtc_signal':
                this.handleWebRTCSignal(message);
                break;
        }
    }
};
```

## Error Handling

### Error Response Structure
```json
{
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable error message",
        "details": {
            "field": "Specific field with error",
            "reason": "Detailed error reason"
        }
    }
}
```

### Error Handling Example
```go
func (c *APIClient) handleError(resp *http.Response) error {
    var errorResponse struct {
        Error struct {
            Code    string                 `json:"code"`
            Message string                 `json:"message"`
            Details map[string]interface{} `json:"details"`
        } `json:"error"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
        return fmt.Errorf("failed to decode error response: %v", err)
    }

    return fmt.Errorf("API error: %s - %s", errorResponse.Error.Code, errorResponse.Error.Message)
}
```

## Rate Limiting

### Rate Limit Headers
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1623456789
```

### Rate Limit Handling
```go
func (c *APIClient) checkRateLimit(resp *http.Response) error {
    limit := resp.Header.Get("X-RateLimit-Limit")
    remaining := resp.Header.Get("X-RateLimit-Remaining")
    reset := resp.Header.Get("X-RateLimit-Reset")

    if remaining == "0" {
        resetTime, _ := strconv.ParseInt(reset, 10, 64)
        return fmt.Errorf("rate limit exceeded. Reset at: %v", time.Unix(resetTime, 0))
    }

    return nil
}
```

## Best Practices

### Connection Management
- Use connection pooling
- Implement retry mechanisms
- Handle timeouts appropriately
- Keep connections alive when possible

### Error Handling
- Implement proper error handling
- Log errors appropriately
- Provide meaningful error messages
- Handle network errors gracefully

### Security
- Use HTTPS for all communications
- Implement proper authentication
- Validate all inputs
- Sanitize all outputs

### Performance
- Implement caching where appropriate
- Use compression for large payloads
- Optimize request frequency
- Monitor response times 