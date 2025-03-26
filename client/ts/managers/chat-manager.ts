/**
 * Chat Manager for chat operations and external chat service integration
 */
import { HttpClient } from '../http-client';
import { WebSocketMessage, PaginatedResponse } from '../types';
import { EventEmitter } from 'events';
import { PlatformSDK } from '../index';
import { formatTimestamp, getUserTimezone } from '../utils';
import { ChatEvent } from '../types/events';
import { TrackingManager } from './tracking-manager';

interface ChatMessage {
  id?: string;
  userId: string;
  username?: string;
  channel: string;
  content: string;
  timestamp: number;
  formattedTime?: string;
  type?: string;
  attachments?: any[];
}

interface OnlineUser {
  userId: string;
  username: string;
  status: 'online' | 'offline' | 'away';
  lastActive?: number;
}

interface ChatOptions {
  historyLimit?: number;
  encryption?: boolean;
}

/**
 * ChatManager handles real-time messaging functionality
 */
export class ChatManager extends EventEmitter {
  private sdk: PlatformSDK;
  private socket: WebSocket | null = null;
  private connected = false;
  private channels: Set<string> = new Set();
  private channelKeys: Map<string, string> = new Map();
  private onlineUsers: Map<string, OnlineUser[]> = new Map();
  private messageHistory: Map<string, ChatMessage[]> = new Map();
  private pendingMessages: Map<string, ChatMessage[]> = new Map();
  private options: ChatOptions = {
    historyLimit: 50,
    encryption: true
  };
  private encoder: TextEncoder;
  private decoder: TextDecoder;
  private cryptoKeyCache: Map<string, CryptoKey> = new Map();
  private tracking: TrackingManager;

  /**
   * Creates a new ChatManager instance
   * @param sdk - The PlatformSDK instance
   */
  constructor(sdk: PlatformSDK, options?: ChatOptions) {
    super();
    this.sdk = sdk;
    this.tracking = new TrackingManager(sdk);
    
    if (options) {
      this.options = { ...this.options, ...options };
    }
    
    // Initialize text encoder/decoder for encryption
    this.encoder = new TextEncoder();
    this.decoder = new TextDecoder();
  }

  /**
   * Connects to the WebSocket server
   */
  async connect(): Promise<void> {
    if (this.connected) {
      return;
    }

    const startTime = Date.now();
    try {
      const token = await this.sdk.auth.getToken();
      const baseUrl = this.sdk.config.wsUrl || this.sdk.config.apiUrl.replace(/^http/, 'ws');
      const wsUrl = `${baseUrl}/ws?token=${token}`;

      return new Promise((resolve, reject) => {
        this.socket = new WebSocket(wsUrl);

        this.socket.onopen = () => {
          this.connected = true;
          this.emit(ChatEvent.CONNECTED);
          this.tracking.trackChatEvent(ChatEvent.CONNECTED, {
            connectionTime: Date.now() - startTime
          });
          
          // Rejoin all previously joined channels
          this.channels.forEach(channel => {
            this.joinChannel(channel);
          });
          
          resolve();
        };

        this.socket.onclose = () => {
          this.connected = false;
          this.emit(ChatEvent.DISCONNECTED);
          this.tracking.trackChatEvent(ChatEvent.DISCONNECTED);
          
          // Attempt to reconnect after a delay
          setTimeout(() => {
            this.connect();
          }, 5000);
        };

        this.socket.onerror = (error) => {
          this.emit(ChatEvent.ERROR, error);
          this.tracking.trackError(error, { context: 'websocket_connection' });
          reject(error);
        };

        this.socket.onmessage = async (event) => {
          try {
            const message = JSON.parse(event.data);
            
            // Handle different message types
            switch (message.type) {
              case ChatEvent.CHAT_MESSAGE:
                await this.handleChatMessage(message);
                break;
                
              case ChatEvent.USER_JOINED_WS:
                this.handleUserJoined(message);
                break;
                
              case ChatEvent.USER_LEFT_WS:
                this.handleUserLeft(message);
                break;
                
              case ChatEvent.KEY_EXCHANGE:
                await this.handleKeyExchange(message);
                break;
                
              default:
                this.emit(ChatEvent.MESSAGE, message);
            }
          } catch (error) {
            this.emit(ChatEvent.ERROR, error);
          }
        };
      });
    } catch (error) {
      this.tracking.trackError(error, { context: 'chat_connection' });
      throw error;
    }
  }

  /**
   * Disconnects from the WebSocket server
   */
  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
      this.connected = false;
      this.channels.clear();
      this.channelKeys.clear();
      this.onlineUsers.clear();
      this.messageHistory.clear();
      this.pendingMessages.clear();
    }
  }

  /**
   * Joins a chat channel
   * @param channel - The channel to join
   */
  async joinChannel(channel: string): Promise<void> {
    if (!this.connected) {
      await this.connect();
    }

    const startTime = Date.now();
    try {
      // Add to tracked channels
      this.channels.add(channel);
      
      // Load initial message history
      await this.loadMessageHistory(channel);
      
      // Get online users
      await this.getOnlineUsers(channel);
      
      // Send join message
      this.sendToSocket({
        type: ChatEvent.JOIN_CHANNEL,
        channel,
        payload: JSON.stringify({ channel })
      });
      
      this.emit(ChatEvent.CHANNEL_JOINED, channel);
      this.tracking.trackChatEvent(ChatEvent.CHANNEL_JOINED, {
        channel,
        joinTime: Date.now() - startTime
      });
    } catch (error) {
      this.tracking.trackError(error, { context: 'join_channel', channel });
      throw error;
    }
  }

  /**
   * Leaves a chat channel
   * @param channel - The channel to leave
   */
  leaveChannel(channel: string): void {
    if (!this.connected || !this.socket) {
      return;
    }

    this.sendToSocket({
      type: ChatEvent.LEAVE_CHANNEL,
      channel,
      payload: JSON.stringify({ channel })
    });

    this.channels.delete(channel);
    this.channelKeys.delete(channel);
    this.onlineUsers.delete(channel);
    this.messageHistory.delete(channel);
    this.pendingMessages.delete(channel);
    
    this.emit(ChatEvent.CHANNEL_LEFT, channel);
  }

  /**
   * Sends a message to a channel
   * @param channel - The channel to send the message to
   * @param content - The message content
   * @param attachments - Optional attachments
   */
  async sendMessage(channel: string, content: string, attachments?: any[]): Promise<void> {
    if (!this.connected || !this.socket) {
      throw new Error('Not connected to WebSocket server');
    }

    if (!this.channels.has(channel)) {
      throw new Error(`Not joined to channel: ${channel}`);
    }

    const startTime = Date.now();
    const messageTimestamp = Date.now();
    
    const message: ChatMessage = {
      userId: this.sdk.auth.getUserId() || 'anonymous',
      username: this.sdk.auth.getUsername() || 'Anonymous User',
      channel,
      content,
      timestamp: messageTimestamp,
      formattedTime: formatTimestamp(messageTimestamp, 'short'),
      attachments
    };

    try {
      // Store locally first (optimistic UI)
      if (!this.messageHistory.has(channel)) {
        this.messageHistory.set(channel, []);
      }
      this.messageHistory.get(channel)?.push(message);
      this.emit(ChatEvent.MESSAGE_SENT, message);
      this.tracking.trackChatEvent(ChatEvent.MESSAGE_SENT, {
        channel,
        messageLength: content.length,
        hasAttachments: attachments && attachments.length > 0
      });

      // Track pending message
      if (!this.pendingMessages.has(channel)) {
        this.pendingMessages.set(channel, []);
      }
      this.pendingMessages.get(channel)?.push(message);

      // Prepare the message for sending
      const messagePayload = {
        content,
        attachments,
        metadata: {
          clientId: messageTimestamp.toString(),
          timezone: getUserTimezone()
        }
      };

      // Encrypt the message if encryption is enabled
      if (this.options.encryption && this.channelKeys.has(channel)) {
        try {
          const encryptedData = await this.encryptData(
            JSON.stringify(messagePayload),
            channel
          );
          
          this.sendToSocket({
            type: ChatEvent.CHAT_MESSAGE,
            channel,
            encryptedData
          });
        } catch (error) {
          this.tracking.trackError(error, { context: 'send_encrypted_message', channel });
          throw error;
        }
      } else {
        // Fallback to unencrypted message
        this.sendToSocket({
          type: ChatEvent.CHAT_MESSAGE,
          channel,
          payload: JSON.stringify(messagePayload)
        });
      }

      this.tracking.trackMetric('message_send_time', Date.now() - startTime, {
        channel,
        encrypted: this.options.encryption && this.channelKeys.has(channel)
      });
    } catch (error) {
      this.tracking.trackError(error, { context: 'send_message', channel });
      throw error;
    }
  }

  /**
   * Gets the online users in a channel
   * @param channel - The channel to get online users for
   */
  async getOnlineUsers(channel: string): Promise<OnlineUser[]> {
    try {
      const response = await this.sdk.apiClient.get<OnlineUser[]>(`/api/v1/channels/${channel}/users`);
      
      if (response && Array.isArray(response)) {
        this.onlineUsers.set(channel, response);
        this.emit(ChatEvent.ONLINE_USERS_UPDATED, channel, response);
        return response;
      }
      
      return [];
    } catch (error) {
      this.emit(ChatEvent.ERROR, error);
      return [];
    }
  }

  /**
   * Loads message history for a channel
   * @param channel - The channel to load history for
   * @param limit - Number of messages to load
   * @param before - Timestamp to load messages before
   */
  private async loadMessageHistory(channel: string, limit?: number, before?: number): Promise<ChatMessage[]> {
    try {
      const response = await this.sdk.apiClient.get<PaginatedResponse<ChatMessage>>(
        `/api/v1/channels/${channel}/messages`,
        {
          limit: limit || this.options.historyLimit,
          before: before ? before.toString() : undefined
        }
      );
      
      if (response && response.data) {
        // If we already have messages for this channel, merge them
        const existingMessages = this.messageHistory.get(channel) || [];
        
        // Filter out duplicates
        const newMessages = response.data.filter(msg => 
          !existingMessages.some(existing => existing.id === msg.id)
        );
        
        // Add formatted timestamps to all messages
        newMessages.forEach(msg => {
          msg.formattedTime = formatTimestamp(msg.timestamp, 'short');
        });
        
        // Sort by timestamp
        const allMessages = [...existingMessages, ...newMessages].sort((a, b) => a.timestamp - b.timestamp);
        
        // Store in message history
        this.messageHistory.set(channel, allMessages);
        
        // Emit event
        this.emit(ChatEvent.HISTORY_LOADED, channel, allMessages);
        
        return allMessages;
      }
      
      return [];
    } catch (error) {
      this.emit(ChatEvent.ERROR, error);
      return [];
    }
  }

  /**
   * Loads more message history for a channel
   * @param channel - The channel to load more history for
   * @param limit - Number of messages to load
   */
  async loadMoreMessages(channel: string, limit?: number): Promise<ChatMessage[]> {
    if (!this.messageHistory.has(channel)) {
      return this.loadMessageHistory(channel, limit);
    }
    
    const messages = this.messageHistory.get(channel) || [];
    if (messages.length === 0) {
      return this.loadMessageHistory(channel, limit);
    }
    
    // Find the oldest message timestamp
    const oldestMessage = messages.reduce(
      (oldest, msg) => msg.timestamp < oldest.timestamp ? msg : oldest,
      messages[0]
    );
    
    return this.loadMessageHistory(channel, limit, oldestMessage.timestamp);
  }

  /**
   * Gets messages for a channel from local cache
   * @param channel - The channel to get messages for
   */
  getMessages(channel: string): ChatMessage[] {
    return this.messageHistory.get(channel) || [];
  }

  /**
   * Gets online users for a channel from local cache
   * @param channel - The channel to get online users for
   */
  getCachedOnlineUsers(channel: string): OnlineUser[] {
    return this.onlineUsers.get(channel) || [];
  }

  /**
   * Handles an incoming chat message
   * @param message - The message to handle
   */
  private async handleChatMessage(message: any): Promise<void> {
    const startTime = Date.now();
    const { channel, userId, timestamp, encryptedData, payload } = message;
    
    // Skip if we're not in this channel
    if (!this.channels.has(channel)) {
      return;
    }
    
    try {
      let messageContent: any;
      
      // Handle encrypted message
      if (encryptedData && this.options.encryption && this.channelKeys.has(channel)) {
        try {
          const decryptedData = await this.decryptData(encryptedData, channel);
          messageContent = JSON.parse(decryptedData);
        } catch (error) {
          this.tracking.trackError(error, { context: 'decrypt_message', channel });
          this.emit(ChatEvent.ERROR, new Error(`Failed to decrypt message: ${error}`));
          return;
        }
      } else if (payload) {
        // Handle unencrypted message
        try {
          messageContent = typeof payload === 'string' ? JSON.parse(payload) : payload;
        } catch (error) {
          this.tracking.trackError(error, { context: 'parse_message', channel });
          this.emit(ChatEvent.ERROR, new Error(`Failed to parse message payload: ${error}`));
          return;
        }
      } else {
        this.tracking.trackError(new Error('Invalid message format'), { context: 'handle_message', channel });
        this.emit(ChatEvent.ERROR, new Error('Invalid message format: no payload or encrypted data'));
        return;
      }
      
      const messageTimestamp = timestamp || Date.now();
      
      // Build the message object
      const chatMessage: ChatMessage = {
        userId,
        channel,
        content: messageContent.content,
        timestamp: messageTimestamp,
        formattedTime: formatTimestamp(messageTimestamp, 'short'),
        attachments: messageContent.attachments
      };
      
      // Remove from pending if it matches
      this.removePendingMessage(channel, chatMessage);
      
      // Add to message history
      if (!this.messageHistory.has(channel)) {
        this.messageHistory.set(channel, []);
      }
      this.messageHistory.get(channel)?.push(chatMessage);
      
      // Emit the message event
      this.emit(ChatEvent.MESSAGE_RECEIVED, chatMessage);
      this.tracking.trackChatEvent(ChatEvent.MESSAGE_RECEIVED, {
        channel,
        messageLength: chatMessage.content.length,
        hasAttachments: chatMessage.attachments && chatMessage.attachments.length > 0,
        processingTime: Date.now() - startTime
      });
    } catch (error) {
      this.tracking.trackError(error, { context: 'handle_chat_message', channel });
      throw error;
    }
  }

  /**
   * Handles a user joined event
   * @param message - The user joined message
   */
  private handleUserJoined(message: any): void {
    const { channel, userId } = message;
    
    // Skip if we're not in this channel
    if (!this.channels.has(channel)) {
      return;
    }
    
    // Update online users
    this.getOnlineUsers(channel);
    
    // Emit the user joined event
    this.emit(ChatEvent.USER_JOINED, channel, userId);
  }

  /**
   * Handles a user left event
   * @param message - The user left message
   */
  private handleUserLeft(message: any): void {
    const { channel, userId } = message;
    
    // Skip if we're not in this channel
    if (!this.channels.has(channel)) {
      return;
    }
    
    // Update online users
    const users = this.onlineUsers.get(channel) || [];
    const updatedUsers = users.filter(user => user.userId !== userId);
    this.onlineUsers.set(channel, updatedUsers);
    
    // Emit the user left event
    this.emit(ChatEvent.USER_LEFT, channel, userId);
    this.emit(ChatEvent.ONLINE_USERS_UPDATED, channel, updatedUsers);
  }

  /**
   * Handles a key exchange message
   * @param message - The key exchange message
   */
  private async handleKeyExchange(message: any): Promise<void> {
    const { channel, payload } = message;
    
    // Skip if we're not in this channel
    if (!this.channels.has(channel)) {
      return;
    }
    
    try {
      const keyData = typeof payload === 'string' ? JSON.parse(payload) : payload;
      
      if (keyData.channelKey) {
        // Store the channel key
        this.channelKeys.set(channel, keyData.channelKey);
        
        // Acknowledge key receipt
        this.sendToSocket({
          type: ChatEvent.KEY_EXCHANGE_RESPONSE,
          channel,
          payload: JSON.stringify({
            channel,
            channelKey: keyData.channelKey
          })
        });
        
        // Generate a crypto key from the channel key
        await this.getCryptoKeyForChannel(channel);
        
        // Emit the key exchange event
        this.emit(ChatEvent.KEY_EXCHANGE_COMPLETE, channel);
        
        // Process any pending messages now that we have the key
        const pendingMessages = this.pendingMessages.get(channel) || [];
        for (const msg of pendingMessages) {
          await this.sendMessage(channel, msg.content, msg.attachments);
        }
        this.pendingMessages.set(channel, []);
      }
    } catch (error) {
      this.emit(ChatEvent.ERROR, new Error(`Failed to process key exchange: ${error}`));
    }
  }

  /**
   * Removes a message from the pending messages list
   * @param channel - The channel the message was sent to
   * @param message - The message to remove
   */
  private removePendingMessage(channel: string, message: ChatMessage): void {
    const pendingMessages = this.pendingMessages.get(channel) || [];
    
    // Find a matching message based on user ID, close timestamp, and similar content
    const index = pendingMessages.findIndex(
      msg => msg.userId === message.userId &&
             Math.abs(msg.timestamp - message.timestamp) < 5000 &&
             msg.content === message.content
    );
    
    if (index !== -1) {
      pendingMessages.splice(index, 1);
      this.pendingMessages.set(channel, pendingMessages);
    }
  }

  /**
   * Sends a message to the WebSocket
   * @param message - The message to send
   */
  private sendToSocket(message: any): void {
    if (!this.connected || !this.socket) {
      throw new Error('Not connected to WebSocket server');
    }
    
    this.socket.send(JSON.stringify(message));
  }

  /**
   * Gets or creates a CryptoKey for a channel
   * @param channel - The channel to get a key for
   */
  private async getCryptoKeyForChannel(channel: string): Promise<CryptoKey> {
    // Check if we already have a key for this channel
    if (this.cryptoKeyCache.has(channel)) {
      return this.cryptoKeyCache.get(channel)!;
    }
    
    // Get the channel key
    const channelKey = this.channelKeys.get(channel);
    if (!channelKey) {
      throw new Error(`No encryption key available for channel: ${channel}`);
    }
    
    // Convert the key string to bytes
    const keyData = this.encoder.encode(channelKey);
    
    // Create a digest of the key
    const keyDigest = await crypto.subtle.digest('SHA-256', keyData);
    
    // Import the key
    const cryptoKey = await crypto.subtle.importKey(
      'raw',
      keyDigest,
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt', 'decrypt']
    );
    
    // Cache the key
    this.cryptoKeyCache.set(channel, cryptoKey);
    
    return cryptoKey;
  }

  /**
   * Encrypts data for a channel
   * @param data - The data to encrypt
   * @param channel - The channel to encrypt for
   */
  private async encryptData(data: string, channel: string): Promise<string> {
    // Get the crypto key for this channel
    const cryptoKey = await this.getCryptoKeyForChannel(channel);
    
    // Generate a random IV
    const iv = crypto.getRandomValues(new Uint8Array(12));
    
    // Encrypt the data
    const dataBytes = this.encoder.encode(data);
    const encryptedBytes = await crypto.subtle.encrypt(
      {
        name: 'AES-GCM',
        iv
      },
      cryptoKey,
      dataBytes
    );
    
    // Combine IV and encrypted data
    const result = new Uint8Array(iv.length + new Uint8Array(encryptedBytes).length);
    result.set(iv);
    result.set(new Uint8Array(encryptedBytes), iv.length);
    
    // Convert to base64
    return bufferToBase64(result);
  }

  /**
   * Decrypts data from a channel
   * @param encryptedData - The encrypted data
   * @param channel - The channel the data was encrypted for
   */
  private async decryptData(encryptedData: string, channel: string): Promise<string> {
    // Get the crypto key for this channel
    const cryptoKey = await this.getCryptoKeyForChannel(channel);
    
    // Convert from base64
    const encryptedBytes = base64ToBuffer(encryptedData);
    
    // Extract IV and encrypted data
    const iv = encryptedBytes.slice(0, 12);
    const data = encryptedBytes.slice(12);
    
    // Decrypt the data
    const decryptedBytes = await crypto.subtle.decrypt(
      {
        name: 'AES-GCM',
        iv
      },
      cryptoKey,
      data
    );
    
    // Convert to string
    return this.decoder.decode(decryptedBytes);
  }

  destroy(): void {
    this.disconnect();
    this.tracking.destroy();
  }
}

/**
 * Converts an ArrayBuffer to a base64 string
 * @param buffer - The buffer to convert
 */
function bufferToBase64(buffer: ArrayBuffer | Uint8Array): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

/**
 * Converts a base64 string to a Uint8Array
 * @param base64 - The base64 string to convert
 */
function base64ToBuffer(base64: string): Uint8Array {
  const binaryString = atob(base64);
  const bytes = new Uint8Array(binaryString.length);
  for (let i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes;
} 