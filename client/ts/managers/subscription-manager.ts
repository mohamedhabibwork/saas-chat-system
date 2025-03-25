/**
 * Subscription Manager for subscription operations
 */
import { HttpClient } from '../http-client';

/**
 * Manages subscription operations
 */
export class SubscriptionManager {
  private client: HttpClient;

  /**
   * Creates a new subscription manager
   * @param client HTTP client for API requests
   */
  constructor(client: HttpClient) {
    this.client = client;
  }

  /**
   * Gets the current subscription plan
   * @returns Promise with current plan data
   */
  async getCurrentPlan<T = any>(): Promise<T> {
    return this.client.get<T>('/subscription/current');
  }

  /**
   * Gets subscription usage
   * @returns Promise with usage data
   */
  async getUsage<T = any>(): Promise<T> {
    return this.client.get<T>('/subscription/usage');
  }

  /**
   * Upgrades to a different plan
   * @param planId Plan ID to upgrade to
   * @returns Promise with upgrade result
   */
  async upgrade<T = any>(planId: string): Promise<T> {
    return this.client.post<T>('/subscription/upgrade', { planId });
  }

  /**
   * Cancels the current subscription
   * @returns Promise with cancellation result
   */
  async cancel<T = any>(): Promise<T> {
    return this.client.post<T>('/subscription/cancel');
  }

  /**
   * Gets subscription invoices
   * @returns Promise with invoice list
   */
  async getInvoices<T = any>(): Promise<T> {
    return this.client.get<T>('/subscription/invoices');
  }
} 