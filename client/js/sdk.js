/**
 * Platform SDK for JavaScript
 * A comprehensive client library for integrating with our platform
 */

class PlatformSDK {
    constructor(config) {
        this.baseURL = config.baseURL || 'https://api.your-platform.com';
        this.apiKey = config.apiKey;
        this.client = new HttpClient(this.baseURL, this.apiKey);
        this.webrtc = new WebRTCManager(this.client);
        this.files = new FileManager(this.client);
        this.channels = new ChannelManager(this.client);
        this.bots = new BotManager(this.client);
        this.subscription = new SubscriptionManager(this.client);
    }

    // Authentication methods
    async login(email, password) {
        return this.client.post('/auth/login', { email, password });
    }

    async register(userData) {
        return this.client.post('/auth/register', userData);
    }

    async resetPassword(email) {
        return this.client.post('/auth/reset-password', { email });
    }

    async verifyResetToken(token, newPassword) {
        return this.client.post('/auth/verify-reset', { token, newPassword });
    }

    // WebSocket connection for real-time features
    connectWebSocket() {
        return this.client.connectWebSocket();
    }
}

class HttpClient {
    constructor(baseURL, apiKey) {
        this.baseURL = baseURL;
        this.apiKey = apiKey;
        this.headers = {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${apiKey}`
        };
    }

    async request(method, endpoint, data = null) {
        const url = `${this.baseURL}${endpoint}`;
        const options = {
            method,
            headers: this.headers
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        const response = await fetch(url, options);
        const result = await response.json();

        if (!response.ok) {
            throw new Error(result.message || 'API request failed');
        }

        return result;
    }

    get(endpoint) {
        return this.request('GET', endpoint);
    }

    post(endpoint, data) {
        return this.request('POST', endpoint, data);
    }

    put(endpoint, data) {
        return this.request('PUT', endpoint, data);
    }

    delete(endpoint) {
        return this.request('DELETE', endpoint);
    }

    connectWebSocket() {
        const ws = new WebSocket(`${this.baseURL.replace('http', 'ws')}/ws`);
        
        ws.onopen = () => {
            ws.send(JSON.stringify({
                type: 'auth',
                token: this.apiKey
            }));
        };

        return ws;
    }
}

class WebRTCManager {
    constructor(client) {
        this.client = client;
        this.peerConnections = new Map();
    }

    async joinChannel(channelId) {
        const peerConnection = new RTCPeerConnection({
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' }
            ]
        });

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
        await this.client.post(`/webrtc/offer/${channelId}`, { offer });

        this.peerConnections.set(channelId, peerConnection);
        return peerConnection;
    }

    async startScreenShare(channelId) {
        const stream = await navigator.mediaDevices.getDisplayMedia({
            video: true
        });

        const peerConnection = this.peerConnections.get(channelId);
        if (peerConnection) {
            stream.getTracks().forEach(track => {
                peerConnection.addTrack(track, stream);
            });
        }

        return stream;
    }

    async leaveChannel(channelId) {
        const peerConnection = this.peerConnections.get(channelId);
        if (peerConnection) {
            peerConnection.close();
            this.peerConnections.delete(channelId);
        }
    }
}

class FileManager {
    constructor(client) {
        this.client = client;
    }

    async upload(file, options = {}) {
        const formData = new FormData();
        formData.append('file', file);

        if (options.channelId) {
            formData.append('channelId', options.channelId);
        }

        return this.client.post('/files/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        });
    }

    async download(fileId) {
        const response = await this.client.get(`/files/${fileId}`);
        return response.blob();
    }

    async delete(fileId) {
        return this.client.delete(`/files/${fileId}`);
    }

    async list(options = {}) {
        return this.client.get('/files', options);
    }

    async share(fileId, users) {
        return this.client.post(`/files/${fileId}/share`, { users });
    }
}

class ChannelManager {
    constructor(client) {
        this.client = client;
    }

    async create(channelData) {
        return this.client.post('/channels', channelData);
    }

    async get(channelId) {
        return this.client.get(`/channels/${channelId}`);
    }

    async list() {
        return this.client.get('/channels');
    }

    async update(channelId, channelData) {
        return this.client.put(`/channels/${channelId}`, channelData);
    }

    async delete(channelId) {
        return this.client.delete(`/channels/${channelId}`);
    }

    async addMember(channelId, userId) {
        return this.client.post(`/channels/${channelId}/members`, { userId });
    }

    async removeMember(channelId, userId) {
        return this.client.delete(`/channels/${channelId}/members/${userId}`);
    }

    async sendMessage(channelId, content) {
        return this.client.post(`/channels/${channelId}/messages`, { content });
    }

    async getMessages(channelId, options = {}) {
        return this.client.get(`/channels/${channelId}/messages`, options);
    }
}

class BotManager {
    constructor(client) {
        this.client = client;
    }

    async create(botConfig) {
        return this.client.post('/bots', botConfig);
    }

    async get(botId) {
        return this.client.get(`/bots/${botId}`);
    }

    async list() {
        return this.client.get('/bots');
    }

    async update(botId, botConfig) {
        return this.client.put(`/bots/${botId}`, botConfig);
    }

    async delete(botId) {
        return this.client.delete(`/bots/${botId}`);
    }

    async sendMessage(botId, message) {
        return this.client.post(`/bots/${botId}/messages`, { message });
    }
}

class SubscriptionManager {
    constructor(client) {
        this.client = client;
    }

    async getCurrentPlan() {
        return this.client.get('/subscription/current');
    }

    async getUsage() {
        return this.client.get('/subscription/usage');
    }

    async upgrade(planId) {
        return this.client.post('/subscription/upgrade', { planId });
    }

    async cancel() {
        return this.client.post('/subscription/cancel');
    }

    async getInvoices() {
        return this.client.get('/subscription/invoices');
    }
}

// Export the SDK
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PlatformSDK;
} else if (typeof window !== 'undefined') {
    window.PlatformSDK = PlatformSDK;
} 