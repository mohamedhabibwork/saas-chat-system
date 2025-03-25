/**
 * Channel Manager for channel operations
 */
import { HttpClient } from '../http-client';
import { ChannelData, MessageOptions } from '../types';

/**
 * Manages channel operations
 */
export class ChannelManager {
  private client: HttpClient;

  /**
   * Creates a new channel manager
   * @param client HTTP client for API requests
   */
  constructor(client: HttpClient) {
    this.client = client;
  }

  /**
   * Creates a new channel
   * @param channelData Channel data for creation
   * @returns Promise with created channel data
   */
  async create<T = any>(channelData: ChannelData): Promise<T> {
    return this.client.post<T>('/channels', channelData);
  }

  /**
   * Gets a channel by ID
   * @param channelId Channel ID
   * @returns Promise with channel data
   */
  async get<T = any>(channelId: string): Promise<T> {
    return this.client.get<T>(`/channels/${channelId}`);
  }

  /**
   * Lists channels
   * @param options List options
   * @returns Promise with channel list
   */
  async list<T = any>(options: Record<string, any> = {}): Promise<T> {
    return this.client.get<T>('/channels', options);
  }

  /**
   * Updates a channel
   * @param channelId Channel ID
   * @param channelData Channel data for update
   * @returns Promise with updated channel data
   */
  async update<T = any>(channelId: string, channelData: Partial<ChannelData>): Promise<T> {
    return this.client.put<T>(`/channels/${channelId}`, channelData);
  }

  /**
   * Deletes a channel
   * @param channelId Channel ID
   * @returns Promise with deletion result
   */
  async delete<T = any>(channelId: string): Promise<T> {
    return this.client.delete<T>(`/channels/${channelId}`);
  }

  /**
   * Adds a member to a channel
   * @param channelId Channel ID
   * @param userId User ID to add
   * @returns Promise with result
   */
  async addMember<T = any>(channelId: string, userId: string): Promise<T> {
    return this.client.post<T>(`/channels/${channelId}/members`, { userId });
  }

  /**
   * Removes a member from a channel
   * @param channelId Channel ID
   * @param userId User ID to remove
   * @returns Promise with result
   */
  async removeMember<T = any>(channelId: string, userId: string): Promise<T> {
    return this.client.delete<T>(`/channels/${channelId}/members/${userId}`);
  }

  /**
   * Sends a message to a channel
   * @param channelId Channel ID
   * @param content Message content
   * @returns Promise with sent message data
   */
  async sendMessage<T = any>(channelId: string, content: string): Promise<T> {
    return this.client.post<T>(`/channels/${channelId}/messages`, { content });
  }

  /**
   * Gets messages from a channel
   * @param channelId Channel ID
   * @param options Message options
   * @returns Promise with message list
   */
  async getMessages<T = any>(channelId: string, options: MessageOptions = {}): Promise<T> {
    return this.client.get<T>(`/channels/${channelId}/messages`, options);
  }
} 