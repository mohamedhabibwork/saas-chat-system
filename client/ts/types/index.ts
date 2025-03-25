export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  hasMore: boolean;
}

export interface PlatformSDK {
  auth: {
    getToken(): Promise<string>;
    getUserId(): string | null;
    getUsername(): string | null;
  };
  config: {
    apiUrl: string;
    wsUrl?: string;
  };
  apiClient: {
    get<T>(path: string, params?: Record<string, any>): Promise<T>;
    post<T>(path: string, data: any): Promise<T>;
    put<T>(path: string, data: any): Promise<T>;
    delete<T>(path: string): Promise<T>;
  };
}

export interface WebSocketMessage {
  type: string;
  data: any;
} 