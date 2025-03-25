# Encryption Implementation Guide

## Overview

This document provides technical details about the encryption implementation in both the client SDK and server components.

## Client-Side Implementation (TypeScript)

### ChatManager Encryption

The `ChatManager` class handles end-to-end encryption for chat messages using the Web Crypto API.

```typescript
class ChatManager {
  private channelKeys: Map<string, string>;
  private cryptoKeyCache: Map<string, CryptoKey>;
  private encoder: TextEncoder;
  private decoder: TextDecoder;
}
```

### Key Management

1. **Channel Key Storage**
   ```typescript
   // Store channel key received from server
   this.channelKeys.set(channel, channelKey);
   
   // Generate CryptoKey from channel key
   const keyData = this.encoder.encode(channelKey);
   const keyDigest = await crypto.subtle.digest('SHA-256', keyData);
   const cryptoKey = await crypto.subtle.importKey(
     'raw',
     keyDigest,
     { name: 'AES-GCM', length: 256 },
     false,
     ['encrypt', 'decrypt']
   );
   ```

2. **Message Encryption**
   ```typescript
   async encryptData(data: string, channel: string): Promise<string> {
     const cryptoKey = await this.getCryptoKeyForChannel(channel);
     const iv = crypto.getRandomValues(new Uint8Array(12));
     const dataBytes = this.encoder.encode(data);
     const encryptedBytes = await crypto.subtle.encrypt(
       { name: 'AES-GCM', iv },
       cryptoKey,
       dataBytes
     );
     
     // Combine IV and encrypted data
     const result = new Uint8Array(iv.length + new Uint8Array(encryptedBytes).length);
     result.set(iv);
     result.set(new Uint8Array(encryptedBytes), iv.length);
     
     return bufferToBase64(result);
   }
   ```

3. **Message Decryption**
   ```typescript
   async decryptData(encryptedData: string, channel: string): Promise<string> {
     const cryptoKey = await this.getCryptoKeyForChannel(channel);
     const encryptedBytes = base64ToBuffer(encryptedData);
     
     // Extract IV and data
     const iv = encryptedBytes.slice(0, 12);
     const data = encryptedBytes.slice(12);
     
     const decryptedBytes = await crypto.subtle.decrypt(
       { name: 'AES-GCM', iv },
       cryptoKey,
       data
     );
     
     return this.decoder.decode(decryptedBytes);
   }
   ```

## Server-Side Implementation (Go)

### Encryption Service

The `encryption.Service` struct provides encryption functionality for the server:

```go
type Service struct {
    config Config
    key    []byte
}

type Config struct {
    Algorithm    string
    KeyDerivation struct {
        Algorithm   string
        Iterations  int
        SaltLength  int
    }
    IVLength  int
    TagLength int
    KeyLength int
    Salt      string
    Key       string
}
```

### Key Derivation

```go
// Derive encryption key using PBKDF2
key = pbkdf2.Key(
    key,
    salt,
    config.KeyDerivation.Iterations,
    config.KeyLength,
    sha256.New
)
```

### Message Encryption/Decryption

```go
// Encrypt
func (s *Service) Encrypt(plaintext []byte) ([]byte, error) {
    block, _ := aes.NewCipher(s.key)
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    io.ReadFull(rand.Reader, nonce)
    return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt
func (s *Service) Decrypt(ciphertext []byte) ([]byte, error) {
    block, _ := aes.NewCipher(s.key)
    gcm, _ := cipher.NewGCM(block)
    if len(ciphertext) < gcm.NonceSize() {
        return nil, errors.New("ciphertext too short")
    }
    nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
    return gcm.Open(nil, nonce, ciphertext, nil)
}
```

## WebSocket Implementation

### Message Structure

```typescript
interface Message {
  type: string;
  channel: string;
  userId: string;
  timestamp: number;
  payload?: any;
  encryptedData?: string;
  publicKey?: string;
}
```

### Encryption Flow in WebSocket Handler

1. **Client Joins Channel**
   ```go
   func (h *Hub) JoinChannel(client *Client, channel string) {
       // Generate channel key
       channelKey := generateRandomKey(32)
       
       // Send key exchange message
       keyExchangeMsg := &Message{
           Type:    "key_exchange",
           Channel: channel,
           Payload: json.RawMessage(`{"channelKey":"` + channelKey + `"}`),
       }
       client.send <- msgBytes
   }
   ```

2. **Message Broadcasting**
   ```go
   func (h *Hub) broadcast(message *Message) {
       clients := h.channels[message.Channel]
       for client := range clients {
           // Encrypt message if needed
           if message.EncryptedData == "" && client.encryptionKeys[message.Channel] != "" {
               encryptedData, _ := h.encryptionService.EncryptString(string(message.Payload))
               message.EncryptedData = encryptedData
               message.Payload = nil
           }
           client.send <- messageBytes
       }
   }
   ```

## API Encryption Middleware

### Request/Response Encryption

```go
func (m *EncryptionMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip certain endpoints
        if shouldSkipEncryption(r.URL.Path) {
            next.ServeHTTP(w, r)
            return
        }

        // Decrypt request
        if r.Method != http.MethodGet && r.Body != nil {
            var encReq EncryptedRequest
            json.NewDecoder(r.Body).Decode(&encReq)
            decrypted, _ := m.encryptionService.DecryptString(encReq.EncryptedData)
            r.Body = ioutil.NopCloser(bytes.NewReader([]byte(decrypted)))
        }

        // Capture and encrypt response
        rw := newResponseWriter(w)
        next.ServeHTTP(rw, r)
        
        if rw.Status() >= 200 && rw.Status() < 300 {
            encrypted, _ := m.encryptionService.EncryptString(string(rw.Body()))
            json.NewEncoder(w).Encode(EncryptedResponse{EncryptedData: encrypted})
        }
    })
}
```

## Testing

### Unit Tests

1. **Encryption/Decryption**
   ```typescript
   describe('ChatManager encryption', () => {
     it('should encrypt and decrypt messages correctly', async () => {
       const message = 'Hello, World!';
       const encrypted = await chatManager.encryptData(message, 'channel1');
       const decrypted = await chatManager.decryptData(encrypted, 'channel1');
       expect(decrypted).toBe(message);
     });
   });
   ```

2. **Key Exchange**
   ```typescript
   it('should handle key exchange messages', async () => {
     const keyExchangeMsg = {
       type: 'key_exchange',
       channel: 'channel1',
       payload: { channelKey: 'test-key' }
     };
     await chatManager.handleKeyExchange(keyExchangeMsg);
     expect(chatManager.hasChannelKey('channel1')).toBe(true);
   });
   ```

### Integration Tests

```typescript
describe('End-to-end encryption', () => {
  it('should handle encrypted chat messages', async () => {
    const client1 = new ChatManager(sdk);
    const client2 = new ChatManager(sdk);
    
    await client1.joinChannel('test-channel');
    await client2.joinChannel('test-channel');
    
    const message = 'Secret message';
    await client1.sendMessage('test-channel', message);
    
    // Wait for message to be received
    const receivedMsg = await new Promise(resolve => {
      client2.on('message_received', resolve);
    });
    
    expect(receivedMsg.content).toBe(message);
  });
});
```

## Security Considerations

1. **Key Storage**
   - Use secure key storage mechanisms
   - Implement key rotation
   - Protect against key extraction

2. **Error Handling**
   - Graceful degradation
   - Secure error logging
   - User notification

3. **Performance**
   - Cache derived keys
   - Optimize message encryption
   - Handle large messages

4. **Monitoring**
   - Track encryption failures
   - Monitor key usage
   - Alert on anomalies

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2024-03-01 | Initial implementation |
| 1.1.0 | 2024-03-15 | Added key rotation | 