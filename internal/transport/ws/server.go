package ws

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"game-server/internal/usecase"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	readTimeout  = 60 * time.Second
	writeTimeout = 10 * time.Second
)

var errInvalidHandshake = errors.New("first message must be create_lobby or join_lobby")

type Server struct {
	service *usecase.LobbyService

	mu      sync.RWMutex
	clients map[string]map[string]*clientConn // lobbyCode -> playerID -> conn
}

type clientConn struct {
	ws *websocket.Conn
	mu sync.Mutex // serialize writes

	lobbyCode string
	playerID  string
}

func NewServer(service *usecase.LobbyService) *Server {
	return &Server{
		service: service,
		clients: make(map[string]map[string]*clientConn),
	}
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			CompressionMode: websocket.CompressionContextTakeover,
		})
		if err != nil {
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "bye")

		cc := &clientConn{ws: c}
		ctx := r.Context()

		if err := s.handshake(ctx, cc); err != nil {
			_ = cc.send(ctx, ServerMessage{Type: "error", Message: err.Error()})
			return
		}
		defer s.unregister(cc)

		_ = s.broadcastLobbyState(ctx, cc.lobbyCode)

		for {
			var msg ClientMessage
			readCtx, cancel := context.WithTimeout(ctx, readTimeout)
			err := wsjson.Read(readCtx, c, &msg)
			cancel()
			if err != nil {
				return
			}

			switch msg.Type {
			case "start_game":
				err = s.service.StartGame(cc.lobbyCode)
			case "play_card":
				// If targetId is set, treat as accusation, otherwise score.
				if msg.TargetID != "" {
					err = s.service.PlayAccusation(cc.lobbyCode, cc.playerID, msg.HandIndex, msg.TargetID)
				} else {
					err = s.service.PlayScore(cc.lobbyCode, cc.playerID, msg.HandIndex)
				}
			case "call_over":
				err = s.service.CallOver(cc.lobbyCode, cc.playerID)
			default:
				err = nil
				_ = cc.send(ctx, ServerMessage{Type: "error", Message: "unknown message type"})
			}

			if err != nil {
				_ = cc.send(ctx, ServerMessage{Type: "error", Message: err.Error()})
				continue
			}
			_ = s.broadcastLobbyState(ctx, cc.lobbyCode)
		}
	})
}

func (s *Server) handshake(ctx context.Context, cc *clientConn) error {
	var msg ClientMessage
	readCtx, cancel := context.WithTimeout(ctx, readTimeout)
	err := wsjson.Read(readCtx, cc.ws, &msg)
	cancel()
	if err != nil {
		return err
	}

	switch msg.Type {
	case "create_lobby":
		res, err := s.service.CreateLobby(msg.Name)
		if err != nil {
			return err
		}
		cc.lobbyCode = res.LobbyCode
		cc.playerID = res.PlayerID
		s.register(cc)
		_ = cc.send(ctx, ServerMessage{Type: "lobby_created", Code: res.LobbyCode, PlayerID: res.PlayerID})
		return nil
	case "join_lobby":
		res, err := s.service.JoinLobby(msg.Code, msg.Name)
		if err != nil {
			return err
		}
		cc.lobbyCode = res.LobbyCode
		cc.playerID = res.PlayerID
		s.register(cc)
		_ = cc.send(ctx, ServerMessage{Type: "lobby_joined", Code: res.LobbyCode, PlayerID: res.PlayerID})
		return nil
	default:
		return errInvalidHandshake
	}
}

func (s *Server) register(cc *clientConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.clients[cc.lobbyCode]
	if !ok {
		m = make(map[string]*clientConn)
		s.clients[cc.lobbyCode] = m
	}
	m[cc.playerID] = cc
}

func (s *Server) unregister(cc *clientConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.clients[cc.lobbyCode]
	if !ok {
		return
	}
	delete(m, cc.playerID)
	if len(m) == 0 {
		delete(s.clients, cc.lobbyCode)
	}
}

func (cc *clientConn) send(ctx context.Context, msg ServerMessage) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	return wsjson.Write(writeCtx, cc.ws, msg)
}

func (s *Server) broadcastLobbyState(ctx context.Context, lobbyCode string) error {
	playerIDs, err := s.service.LobbyPlayerIDs(lobbyCode)
	if err != nil {
		return err
	}

	// Snapshot current connections.
	s.mu.RLock()
	conns := make(map[string]*clientConn)
	for id, cc := range s.clients[lobbyCode] {
		conns[id] = cc
	}
	s.mu.RUnlock()

	for _, playerID := range playerIDs {
		cc := conns[playerID]
		if cc == nil {
			continue
		}
		view, err := s.service.ViewForPlayer(lobbyCode, playerID)
		if err != nil {
			_ = cc.send(ctx, ServerMessage{Type: "error", Message: err.Error()})
			continue
		}
		_ = cc.send(ctx, ServerMessage{Type: "state", State: view})
	}
	return nil
}
