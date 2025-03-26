export type EventType = 'user_event' | 'chat_event' | 'system_event' | 'error_event';

export interface TrackingEvent {
  id: string;
  event_type: EventType;
  timestamp: Date;
  user_id: string;
  tenant_id: string;
  metadata?: Record<string, any>;
  created_at: Date;
  updated_at: Date;
}

export interface TrackingMetric {
  id: string;
  name: string;
  value: number;
  timestamp: Date;
  user_id: string;
  tenant_id: string;
  metadata?: Record<string, any>;
  created_at: Date;
  updated_at: Date;
}

export interface TrackingError {
  id: string;
  message: string;
  stack?: string;
  timestamp: Date;
  user_id: string;
  tenant_id: string;
  metadata?: Record<string, any>;
  created_at: Date;
  updated_at: Date;
}

export interface Location {
  id: string;
  user_id: string;
  tenant_id: string;
  latitude: number;
  longitude: number;
  accuracy?: number;
  altitude?: number;
  speed?: number;
  heading?: number;
  timestamp: Date;
  metadata?: Record<string, any>;
  created_at: Date;
  updated_at: Date;
}

export interface LocationHistory {
  id: string;
  user_id: string;
  tenant_id: string;
  locations: Location[];
  start_time: Date;
  end_time: Date;
  created_at: Date;
  updated_at: Date;
}

export interface LocationStats {
  total_locations: number;
  total_history: number;
  last_location: Location;
  last_update: Date;
  average_speed: number;
  max_speed: number;
  total_distance: number;
}

export interface TrackingStats {
  total_events: number;
  total_metrics: number;
  total_errors: number;
  event_types: string[];
  metric_names: string[];
  error_messages: string[];
} 