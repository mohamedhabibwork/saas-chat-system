/**
 * Bot Manager for bot operations
 */
import { HttpClient } from '../http-client';
import { BotConfig } from '../types';

/**
 * Manages bot operations
 */
export class BotManager {
  private client: HttpClient;

  /**
   * Creates a new bot manager
   * @param client HTTP client for API requests
   */
  constructor(client: HttpClient) {
    this.client = client;
  }

  /**
   * Creates a new bot
   * @param botConfig Bot configuration
   * @returns Promise with created bot data
   */
  async create<T = any>(botConfig: BotConfig): Promise<T> {
    return this.client.post<T>('/bots', botConfig);
  }

  /**
   * Gets a bot by ID
   * @param botId Bot ID
   * @returns Promise with bot data
   */
  async get<T = any>(botId: string): Promise<T> {
    return this.client.get<T>(`/bots/${botId}`);
  }

  /**
   * Lists bots
   * @param options List options
   * @returns Promise with bot list
   */
  async list<T = any>(options: Record<string, any> = {}): Promise<T> {
    return this.client.get<T>('/bots', options);
  }

  /**
   * Updates a bot
   * @param botId Bot ID
   * @param botConfig Bot configuration for update
   * @returns Promise with updated bot data
   */
  async update<T = any>(botId: string, botConfig: Partial<BotConfig>): Promise<T> {
    return this.client.put<T>(`/bots/${botId}`, botConfig);
  }

  /**
   * Deletes a bot
   * @param botId Bot ID
   * @returns Promise with deletion result
   */
  async delete<T = any>(botId: string): Promise<T> {
    return this.client.delete<T>(`/bots/${botId}`);
  }

  /**
   * Sends a message to a bot
   * @param botId Bot ID
   * @param message Message content
   * @returns Promise with bot response
   */
  async sendMessage<T = any>(botId: string, message: string): Promise<T> {
    return this.client.post<T>(`/bots/${botId}/messages`, { message });
  }
} 