package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"awesomeProject/internal/models"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

// WebRTCService handles WebRTC connections and media streams
type WebRTCService struct {
	db Database
	// WebRTC configuration
	config webrtc.Configuration
	// Active connections map
	connections sync.Map
	// Media tracks map
	tracks sync.Map
}

// NewWebRTCService creates a new WebRTC service
func NewWebRTCService(db Database) *WebRTCService {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	return &WebRTCService{
		db:     db,
		config: config,
	}
}

// CreatePeerConnection creates a new WebRTC peer connection
func (s *WebRTCService) CreatePeerConnection(channelID int, userID int, streamType string) (*webrtc.PeerConnection, error) {
	// Create new peer connection
	peerConnection, err := webrtc.NewPeerConnection(s.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %v", err)
	}

	// Store connection
	connectionKey := fmt.Sprintf("%d_%d", channelID, userID)
	s.connections.Store(connectionKey, peerConnection)

	// Handle ICE candidates
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			// Store ICE candidate for later use
			s.storeICECandidate(channelID, userID, candidate)
		}
	})

	// Handle connection state changes
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		s.handleConnectionStateChange(channelID, userID, state)
	})

	return peerConnection, nil
}

// AddTrack adds a media track to the peer connection
func (s *WebRTCService) AddTrack(channelID int, userID int, track *webrtc.TrackLocalStaticSample) error {
	connectionKey := fmt.Sprintf("%d_%d", channelID, userID)
	peerConnection, ok := s.connections.Load(connectionKey)
	if !ok {
		return fmt.Errorf("peer connection not found")
	}

	pc := peerConnection.(*webrtc.PeerConnection)
	_, err := pc.AddTrack(track)
	if err != nil {
		return fmt.Errorf("failed to add track: %v", err)
	}

	// Store track
	trackKey := fmt.Sprintf("%d_%d_%s", channelID, userID, track.ID())
	s.tracks.Store(trackKey, track)

	return nil
}

// HandleOffer handles an incoming WebRTC offer
func (s *WebRTCService) HandleOffer(channelID int, userID int, offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	connectionKey := fmt.Sprintf("%d_%d", channelID, userID)
	peerConnection, ok := s.connections.Load(connectionKey)
	if !ok {
		return nil, fmt.Errorf("peer connection not found")
	}

	pc := peerConnection.(*webrtc.PeerConnection)

	// Set remote description
	if err := pc.SetRemoteDescription(offer); err != nil {
		return nil, fmt.Errorf("failed to set remote description: %v", err)
	}

	// Create answer
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create answer: %v", err)
	}

	// Set local description
	if err := pc.SetLocalDescription(answer); err != nil {
		return nil, fmt.Errorf("failed to set local description: %v", err)
	}

	return &answer, nil
}

// HandleAnswer handles an incoming WebRTC answer
func (s *WebRTCService) HandleAnswer(channelID int, userID int, answer webrtc.SessionDescription) error {
	connectionKey := fmt.Sprintf("%d_%d", channelID, userID)
	peerConnection, ok := s.connections.Load(connectionKey)
	if !ok {
		return fmt.Errorf("peer connection not found")
	}

	pc := peerConnection.(*webrtc.PeerConnection)
	return pc.SetRemoteDescription(answer)
}

// HandleICECandidate handles an incoming ICE candidate
func (s *WebRTCService) HandleICECandidate(channelID int, userID int, candidate webrtc.ICECandidateInit) error {
	connectionKey := fmt.Sprintf("%d_%d", channelID, userID)
	peerConnection, ok := s.connections.Load(connectionKey)
	if !ok {
		return fmt.Errorf("peer connection not found")
	}

	pc := peerConnection.(*webrtc.PeerConnection)
	return pc.AddICECandidate(candidate)
}

// CloseConnection closes a WebRTC connection
func (s *WebRTCService) CloseConnection(channelID int, userID int) error {
	connectionKey := fmt.Sprintf("%d_%d", channelID, userID)
	peerConnection, ok := s.connections.Load(connectionKey)
	if !ok {
		return fmt.Errorf("peer connection not found")
	}

	pc := peerConnection.(*webrtc.PeerConnection)
	if err := pc.Close(); err != nil {
		return fmt.Errorf("failed to close peer connection: %v", err)
	}

	s.connections.Delete(connectionKey)
	return nil
}

// Helper functions

func (s *WebRTCService) storeICECandidate(channelID int, userID int, candidate *webrtc.ICECandidate) {
	// Store ICE candidate in database
	query := `
		INSERT INTO webrtc_connections (channel_id, user_id, peer_id, stream_type)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (channel_id, user_id) DO UPDATE
		SET peer_id = $3, updated_at = NOW()
	`
	s.db.Exec(query, channelID, userID, candidate.String(), "video")
}

func (s *WebRTCService) handleConnectionStateChange(channelID int, userID int, state webrtc.PeerConnectionState) {
	// Update connection state in database
	query := `
		UPDATE webrtc_connections
		SET active = $1, updated_at = NOW()
		WHERE channel_id = $2 AND user_id = $3
	`
	active := state == webrtc.PeerConnectionStateConnected
	s.db.Exec(query, active, channelID, userID)
} 