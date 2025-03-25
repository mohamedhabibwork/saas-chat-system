/**
 * HTTP Client for making API requests
 */
import { RequestOptions } from './types';

/**
 * HTTP client for making API requests
 */
export class HttpClient {
  private baseURL: string;
  private apiKey: string;
  private headers: Record<string, string>;

  /**
   * Creates a new HTTP client
   * @param baseURL Base URL for API requests
   * @param apiKey API key for authentication
   */
  constructor(baseURL: string, apiKey: string) {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
    this.headers = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${apiKey}`
    };
  }

  /**
   * Makes a request to the API
   * @param method HTTP method
   * @param endpoint API endpoint
   * @param data Request data
   * @param customHeaders Custom headers
   * @returns Promise with response data
   */
  async request<T>(
    method: string,
    endpoint: string,
    data: any = null,
    customHeaders?: Record<string, string>
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const options: RequestOptions = {
      method,
      headers: { ...this.headers, ...customHeaders }
    };

    if (data) {
      if (data instanceof FormData) {
        // Clone headers and create a new object without Content-Type
        const formDataHeaders = { ...options.headers };
        delete formDataHeaders['Content-Type']; // Fix the linter error
        options.headers = formDataHeaders;
        options.body = data;
      } else {
        options.body = JSON.stringify(data);
      }
    }

    const response = await fetch(url, options);
    const contentType = response.headers.get('Content-Type') || '';
    
    let result;
    if (contentType.includes('application/json')) {
      result = await response.json();
    } else if (contentType.includes('application/octet-stream')) {
      result = await response.blob();
    } else {
      result = await response.text();
    }

    if (!response.ok) {
      const error = new Error((result && result.message) || 'API request failed');
      Object.assign(error, { status: response.status, data: result });
      throw error;
    }

    return result as T;
  }

  /**
   * Makes a GET request
   * @param endpoint API endpoint
   * @param params Query parameters
   * @returns Promise with response data
   */
  async get<T>(endpoint: string, params?: Record<string, any>): Promise<T> {
    if (params) {
      const queryParams = new URLSearchParams();
      for (const key in params) {
        queryParams.set(key, String(params[key]));
      }
      endpoint = `${endpoint}?${queryParams.toString()}`;
    }
    return this.request<T>('GET', endpoint);
  }

  /**
   * Makes a POST request
   * @param endpoint API endpoint
   * @param data Request data
   * @param customHeaders Custom headers
   * @returns Promise with response data
   */
  async post<T>(endpoint: string, data?: any, customHeaders?: Record<string, string>): Promise<T> {
    return this.request<T>('POST', endpoint, data, customHeaders);
  }

  /**
   * Makes a PUT request
   * @param endpoint API endpoint
   * @param data Request data
   * @returns Promise with response data
   */
  async put<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>('PUT', endpoint, data);
  }

  /**
   * Makes a DELETE request
   * @param endpoint API endpoint
   * @returns Promise with response data
   */
  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>('DELETE', endpoint);
  }

  /**
   * Creates a WebSocket connection
   * @returns WebSocket instance
   */
  connectWebSocket(): WebSocket {
    const wsURL = this.baseURL.replace(/^http/, 'ws');
    const ws = new WebSocket(`${wsURL}/ws`);
    
    ws.onopen = () => {
      ws.send(JSON.stringify({
        type: 'auth',
        token: this.apiKey
      }));
    };

    return ws;
  }
} 