# Platform SDK

A comprehensive TypeScript/JavaScript client library for integrating with our platform's features, including file handling, video chat, channel management, and bot integration.

## Installation

### NPM
```bash
npm install platform-sdk
```

### Yarn
```bash
yarn add platform-sdk
```

## Quick Start

```typescript
import { PlatformSDK } from 'platform-sdk';

// Initialize the SDK
const platform = new PlatformSDK({
    baseURL: 'https://api.your-platform.com',
    apiKey: 'your-api-key'
});

// Example: Login and create a channel
async function example() {
    try {
        // Login
        await platform.login('user@example.com', 'password');
        
        // Create a channel
        const channel = await platform.channels.create({
            name: 'My Channel',
            type: 'public',
            description: 'A test channel'
        });
        
        // Join video chat
        await platform.webrtc.joinChannel(channel.id);
        
        // Upload a file
        const file = document.querySelector('input[type="file"]').files[0];
        await platform.files.upload(file, { channelId: channel.id });
        
    } catch (error) {
        console.error('Error:', error);
    }
}
```

## Connecting to Custom Chat Services

One of the key features of our SDK is the ability to connect to your own custom chat service:

```typescript
// Connect to custom chat service
await platform.chat.connectToService('wss://your-chat-service.com/ws', 'auth-token');

// Register handlers for different message types
platform.chat.onMessage('chat_message', (data) => {
    console.log('New message:', data.content);
});

// Send a message through the connected service
platform.chat.sendMessage('chat_message', {
    content: 'Hello world!',
    sender: 'user123'
});

// When finished, disconnect
platform.chat.disconnect();
```

## Features

- **Authentication**: Login, registration, password reset
- **WebRTC Video Chat**: Join channels, screen sharing
- **File Management**: Upload, download, share, and list files
- **Channel Management**: Create, update, delete channels and manage members
- **Bot Integration**: Create, configure, and interact with bots
- **Subscription Management**: Manage subscription plans, check usage, and handle billing
- **Custom Chat Integration**: Connect to your own chat services

## API Documentation

For complete API documentation, see our [Online API Documentation](https://your-platform.com/docs/api).

## TypeScript Support

This SDK is written in TypeScript and provides full type definitions for all APIs.

## Browser Support

The SDK supports all modern browsers:
- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Project Structure

The SDK is organized in a modular way:

```
client/
├── ts/
│   ├── index.ts              - Main entry point
│   ├── types.ts              - Type definitions
│   ├── http-client.ts        - HTTP client for API requests
│   ├── managers/             - Feature-specific managers
│   │   ├── webrtc-manager.ts
│   │   ├── file-manager.ts
│   │   ├── channel-manager.ts
│   │   ├── bot-manager.ts
│   │   ├── subscription-manager.ts
│   │   └── chat-manager.ts
│   └── ...
├── dist/                     - Compiled output
├── package.json
└── ...
```

## Development

### Building the SDK
```bash
npm run build
```

### Running Tests
```bash
npm test
```

### Linting
```bash
npm run lint
```

### Formatting
```bash
npm run format
```

## License

MIT 