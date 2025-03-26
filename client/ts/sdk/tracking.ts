import { ApiClient } from './api-client';
import { TrackingEvent, TrackingMetric, TrackingError, Location, LocationHistory, LocationStats } from '../types/tracking';

export class TrackingSDK {
  private api: ApiClient;

  constructor(api: ApiClient) {
    this.api = api;
  }

  // Event tracking methods
  async trackEvents(events: TrackingEvent[]): Promise<void> {
    await this.api.post('/tracking/events', events);
  }

  async getEvents(params?: {
    startTime?: Date;
    endTime?: Date;
    eventType?: string;
    limit?: number;
    offset?: number;
  }): Promise<TrackingEvent[]> {
    const response = await this.api.get('/tracking/events', { params });
    return response.data;
  }

  // Metric tracking methods
  async trackMetrics(metrics: TrackingMetric[]): Promise<void> {
    await this.api.post('/tracking/metrics', metrics);
  }

  async getMetrics(params?: {
    startTime?: Date;
    endTime?: Date;
    name?: string;
    limit?: number;
    offset?: number;
  }): Promise<TrackingMetric[]> {
    const response = await this.api.get('/tracking/metrics', { params });
    return response.data;
  }

  // Error tracking methods
  async trackErrors(errors: TrackingError[]): Promise<void> {
    await this.api.post('/tracking/errors', errors);
  }

  async getErrors(params?: {
    startTime?: Date;
    endTime?: Date;
    limit?: number;
    offset?: number;
  }): Promise<TrackingError[]> {
    const response = await this.api.get('/tracking/errors', { params });
    return response.data;
  }

  // Statistics methods
  async getStats(params?: {
    startTime?: Date;
    endTime?: Date;
  }): Promise<{
    totalEvents: number;
    totalMetrics: number;
    totalErrors: number;
    eventTypes: string[];
    metricNames: string[];
    errorMessages: string[];
  }> {
    const response = await this.api.get('/tracking/stats', { params });
    return response.data;
  }
}

export class LocationSDK {
  private api: ApiClient;

  constructor(api: ApiClient) {
    this.api = api;
  }

  // Current location methods
  async updateLocation(location: Omit<Location, 'id' | 'created_at' | 'updated_at'>): Promise<Location> {
    const response = await this.api.post('/location/current', location);
    return response.data;
  }

  async getCurrentLocation(): Promise<Location> {
    const response = await this.api.get('/location/current');
    return response.data;
  }

  // Location history methods
  async saveLocationHistory(history: Omit<LocationHistory, 'id' | 'created_at' | 'updated_at'>): Promise<LocationHistory> {
    const response = await this.api.post('/location/history', history);
    return response.data;
  }

  async getLocationHistory(params?: {
    startTime?: Date;
    endTime?: Date;
  }): Promise<Location[]> {
    const response = await this.api.get('/location/history', { params });
    return response.data;
  }

  // Location statistics methods
  async getLocationStats(): Promise<LocationStats> {
    const response = await this.api.get('/location/stats');
    return response.data;
  }

  // Helper methods for location tracking
  async startLocationTracking(options: {
    interval?: number; // Update interval in milliseconds
    onUpdate?: (location: Location) => void;
    onError?: (error: Error) => void;
  }): Promise<() => void> {
    const interval = options.interval || 5000; // Default to 5 seconds
    let isTracking = true;

    const track = async () => {
      if (!isTracking) return;

      try {
        const position = await this.getCurrentPosition();
        const location = await this.updateLocation({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
          accuracy: position.coords.accuracy,
          altitude: position.coords.altitude || undefined,
          speed: position.coords.speed || undefined,
          heading: position.coords.heading || undefined,
          timestamp: new Date(position.timestamp),
          metadata: {
            source: 'browser_geolocation',
            timestamp: position.timestamp,
          },
        });

        options.onUpdate?.(location);
      } catch (error) {
        options.onError?.(error as Error);
      }

      if (isTracking) {
        setTimeout(track, interval);
      }
    };

    track();

    // Return cleanup function
    return () => {
      isTracking = false;
    };
  }

  private getCurrentPosition(): Promise<GeolocationPosition> {
    return new Promise((resolve, reject) => {
      if (!navigator.geolocation) {
        reject(new Error('Geolocation is not supported by your browser'));
        return;
      }

      navigator.geolocation.getCurrentPosition(resolve, reject, {
        enableHighAccuracy: true,
        timeout: 5000,
        maximumAge: 0,
      });
    });
  }
} 