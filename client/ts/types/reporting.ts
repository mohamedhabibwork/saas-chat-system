import { TrackingEvent, TrackingMetric, TrackingError } from './tracking';
import { Location, LocationStats } from './location';

export interface ReportOptions {
    startTime?: Date;
    endTime?: Date;
    userId?: string;
    tenantId?: string;
    format?: 'json' | 'csv' | 'pdf';
}

export interface Report {
    id: string;
    type: 'user_activity' | 'location' | 'system_health';
    options: ReportOptions;
    data: ReportData;
    createdAt: Date;
    status: 'pending' | 'completed' | 'failed';
    error?: string;
}

export type ReportData = UserActivityReportData | LocationReportData | SystemHealthReportData;

export interface UserActivityReportData {
    events: TrackingEvent[];
    metrics: TrackingMetric[];
    errors: TrackingError[];
    summary: {
        event_counts: Record<string, number>;
        metric_averages: Record<string, number>;
        error_counts: Record<string, number>;
    };
}

export interface LocationReportData {
    locations: Location[];
    stats: LocationStats;
    summary: {
        total_points: number;
        total_distance: number;
        average_speed: number;
        max_speed: number;
        start_time: Date;
        end_time: Date;
        duration: number;
    };
}

export interface SystemHealthReportData {
    metrics: TrackingMetric[];
    errors: TrackingError[];
    summary: {
        metric_averages: Record<string, number>;
        error_counts: Record<string, number>;
    };
} 