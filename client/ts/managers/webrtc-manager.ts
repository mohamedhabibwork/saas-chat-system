/**
 * WebRTC Manager for video/audio communication
 */
import { HttpClient } from '../http-client';

/**
 * Manages WebRTC connections
 */
export class WebRTCManager {
  private client: HttpClient;
  private peerConnections: Map<string, RTCPeerConnection>;

  /**
   * Creates a new WebRTC manager
   * @param client HTTP client for API requests
   */
  constructor(client: HttpClient) {
    this.client = client;
    this.peerConnections = new Map();
  }

  /**
   * Joins a channel for video/audio communication
   * @param channelId Channel ID to join
   * @returns Promise with RTCPeerConnection
   */
  async joinChannel(channelId: string): Promise<RTCPeerConnection> {
    const peerConnection = new RTCPeerConnection({
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' }
      ]
    });

    try {
      // Get user media
      const stream = await navigator.mediaDevices.getUserMedia({
        video: true,
        audio: true
      });

      // Add tracks to peer connection
      stream.getTracks().forEach(track => {
        peerConnection.addTrack(track, stream);
      });

      // Create and send offer
      const offer = await peerConnection.createOffer();
      await peerConnection.setLocalDescription(offer);

      // Send offer to server
      await this.client.post(`/webrtc/offer/${channelId}`, { 
        offer: peerConnection.localDescription 
      });

      this.peerConnections.set(channelId, peerConnection);
      return peerConnection;
    } catch (error) {
      peerConnection.close();
      throw error;
    }
  }

  /**
   * Starts screen sharing in a channel
   * @param channelId Channel ID for screen sharing
   * @returns Promise with MediaStream
   */
  async startScreenShare(channelId: string): Promise<MediaStream> {
    const stream = await navigator.mediaDevices.getDisplayMedia({
      video: true
    });

    const peerConnection = this.peerConnections.get(channelId);
    if (peerConnection) {
      stream.getTracks().forEach(track => {
        peerConnection.addTrack(track, stream);
      });
    } else {
      throw new Error('Must join channel before screen sharing');
    }

    return stream;
  }

  /**
   * Leaves a channel and closes the peer connection
   * @param channelId Channel ID to leave
   */
  async leaveChannel(channelId: string): Promise<void> {
    const peerConnection = this.peerConnections.get(channelId);
    if (peerConnection) {
      peerConnection.close();
      this.peerConnections.delete(channelId);
      await this.client.post(`/webrtc/leave/${channelId}`);
    }
  }
} 