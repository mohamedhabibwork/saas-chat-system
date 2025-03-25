/**
 * File Manager for file operations
 */
import { HttpClient } from '../http-client';
import { FileUploadOptions } from '../types';

/**
 * Manages file operations
 */
export class FileManager {
  private client: HttpClient;

  /**
   * Creates a new file manager
   * @param client HTTP client for API requests
   */
  constructor(client: HttpClient) {
    this.client = client;
  }

  /**
   * Uploads a file
   * @param file File to upload
   * @param options Upload options
   * @returns Promise with uploaded file data
   */
  async upload<T = any>(file: File, options: FileUploadOptions = {}): Promise<T> {
    const formData = new FormData();
    formData.append('file', file);

    for (const key in options) {
      formData.append(key, String(options[key]));
    }

    return this.client.post<T>('/files/upload', formData);
  }

  /**
   * Downloads a file
   * @param fileId ID of the file to download
   * @returns Promise with file blob
   */
  async download(fileId: string): Promise<Blob> {
    return this.client.get<Blob>(`/files/${fileId}`);
  }

  /**
   * Deletes a file
   * @param fileId ID of the file to delete
   * @returns Promise with deletion result
   */
  async delete<T = any>(fileId: string): Promise<T> {
    return this.client.delete<T>(`/files/${fileId}`);
  }

  /**
   * Lists files
   * @param options List options
   * @returns Promise with file list
   */
  async list<T = any>(options: Record<string, any> = {}): Promise<T> {
    return this.client.get<T>('/files', options);
  }

  /**
   * Shares a file with users
   * @param fileId ID of the file to share
   * @param users List of user identifiers to share with
   * @returns Promise with sharing result
   */
  async share<T = any>(fileId: string, users: string[]): Promise<T> {
    return this.client.post<T>(`/files/${fileId}/share`, { users });
  }
} 