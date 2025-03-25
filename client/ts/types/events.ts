/**
 * Enum for all chat event names used in the system
 */
export enum ChatEvent {
  // Connection events
  CONNECTED = 'connected',
  DISCONNECTED = 'disconnected',
  ERROR = 'error',

  // Message events
  MESSAGE = 'message',
  MESSAGE_SENT = 'message_sent',
  MESSAGE_RECEIVED = 'message_received',
  HISTORY_LOADED = 'history_loaded',

  // Channel events
  CHANNEL_JOINED = 'channel_joined',
  CHANNEL_LEFT = 'channel_left',
  KEY_EXCHANGE_COMPLETE = 'key_exchange_complete',

  // User events
  USER_JOINED = 'user_joined',
  USER_LEFT = 'user_left',
  ONLINE_USERS_UPDATED = 'online_users_updated',

  // WebSocket message types
  CHAT_MESSAGE = 'chat_message',
  USER_JOINED_WS = 'user_joined',
  USER_LEFT_WS = 'user_left',
  KEY_EXCHANGE = 'key_exchange',
  JOIN_CHANNEL = 'join_channel',
  LEAVE_CHANNEL = 'leave_channel',
  KEY_EXCHANGE_RESPONSE = 'key_exchange_response'
} 