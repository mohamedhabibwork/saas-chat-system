/**
 * Platform SDK for TypeScript
 * A comprehensive client library for integrating with our platform
 */

// Export types
export * from './types';

// Import managers and HTTP client
import { HttpClient } from './http-client';
import { WebRTCManager } from './managers/webrtc-manager';
import { FileManager } from './managers/file-manager';
import { ChannelManager } from './managers/channel-manager';
import { BotManager } from './managers/bot-manager';
import { SubscriptionManager } from './managers/subscription-manager';
import { ChatManager } from './managers/chat-manager';
import { SDKConfig, UserData } from './types';

/**
 * Main SDK Class
 */
export class PlatformSDK {
  private baseURL: string;
  private apiKey: string;
  private client: HttpClient;
  
  public webrtc: WebRTCManager;
  public files: FileManager;
  public channels: ChannelManager;
  public bots: BotManager;
  public subscription: SubscriptionManager;
  public chat: ChatManager;

  /**
   * Creates a new SDK instance
   * @param config SDK configuration
   */
  constructor(config: SDKConfig) {
    this.baseURL = config.baseURL || 'https://api.your-platform.com';
    this.apiKey = config.apiKey;
    this.client = new HttpClient(this.baseURL, this.apiKey);
    
    this.webrtc = new WebRTCManager(this.client);
    this.files = new FileManager(this.client);
    this.channels = new ChannelManager(this.client);
    this.bots = new BotManager(this.client);
    this.subscription = new SubscriptionManager(this.client);
    this.chat = new ChatManager(this.client);
  }

  /**
   * Logs in a user
   * @param email User email
   * @param password User password
   * @returns Promise with login result
   */
  async login<T = any>(email: string, password: string): Promise<T> {
    return this.client.post<T>('/auth/login', { email, password });
  }

  /**
   * Registers a new user
   * @param userData User data for registration
   * @returns Promise with registration result
   */
  async register<T = any>(userData: UserData): Promise<T> {
    return this.client.post<T>('/auth/register', userData);
  }

  /**
   * Requests a password reset
   * @param email User email
   * @returns Promise with reset request result
   */
  async resetPassword<T = any>(email: string): Promise<T> {
    return this.client.post<T>('/auth/reset-password', { email });
  }

  /**
   * Verifies a reset token and sets a new password
   * @param token Reset token
   * @param newPassword New password
   * @returns Promise with verification result
   */
  async verifyResetToken<T = any>(token: string, newPassword: string): Promise<T> {
    return this.client.post<T>('/auth/verify-reset', { token, newPassword });
  }

  /**
   * Creates a WebSocket connection for real-time features
   * @returns WebSocket instance
   */
  connectWebSocket(): WebSocket {
    return this.client.connectWebSocket();
  }
}

// Export managers for direct use if needed
export {
  HttpClient,
  WebRTCManager,
  FileManager,
  ChannelManager,
  BotManager,
  SubscriptionManager,
  ChatManager
}; 