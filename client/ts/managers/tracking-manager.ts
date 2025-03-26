/**
 * Tracking Manager for monitoring user activities and system events
 */
import { PlatformSDK } from '../index';
import { EventEmitter } from 'events';
import { ChatEvent } from '../types/events';

interface TrackingEvent {
  eventType: string;
  timestamp: number;
  userId: string;
  metadata: Record<string, any>;
}

interface TrackingOptions {
  batchSize?: number;
  flushInterval?: number;
  enabled?: boolean;
}

export class TrackingManager extends EventEmitter {
  private sdk: PlatformSDK;
  private events: TrackingEvent[] = [];
  private options: TrackingOptions;
  private flushInterval: NodeJS.Timeout | null = null;

  constructor(sdk: PlatformSDK, options?: TrackingOptions) {
    super();
    this.sdk = sdk;
    this.options = {
      batchSize: 10,
      flushInterval: 30000, // 30 seconds
      enabled: true,
      ...options
    };

    // Start periodic flush
    if (this.options.enabled) {
      this.startFlushInterval();
    }
  }

  /**
   * Track a user event
   * @param eventType - Type of event to track
   * @param metadata - Additional event data
   */
  trackEvent(eventType: string, metadata: Record<string, any> = {}): void {
    if (!this.options.enabled) return;

    const event: TrackingEvent = {
      eventType,
      timestamp: Date.now(),
      userId: this.sdk.auth.getUserId() || 'anonymous',
      metadata: {
        ...metadata,
        platform: 'web',
        version: this.sdk.config.version,
        environment: this.sdk.config.environment
      }
    };

    this.events.push(event);

    // Flush if batch size reached
    if (this.events.length >= this.options.batchSize!) {
      this.flush();
    }
  }

  /**
   * Track chat events
   * @param event - Chat event type
   * @param metadata - Additional event data
   */
  trackChatEvent(event: ChatEvent, metadata: Record<string, any> = {}): void {
    this.trackEvent(`chat.${event}`, metadata);
  }

  /**
   * Track user activity
   * @param activity - Type of activity
   * @param metadata - Additional activity data
   */
  trackActivity(activity: string, metadata: Record<string, any> = {}): void {
    this.trackEvent(`activity.${activity}`, metadata);
  }

  /**
   * Track system events
   * @param event - System event type
   * @param metadata - Additional event data
   */
  trackSystemEvent(event: string, metadata: Record<string, any> = {}): void {
    this.trackEvent(`system.${event}`, metadata);
  }

  /**
   * Track errors
   * @param error - Error object or message
   * @param metadata - Additional error data
   */
  trackError(error: Error | string, metadata: Record<string, any> = {}): void {
    const errorMessage = error instanceof Error ? error.message : error;
    const errorStack = error instanceof Error ? error.stack : undefined;

    this.trackEvent('error', {
      message: errorMessage,
      stack: errorStack,
      ...metadata
    });
  }

  /**
   * Track performance metrics
   * @param metric - Metric name
   * @param value - Metric value
   * @param metadata - Additional metric data
   */
  trackMetric(metric: string, value: number, metadata: Record<string, any> = {}): void {
    this.trackEvent(`metric.${metric}`, {
      value,
      ...metadata
    });
  }

  /**
   * Flush events to the server
   */
  async flush(): Promise<void> {
    if (this.events.length === 0) return;

    const eventsToSend = [...this.events];
    this.events = [];

    try {
      await this.sdk.apiClient.post('/api/v1/tracking/events', {
        events: eventsToSend
      });
    } catch (error) {
      // Put events back in queue if send fails
      this.events = [...eventsToSend, ...this.events];
      this.emit('error', error);
    }
  }

  /**
   * Start periodic flush interval
   */
  private startFlushInterval(): void {
    this.flushInterval = setInterval(() => {
      this.flush();
    }, this.options.flushInterval!);
  }

  /**
   * Stop periodic flush interval
   */
  private stopFlushInterval(): void {
    if (this.flushInterval) {
      clearInterval(this.flushInterval);
      this.flushInterval = null;
    }
  }

  /**
   * Enable tracking
   */
  enable(): void {
    this.options.enabled = true;
    this.startFlushInterval();
  }

  /**
   * Disable tracking
   */
  disable(): void {
    this.options.enabled = false;
    this.stopFlushInterval();
  }

  /**
   * Clean up resources
   */
  destroy(): void {
    this.stopFlushInterval();
    this.flush(); // Flush any remaining events
  }
} 