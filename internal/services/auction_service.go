package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log/slog"
	"sync"
	"time"
)

type Kind uint8

const (

	// PlaceBid Requests
	PlaceBid Kind = iota

	// SuccessfullyPlacedBid OK/Success
	SuccessfullyPlacedBid

	// NewBidPlaced Info
	NewBidPlaced
	AuctionFinished

	// FailedToPlaceBid Errors
	FailedToPlaceBid
	InvalidBody
)

type Message struct {
	Message string    `json:"message,omitempty"`
	Amount  float64   `json:"amount,omitempty"`
	Kind    Kind      `json:"kind"`
	UserId  uuid.UUID `json:"user_id,omitempty"`
}

type Client struct {
	Conn   *websocket.Conn
	UserId uuid.UUID
	Send   chan Message
	Room   *AuctionRoom
}

const (
	MaxMessageSize    = 512
	ReadDeadLine      = time.Second * 20
	WriteWaitDeadline = time.Second * 10
	PingPeriod        = (ReadDeadLine * 9) / 10
)

func (c *Client) ReadEventLoop() {
	defer func() {
		c.Room.Unregister <- c
		if err := c.Conn.Close(); err != nil {
			slog.Info("close connection error", "error", err)
		}
	}()

	c.Conn.SetReadLimit(MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(ReadDeadLine))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(ReadDeadLine))
		return nil
	})

	for {
		var m Message
		m.UserId = c.UserId
		if err := c.Conn.ReadJSON(&m); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Unexpected close error", "error", err)
				return
			}
			c.Room.Broadcast <- Message{
				Message: "this message is invalid",
				UserId:  c.UserId,
				Kind:    InvalidBody,
			}
			continue
		}
		c.Room.Broadcast <- m
	}
}

func (c *Client) WriteEventLoop() {
	t := time.NewTicker(PingPeriod)
	defer func() {
		t.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteJSON(Message{
					Kind:    websocket.CloseMessage,
					Message: "closing connection",
				})
				return
			}

			if message.Kind == AuctionFinished {
				close(c.Send)
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(WriteWaitDeadline))
			if err := c.Conn.WriteJSON(message); err != nil {
				c.Room.Unregister <- c
				return
			}

		case <-t.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(WriteWaitDeadline))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("write ping message error", "error", err)
				return
			}

		}
	}
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userId uuid.UUID) *Client {
	return &Client{
		Conn:   conn,
		UserId: userId,
		Send:   make(chan Message, 512),
		Room:   room,
	}
}

type AuctionRoom struct {
	Id         uuid.UUID
	Context    context.Context
	Clients    map[uuid.UUID]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message

	BidsService *BidsService
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, bids *BidsService) *AuctionRoom {
	return &AuctionRoom{
		Id:          id,
		Context:     ctx,
		BidsService: bids,
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[uuid.UUID]*Client),
	}
}

func (a *AuctionRoom) registerClient(c *Client) {
	slog.Info("New user connected", "Client", c)
	a.Clients[c.UserId] = c
}

func (a *AuctionRoom) unregisterClient(c *Client) {
	slog.Info("New user disconnected", "Client", c)
	delete(a.Clients, c.UserId)
}

func (a *AuctionRoom) broadcastMessage(m Message) {
	slog.Info("New message received", "Room ID", a.Id, "Message", m, "UserID", m.UserId)
	switch m.Kind {
	case AuctionFinished:
		for _, client := range a.Clients {
			client.Send <- m
		}
		return
	case InvalidBody:
		client, ok := a.Clients[m.UserId]
		if !ok {
			slog.Info("client not found", "user_id", m.UserId)
			return
		}
		client.Send <- m

	case PlaceBid:
		// place bid in product
		bid, err := a.BidsService.PlaceBid(a.Context, a.Id, m.UserId, m.Amount)
		if err != nil {
			if errors.Is(err, ErrBidIsTooLow) {
				if client, ok := a.Clients[m.UserId]; ok {
					client.Send <- Message{
						Message: err.Error(),
						Kind:    FailedToPlaceBid,
						UserId:  m.UserId,
					}
				}
				return
			}

			return
		}

		if client, ok := a.Clients[m.UserId]; ok {
			client.Send <- Message{
				Kind:    SuccessfullyPlacedBid,
				UserId:  m.UserId,
				Message: "Your bid has been placed!",
			}
		}

		for id, client := range a.Clients {
			if id == m.UserId {
				continue
			}
			newBidMessage := Message{
				Kind:    NewBidPlaced,
				Message: "New bid has been placed!",
				Amount:  bid.Amount,
				UserId:  m.UserId,
			}

			client.Send <- newBidMessage
		}
	case SuccessfullyPlacedBid:
	case NewBidPlaced:
	case FailedToPlaceBid:
	}
}

func (a *AuctionRoom) Start() {
	slog.Info("Starting Auction Room", "Room ID", a.Id)
	defer func() {
		close(a.Broadcast)
		close(a.Register)
		close(a.Unregister)
	}()

	for {
		select {
		case client := <-a.Register:
			a.registerClient(client)
		case client := <-a.Unregister:
			a.unregisterClient(client)
		case message := <-a.Broadcast:
			a.broadcastMessage(message)
		case <-a.Context.Done():
			slog.Info("AuctionRoom stopped", "Auction ID", a.Id)
			for _, client := range a.Clients {
				client.Send <- Message{
					Kind:    AuctionFinished,
					Message: "Auction has finished",
				}
			}
			return
		}
	}
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}
