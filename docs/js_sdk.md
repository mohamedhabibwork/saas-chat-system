# JavaScript SDK Documentation

## Overview
The Platform SDK for JavaScript provides a comprehensive client library for integrating with our platform's features, including file handling, video chat, channel management, and bot integration.

## Installation

### NPM
```bash
npm install platform-sdk
```

### CDN
```html
<script src="https://cdn.your-platform.com/sdk.js"></script>
```

## Quick Start

```javascript
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

## Authentication

### Login
```javascript
const response = await platform.login(email, password);
```

### Register
```javascript
const response = await platform.register({
    username: 'john_doe',
    email: 'john@example.com',
    password: 'secure_password'
});
```

### Reset Password
```javascript
// Request password reset
await platform.resetPassword('user@example.com');

// Verify reset token
await platform.verifyResetToken(token, newPassword);
```

## WebRTC Video Chat

### Join a Channel
```javascript
const peerConnection = await platform.webrtc.joinChannel(channelId);
```

### Screen Sharing
```javascript
const stream = await platform.webrtc.startScreenShare(channelId);
```

### Leave Channel
```javascript
await platform.webrtc.leaveChannel(channelId);
```

## File Management

### Upload File
```javascript
const file = document.querySelector('input[type="file"]').files[0];
const response = await platform.files.upload(file, {
    channelId: 'optional_channel_id'
});
```

### Download File
```javascript
const blob = await platform.files.download(fileId);
const url = URL.createObjectURL(blob);
```

### Delete File
```javascript
await platform.files.delete(fileId);
```

### List Files
```javascript
const files = await platform.files.list({
    channelId: 'optional_channel_id',
    limit: 10,
    offset: 0
});
```

### Share File
```javascript
await platform.files.share(fileId, ['user1@example.com', 'user2@example.com']);
```

## Channel Management

### Create Channel
```javascript
const channel = await platform.channels.create({
    name: 'Channel Name',
    type: 'public',
    description: 'Channel description'
});
```

### Get Channel
```javascript
const channel = await platform.channels.get(channelId);
```

### List Channels
```javascript
const channels = await platform.channels.list();
```

### Update Channel
```javascript
await platform.channels.update(channelId, {
    name: 'New Name',
    description: 'New description'
});
```

### Delete Channel
```javascript
await platform.channels.delete(channelId);
```

### Channel Members
```javascript
// Add member
await platform.channels.addMember(channelId, userId);

// Remove member
await platform.channels.removeMember(channelId, userId);
```

### Channel Messages
```javascript
// Send message
await platform.channels.sendMessage(channelId, 'Hello, world!');

// Get messages
const messages = await platform.channels.getMessages(channelId, {
    limit: 50,
    before: timestamp
});
```

## Bot Integration

### Create Bot
```javascript
const bot = await platform.bots.create({
    name: 'Support Bot',
    type: 'customer_support',
    config: {
        language: 'en',
        capabilities: ['chat', 'file_handling']
    }
});
```

### Get Bot
```javascript
const bot = await platform.bots.get(botId);
```

### List Bots
```javascript
const bots = await platform.bots.list();
```

### Update Bot
```javascript
await platform.bots.update(botId, {
    name: 'New Name',
    config: {
        language: 'es'
    }
});
```

### Delete Bot
```javascript
await platform.bots.delete(botId);
```

### Send Message to Bot
```javascript
const response = await platform.bots.sendMessage(botId, 'Hello, bot!');
```

## Subscription Management

### Get Current Plan
```javascript
const plan = await platform.subscription.getCurrentPlan();
```

### Get Usage
```javascript
const usage = await platform.subscription.getUsage();
```

### Upgrade Plan
```javascript
await platform.subscription.upgrade('pro_plan_id');
```

### Cancel Subscription
```javascript
await platform.subscription.cancel();
```

### Get Invoices
```javascript
const invoices = await platform.subscription.getInvoices();
```

## WebSocket Integration

### Connect to WebSocket
```javascript
const ws = platform.connectWebSocket();

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    switch (message.type) {
        case 'chat_message':
            handleChatMessage(message);
            break;
        case 'file_upload':
            handleFileUpload(message);
            break;
        case 'webrtc_signal':
            handleWebRTCSignal(message);
            break;
    }
};
```

## Error Handling

The SDK throws errors for various scenarios. Always wrap SDK calls in try-catch blocks:

```javascript
try {
    await platform.channels.create({
        name: 'Test Channel'
    });
} catch (error) {
    console.error('Error creating channel:', error.message);
    // Handle error appropriately
}
```

## Best Practices

1. **Error Handling**
   - Always use try-catch blocks
   - Implement proper error recovery
   - Show user-friendly error messages

2. **Authentication**
   - Store API keys securely
   - Implement token refresh mechanism
   - Handle session expiration

3. **WebRTC**
   - Check browser compatibility
   - Handle media permissions
   - Implement fallback mechanisms

4. **File Handling**
   - Validate file types and sizes
   - Show upload progress
   - Handle network interruptions

5. **Real-time Features**
   - Implement reconnection logic
   - Handle WebSocket disconnections
   - Cache messages when offline

## Browser Support

The SDK supports all modern browsers:
- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Security Considerations

1. **API Keys**
   - Never expose API keys in client-side code
   - Use environment variables
   - Implement proper key rotation

2. **Data Validation**
   - Validate all user inputs
   - Sanitize file uploads
   - Implement rate limiting

3. **WebRTC Security**
   - Use secure WebRTC connections
   - Implement proper authentication
   - Handle media permissions

## Troubleshooting

Common issues and solutions:

1. **Connection Issues**
   - Check network connectivity
   - Verify API endpoint
   - Check CORS settings

2. **Authentication Errors**
   - Verify API key
   - Check token expiration
   - Validate credentials

3. **WebRTC Problems**
   - Check browser permissions
   - Verify STUN/TURN servers
   - Check firewall settings

4. **File Upload Issues**
   - Check file size limits
   - Verify file types
   - Check network stability

## Support

For additional support:
- Documentation: https://docs.your-platform.com
- API Reference: https://api.your-platform.com/docs
- Support Email: support@your-platform.com 