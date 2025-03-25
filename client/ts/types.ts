/**
 * Type definitions for Platform SDK
 */

/**
 * SDK configuration options
 */
export interface SDKConfig {
  baseURL: string;
  apiKey: string;
}

/**
 * User data for registration
 */
export interface UserData {
  username: string;
  email: string;
  password: string;
  [key: string]: any;
}

/**
 * Channel data for creation and updates
 */
export interface ChannelData {
  name: string;
  type: 'public' | 'private' | 'direct';
  description?: string;
  [key: string]: any;
}

/**
 * Bot configuration
 */
export interface BotConfig {
  name: string;
  type: string;
  config: {
    language: string;
    capabilities: string[];
    [key: string]: any;
  };
  [key: string]: any;
}

/**
 * File upload options
 */
export interface FileUploadOptions {
  channelId?: string;
  [key: string]: any;
}

/**
 * Message pagination options
 */
export interface MessageOptions {
  limit?: number;
  before?: string | number;
  [key: string]: any;
}

/**
 * WebSocket message structure
 */
export interface WebSocketMessage {
  type: string;
  data: any;
}

/**
 * HTTP request options
 */
export interface RequestOptions extends RequestInit {
  headers: Record<string, string>;
}

/**
 * API response with pagination
 */
export interface PaginatedResponse<T> {
  items: T[];
  totalCount: number;
  hasMore: boolean;
  nextCursor?: string;
} 