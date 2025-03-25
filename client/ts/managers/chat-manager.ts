/**
 * Chat Manager for chat operations and external chat service integration
 */
import { HttpClient } from '../http-client';
import { WebSocketMessage, PaginatedResponse } from '../types';

interface ChatMessage {
  id: string;
  content: string;
  sender: string;
  timestamp: string;
  type: string;
}

interface OnlineUser {
  id: string;
  username: string;
  status: 'online' | 'away' | 'offline';
  lastSeen?: string;
}

interface ChatOptions {
  limit?: number;
  before?: string;
  after?: string;
}

/**
 * Manages chat operations and external chat service integration
 */
export class ChatManager {
  private client: HttpClient;
  private socket: WebSocket | null = null;
  private messageHandlers: Map<string, (data: any) => void> = new Map();
  private onlineUsers: Map<string, OnlineUser> = new Map();
  private messageHistory: Map<string, ChatMessage[]> = new Map();

  /**
   * Creates a new chat manager
   * @param client HTTP client for API requests
   */
  constructor(client: HttpClient) {
    this.client = client;
  }

  /**
   * Joins a chat channel and loads message history
   * @param channelId Channel ID to join
   * @param options Chat options including pagination
   * @returns Promise with initial messages and online users
   */
  async joinChannel(channelId: string, options: ChatOptions = {}): Promise<{
    messages: ChatMessage[];
    onlineUsers: OnlineUser[];
  }> {
    // Load initial message history
    const history = await this.loadMessageHistory(channelId, options);
    this.messageHistory.set(channelId, history.messages);

    // Connect to WebSocket if not already connected
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      await this.connectToService(this.client.baseURL.replace(/^http/, 'ws'), this.client.apiKey);
    }

    // Join channel through WebSocket
    this.sendMessage('join_channel', { channelId });

    // Get online users
    const onlineUsers = await this.getOnlineUsers(channelId);
    this.updateOnlineUsers(channelId, onlineUsers);

    return {
      messages: history.messages,
      onlineUsers: Array.from(this.onlineUsers.values())
    };
  }

  /**
   * Loads message history for a channel
   * @param channelId Channel ID
   * @param options Pagination options
   * @returns Promise with paginated messages
   */
  private async loadMessageHistory(
    channelId: string,
    options: ChatOptions = {}
  ): Promise<PaginatedResponse<ChatMessage>> {
    const { limit = 10, before, after } = options;
    return this.client.get<PaginatedResponse<ChatMessage>>(`/channels/${channelId}/messages`, {
      limit,
      before,
      after
    });
  }

  /**
   * Gets online users for a channel
   * @param channelId Channel ID
   * @returns Promise with online users
   */
  private async getOnlineUsers(channelId: string): Promise<OnlineUser[]> {
    return this.client.get<OnlineUser[]>(`/channels/${channelId}/online-users`);
  }

  /**
   * Updates the online users list
   * @param channelId Channel ID
   * @param users Updated users list
   */
  private updateOnlineUsers(channelId: string, users: OnlineUser[]): void {
    users.forEach(user => {
      this.onlineUsers.set(user.id, user);
    });
  }

  /**
   * Loads more messages for a channel
   * @param channelId Channel ID
   * @param options Pagination options
   * @returns Promise with additional messages
   */
  async loadMoreMessages(channelId: string, options: ChatOptions = {}): Promise<ChatMessage[]> {
    const currentMessages = this.messageHistory.get(channelId) || [];
    const lastMessage = currentMessages[currentMessages.length - 1];
    
    if (lastMessage) {
      options.before = lastMessage.id;
    }

    const response = await this.loadMessageHistory(channelId, options);
    const newMessages = [...currentMessages, ...response.messages];
    this.messageHistory.set(channelId, newMessages);

    return response.messages;
  }

  /**
   * Connects to a custom chat service
   * @param serviceUrl The URL of the chat service
   * @param authToken Authentication token for the external chat service
   * @returns Promise that resolves when connected
   */
  async connectToService(serviceUrl: string, authToken: string): Promise<void> {
    // Disconnect existing chat if connected
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.close();
    }

    // Connect to the external chat service
    this.socket = new WebSocket(serviceUrl);
    
    this.socket.onopen = () => {
      // Authenticate with the external service
      if (this.socket) {
        this.socket.send(JSON.stringify({
          type: 'auth',
          token: authToken
        }));
      }
    };

    this.socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as WebSocketMessage;
        
        // Handle different message types
        switch (message.type) {
          case 'chat_message':
            this.handleChatMessage(message.data);
            break;
          case 'user_status':
            this.handleUserStatus(message.data);
            break;
          case 'user_joined':
            this.handleUserJoined(message.data);
            break;
          case 'user_left':
            this.handleUserLeft(message.data);
            break;
        }
        
        // Dispatch message to registered handlers
        const handler = this.messageHandlers.get(message.type);
        if (handler) {
          handler(message.data);
        }
        
        // Also dispatch to 'all' handler if registered
        const allHandler = this.messageHandlers.get('all');
        if (allHandler) {
          allHandler(message);
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    this.socket.onerror = (error) => {
      console.error('Chat service connection error:', error);
    };

    return new Promise((resolve, reject) => {
      if (!this.socket) return reject(new Error('Failed to create WebSocket'));
      
      const onOpen = () => {
        if (this.socket) {
          this.socket.removeEventListener('open', onOpen);
          this.socket.removeEventListener('error', onError);
        }
        resolve();
      };
      
      const onError = (event: Event) => {
        if (this.socket) {
          this.socket.removeEventListener('open', onOpen);
          this.socket.removeEventListener('error', onError);
        }
        reject(new Error('Failed to connect to chat service'));
      };
      
      this.socket.addEventListener('open', onOpen);
      this.socket.addEventListener('error', onError);
    });
  }

  /**
   * Handles incoming chat messages
   * @param data Message data
   */
  private handleChatMessage(data: any): void {
    const { channelId, message } = data;
    const messages = this.messageHistory.get(channelId) || [];
    messages.push(message);
    this.messageHistory.set(channelId, messages);
  }

  /**
   * Handles user status updates
   * @param data Status update data
   */
  private handleUserStatus(data: any): void {
    const { userId, status, lastSeen } = data;
    const user = this.onlineUsers.get(userId);
    if (user) {
      user.status = status;
      user.lastSeen = lastSeen;
      this.onlineUsers.set(userId, user);
    }
  }

  /**
   * Handles user join events
   * @param data User join data
   */
  private handleUserJoined(data: any): void {
    const { user } = data;
    this.onlineUsers.set(user.id, user);
  }

  /**
   * Handles user leave events
   * @param data User leave data
   */
  private handleUserLeft(data: any): void {
    const { userId } = data;
    this.onlineUsers.delete(userId);
  }

  /**
   * Sends a message through the connected chat service
   * @param type Message type
   * @param data Message data
   * @throws Error if not connected
   */
  sendMessage(type: string, data: any): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      throw new Error('Not connected to chat service');
    }

    this.socket.send(JSON.stringify({
      type,
      data
    }));
  }

  /**
   * Registers a handler for specific message types
   * @param type Message type to handle
   * @param handler Handler function
   */
  onMessage(type: string, handler: (data: any) => void): void {
    this.messageHandlers.set(type, handler);
  }

  /**
   * Gets current online users
   * @returns Array of online users
   */
  getOnlineUsers(): OnlineUser[] {
    return Array.from(this.onlineUsers.values());
  }

  /**
   * Gets message history for a channel
   * @param channelId Channel ID
   * @returns Array of messages
   */
  getMessageHistory(channelId: string): ChatMessage[] {
    return this.messageHistory.get(channelId) || [];
  }

  /**
   * Closes the chat connection
   */
  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this.onlineUsers.clear();
    this.messageHistory.clear();
  }
} 