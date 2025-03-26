import { ApiClient } from './api-client';
import { Report, ReportOptions } from '../types/reporting';

export class ReportingSDK {
    private apiClient: ApiClient;

    constructor(apiClient: ApiClient) {
        this.apiClient = apiClient;
    }

    /**
     * Generate a user activity report
     * @param options Report generation options
     * @returns Promise<Report>
     */
    async generateUserActivityReport(options?: Partial<ReportOptions>): Promise<Report> {
        const response = await this.apiClient.post('/reports/user-activity', options);
        return response.data;
    }

    /**
     * Generate a location report
     * @param options Report generation options
     * @returns Promise<Report>
     */
    async generateLocationReport(options?: Partial<ReportOptions>): Promise<Report> {
        const response = await this.apiClient.post('/reports/location', options);
        return response.data;
    }

    /**
     * Generate a system health report
     * @param options Report generation options
     * @returns Promise<Report>
     */
    async generateSystemHealthReport(options?: Partial<ReportOptions>): Promise<Report> {
        const response = await this.apiClient.post('/reports/system-health', options);
        return response.data;
    }

    /**
     * Export report data to CSV format
     * @param report Report to export
     * @returns Promise<string> CSV data
     */
    async exportToCSV(report: Report): Promise<string> {
        const csvData = this.convertToCSV(report.data);
        return csvData;
    }

    /**
     * Export report data to JSON format
     * @param report Report to export
     * @returns Promise<string> JSON data
     */
    async exportToJSON(report: Report): Promise<string> {
        return JSON.stringify(report.data, null, 2);
    }

    /**
     * Convert report data to CSV format
     * @param data Report data to convert
     * @returns string CSV data
     */
    private convertToCSV(data: any): string {
        const lines: string[] = [];
        
        // Add headers
        const headers = this.getObjectKeys(data);
        lines.push(headers.join(','));

        // Add data rows
        const rows = this.getObjectValues(data);
        rows.forEach(row => {
            const values = headers.map(header => {
                const value = row[header];
                return typeof value === 'string' ? `"${value}"` : value;
            });
            lines.push(values.join(','));
        });

        return lines.join('\n');
    }

    /**
     * Get all object keys recursively
     * @param obj Object to get keys from
     * @returns string[] Array of keys
     */
    private getObjectKeys(obj: any): string[] {
        const keys: string[] = [];
        
        for (const key in obj) {
            if (typeof obj[key] === 'object' && obj[key] !== null) {
                const nestedKeys = this.getObjectKeys(obj[key]);
                keys.push(...nestedKeys.map(k => `${key}.${k}`));
            } else {
                keys.push(key);
            }
        }

        return keys;
    }

    /**
     * Get all object values recursively
     * @param obj Object to get values from
     * @returns any[] Array of values
     */
    private getObjectValues(obj: any): any[] {
        const values: any[] = [];
        
        for (const key in obj) {
            if (typeof obj[key] === 'object' && obj[key] !== null) {
                const nestedValues = this.getObjectValues(obj[key]);
                values.push(...nestedValues);
            } else {
                values.push(obj[key]);
            }
        }

        return values;
    }
} 