# WebSocket API Documentation

## Overview

GoBid utilizes WebSockets for real-time auction bidding functionality. Users can connect to auction rooms and place bids in real-time, receiving immediate feedback and updates from other participants.

## WebSocket Endpoint

### Subscribe to Product Auction
```
GET /api/v1/products/ws/subscribe/{product_id}
```

**Authentication Required**: JWT token via Authorization header or query parameter

**Parameters:**
- `product_id` (path): UUID of the product/auction to subscribe to

## Message Structure

All WebSocket messages follow a standardized JSON structure:

```json
{
  "kind": 0,
  "message": "string",
  "amount": 0.0,
  "user_id": "uuid"
}
```

### Message Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `kind` | `uint8` | Yes | Message type identifier (see Kind Types below) |
| `message` | `string` | No | Human-readable message content |
| `amount` | `float64` | No | Bid amount (for bid-related messages) |
| `user_id` | `uuid` | No | User ID (automatically set by server) |

## Kind Types

The `kind` field determines the message type and purpose:

### Client → Server Messages

| Kind | Value | Name | Description |
|------|-------|------|-------------|
| 0 | `PlaceBid` | Place Bid | Client requests to place a new bid |

### Server → Client Messages

| Kind | Value | Name | Description |
|------|-------|------|-------------|
| 1 | `SuccessfullyPlacedBid` | Bid Success | Confirms client's bid was accepted |
| 2 | `NewBidPlaced` | New Bid | Notifies about other users' bids |
| 3 | `AuctionFinished` | Auction End | Indicates auction has ended |
| 4 | `FailedToPlaceBid` | Bid Failed | Bid was rejected (e.g., too low) |
| 5 | `InvalidBody` | Invalid Message | Message format was invalid |

## Message Examples

### Placing a Bid (Client → Server)
```json
{
  "kind": 0,
  "amount": 150.50
}
```

### Successful Bid Response (Server → Client)
```json
{
  "kind": 1,
  "message": "Your bid has been placed!",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

### New Bid Notification (Server → Other Clients)
```json
{
  "kind": 2,
  "message": "New bid has been placed!",
  "amount": 150.50,
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

### Failed Bid Response (Server → Client)
```json
{
  "kind": 4,
  "message": "Bid amount is too low. Minimum bid is $200.00",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

### Auction Finished (Server → All Clients)
```json
{
  "kind": 3,
  "message": "Auction has finished"
}
```

### Invalid Message (Server → Client)
```json
{
  "kind": 5,
  "message": "this message is invalid",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

## Connection Lifecycle

### 1. Authentication
- Include JWT token in Authorization header or as query parameter
- Server validates token and extracts user ID
- Invalid tokens result in HTTP 401 response

### 2. Connection Establishment
- Client connects via WebSocket upgrade
- Server registers client in the auction room
- Client begins receiving real-time updates

### 3. Message Exchange
- Client sends bid messages with `kind: 0`
- Server responds with success/failure messages
- Server broadcasts bid updates to all room participants

### 4. Connection Termination
- Auction ends: Server sends `AuctionFinished` message
- Client disconnect: Server automatically unregisters client
- Network issues: Connection cleanup handled automatically

## Connection Parameters

### Timeouts
- **Read Deadline**: 20 seconds
- **Write Deadline**: 10 seconds
- **Ping Period**: 18 seconds (90% of read deadline)

### Limits
- **Max Message Size**: 512 bytes
- **Send Buffer**: 512 messages per client

### Keep-Alive
- Server sends ping messages every 18 seconds
- Client must respond with pong to maintain connection
- Missing pong responses trigger connection cleanup

## Error Handling

### Connection Errors
- **401 Unauthorized**: Invalid or missing JWT token
- **400 Bad Request**: Invalid product ID or auction ended
- **404 Not Found**: Product does not exist
- **500 Internal Server Error**: Server-side issues

### Message Errors
- **InvalidBody** (`kind: 5`): Malformed JSON or invalid message structure
- **FailedToPlaceBid** (`kind: 4`): Business logic errors (bid too low, auction rules)

## Security Considerations

- All connections require valid JWT authentication
- User ID is extracted from token claims, not client message
- Message validation prevents injection attacks
- Connection limits prevent resource exhaustion
- Automatic cleanup prevents memory leaks

## Rate Limiting

Currently, no explicit rate limiting is implemented, but consider:
- Message size limits (512 bytes)
- Connection timeouts
- Buffer size limits (512 messages)

For production deployments, implement additional rate limiting based on:
- Messages per second per user
- Bid frequency limits
- Connection attempt limits