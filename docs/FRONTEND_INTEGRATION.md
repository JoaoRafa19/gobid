# Frontend Integration Guide

## Overview

This guide provides comprehensive examples and best practices for integrating with the GoBid WebSocket API from frontend applications.

## JavaScript/TypeScript Integration

### Basic WebSocket Connection

```javascript
class GoBidWebSocket {
  constructor(productId, authToken) {
    this.productId = productId;
    this.authToken = authToken;
    this.socket = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.listeners = {};
  }

  connect() {
    const wsUrl = `ws://localhost:3080/api/v1/products/ws/subscribe/${this.productId}`;

    this.socket = new WebSocket(wsUrl, [], {
      headers: {
        'Authorization': `Bearer ${this.authToken}`
      }
    });

    this.socket.onopen = this.onOpen.bind(this);
    this.socket.onmessage = this.onMessage.bind(this);
    this.socket.onclose = this.onClose.bind(this);
    this.socket.onerror = this.onError.bind(this);
  }

  onOpen(event) {
    console.log('Connected to auction room:', this.productId);
    this.reconnectAttempts = 0;
    this.emit('connected', event);
  }

  onMessage(event) {
    try {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    } catch (error) {
      console.error('Failed to parse message:', error);
    }
  }

  onClose(event) {
    console.log('Disconnected from auction room');
    this.emit('disconnected', event);

    if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnect();
    }
  }

  onError(error) {
    console.error('WebSocket error:', error);
    this.emit('error', error);
  }

  handleMessage(message) {
    const { kind, amount, user_id, message: msg } = message;

    switch (kind) {
      case 1: // SuccessfullyPlacedBid
        this.emit('bidSuccess', { message: msg, userId: user_id });
        break;
      case 2: // NewBidPlaced
        this.emit('newBid', { amount, userId: user_id, message: msg });
        break;
      case 3: // AuctionFinished
        this.emit('auctionFinished', { message: msg });
        this.disconnect();
        break;
      case 4: // FailedToPlaceBid
        this.emit('bidError', { message: msg, userId: user_id });
        break;
      case 5: // InvalidBody
        this.emit('invalidMessage', { message: msg });
        break;
      default:
        console.warn('Unknown message kind:', kind);
    }
  }

  placeBid(amount) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      const message = {
        kind: 0, // PlaceBid
        amount: parseFloat(amount)
      };
      this.socket.send(JSON.stringify(message));
    } else {
      throw new Error('WebSocket is not connected');
    }
  }

  disconnect() {
    if (this.socket) {
      this.socket.close(1000, 'Client disconnect');
    }
  }

  reconnect() {
    setTimeout(() => {
      this.reconnectAttempts++;
      console.log(`Reconnection attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);
      this.connect();
    }, Math.pow(2, this.reconnectAttempts) * 1000); // Exponential backoff
  }

  on(event, callback) {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(callback);
  }

  emit(event, data) {
    if (this.listeners[event]) {
      this.listeners[event].forEach(callback => callback(data));
    }
  }
}
```

### TypeScript Types

```typescript
// Message Types
interface WebSocketMessage {
  kind: MessageKind;
  message?: string;
  amount?: number;
  user_id?: string;
}

enum MessageKind {
  PlaceBid = 0,
  SuccessfullyPlacedBid = 1,
  NewBidPlaced = 2,
  AuctionFinished = 3,
  FailedToPlaceBid = 4,
  InvalidBody = 5
}

// Event Types
interface BidSuccessEvent {
  message: string;
  userId: string;
}

interface NewBidEvent {
  amount: number;
  userId: string;
  message: string;
}

interface BidErrorEvent {
  message: string;
  userId: string;
}

interface AuctionFinishedEvent {
  message: string;
}

// Usage with TypeScript
class TypedGoBidWebSocket extends GoBidWebSocket {
  on(event: 'connected', callback: (event: Event) => void): void;
  on(event: 'disconnected', callback: (event: CloseEvent) => void): void;
  on(event: 'bidSuccess', callback: (data: BidSuccessEvent) => void): void;
  on(event: 'newBid', callback: (data: NewBidEvent) => void): void;
  on(event: 'bidError', callback: (data: BidErrorEvent) => void): void;
  on(event: 'auctionFinished', callback: (data: AuctionFinishedEvent) => void): void;
  on(event: string, callback: (data: any) => void): void {
    super.on(event, callback);
  }
}
```

### React Hook Example

```jsx
import { useState, useEffect, useRef } from 'react';

const useAuctionWebSocket = (productId, authToken) => {
  const [connectionStatus, setConnectionStatus] = useState('disconnected');
  const [currentBid, setCurrentBid] = useState(null);
  const [bidHistory, setBidHistory] = useState([]);
  const [error, setError] = useState(null);
  const socketRef = useRef(null);

  useEffect(() => {
    if (!productId || !authToken) return;

    const socket = new GoBidWebSocket(productId, authToken);
    socketRef.current = socket;

    socket.on('connected', () => {
      setConnectionStatus('connected');
      setError(null);
    });

    socket.on('disconnected', () => {
      setConnectionStatus('disconnected');
    });

    socket.on('bidSuccess', (data) => {
      console.log('Bid placed successfully:', data.message);
    });

    socket.on('newBid', (data) => {
      setCurrentBid({
        amount: data.amount,
        userId: data.userId,
        timestamp: new Date()
      });

      setBidHistory(prev => [...prev, {
        amount: data.amount,
        userId: data.userId,
        timestamp: new Date()
      }]);
    });

    socket.on('bidError', (data) => {
      setError(data.message);
    });

    socket.on('auctionFinished', (data) => {
      setConnectionStatus('finished');
      console.log('Auction finished:', data.message);
    });

    socket.on('error', (error) => {
      setError('Connection error occurred');
      setConnectionStatus('error');
    });

    socket.connect();

    return () => {
      socket.disconnect();
    };
  }, [productId, authToken]);

  const placeBid = (amount) => {
    if (socketRef.current) {
      try {
        socketRef.current.placeBid(amount);
        setError(null);
      } catch (err) {
        setError(err.message);
      }
    }
  };

  return {
    connectionStatus,
    currentBid,
    bidHistory,
    error,
    placeBid
  };
};

// React Component Example
const AuctionRoom = ({ productId, authToken }) => {
  const { connectionStatus, currentBid, bidHistory, error, placeBid } = useAuctionWebSocket(productId, authToken);
  const [bidAmount, setBidAmount] = useState('');

  const handlePlaceBid = (e) => {
    e.preventDefault();
    if (bidAmount && parseFloat(bidAmount) > 0) {
      placeBid(bidAmount);
      setBidAmount('');
    }
  };

  return (
    <div className="auction-room">
      <div className="connection-status">
        Status: <span className={`status-${connectionStatus}`}>{connectionStatus}</span>
      </div>

      {error && (
        <div className="error-message">
          {error}
        </div>
      )}

      {currentBid && (
        <div className="current-bid">
          <h3>Current Highest Bid: ${currentBid.amount}</h3>
          <p>By: {currentBid.userId}</p>
        </div>
      )}

      <form onSubmit={handlePlaceBid}>
        <input
          type="number"
          step="0.01"
          value={bidAmount}
          onChange={(e) => setBidAmount(e.target.value)}
          placeholder="Enter bid amount"
          disabled={connectionStatus !== 'connected'}
        />
        <button type="submit" disabled={connectionStatus !== 'connected'}>
          Place Bid
        </button>
      </form>

      <div className="bid-history">
        <h4>Bid History</h4>
        {bidHistory.map((bid, index) => (
          <div key={index} className="bid-item">
            ${bid.amount} - {bid.userId} at {bid.timestamp.toLocaleTimeString()}
          </div>
        ))}
      </div>
    </div>
  );
};
```

### Vue.js Example

```vue
<template>
  <div class="auction-room">
    <div class="connection-status">
      Status: {{ connectionStatus }}
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <div v-if="currentBid" class="current-bid">
      <h3>Current Highest Bid: ${{ currentBid.amount }}</h3>
      <p>By: {{ currentBid.userId }}</p>
    </div>

    <form @submit.prevent="handlePlaceBid">
      <input
        v-model="bidAmount"
        type="number"
        step="0.01"
        placeholder="Enter bid amount"
        :disabled="connectionStatus !== 'connected'"
      />
      <button type="submit" :disabled="connectionStatus !== 'connected'">
        Place Bid
      </button>
    </form>

    <div class="bid-history">
      <h4>Bid History</h4>
      <div v-for="(bid, index) in bidHistory" :key="index" class="bid-item">
        ${{ bid.amount }} - {{ bid.userId }} at {{ formatTime(bid.timestamp) }}
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue';

export default {
  name: 'AuctionRoom',
  props: {
    productId: String,
    authToken: String
  },
  setup(props) {
    const connectionStatus = ref('disconnected');
    const currentBid = ref(null);
    const bidHistory = ref([]);
    const error = ref(null);
    const bidAmount = ref('');
    let socket = null;

    const connectToAuction = () => {
      socket = new GoBidWebSocket(props.productId, props.authToken);

      socket.on('connected', () => {
        connectionStatus.value = 'connected';
        error.value = null;
      });

      socket.on('newBid', (data) => {
        currentBid.value = {
          amount: data.amount,
          userId: data.userId,
          timestamp: new Date()
        };

        bidHistory.value.push({
          amount: data.amount,
          userId: data.userId,
          timestamp: new Date()
        });
      });

      socket.on('bidError', (data) => {
        error.value = data.message;
      });

      socket.connect();
    };

    const handlePlaceBid = () => {
      if (bidAmount.value && parseFloat(bidAmount.value) > 0) {
        socket.placeBid(bidAmount.value);
        bidAmount.value = '';
      }
    };

    const formatTime = (timestamp) => {
      return timestamp.toLocaleTimeString();
    };

    onMounted(() => {
      if (props.productId && props.authToken) {
        connectToAuction();
      }
    });

    onUnmounted(() => {
      if (socket) {
        socket.disconnect();
      }
    });

    return {
      connectionStatus,
      currentBid,
      bidHistory,
      error,
      bidAmount,
      handlePlaceBid,
      formatTime
    };
  }
};
</script>
```

## Best Practices

### 1. Authentication
```javascript
// Always include JWT token in connection
const token = localStorage.getItem('authToken');
if (!token) {
  throw new Error('Authentication required');
}

// Handle token expiration
socket.on('error', (error) => {
  if (error.code === 401) {
    // Redirect to login or refresh token
    window.location.href = '/login';
  }
});
```

### 2. Error Handling
```javascript
// Implement comprehensive error handling
socket.on('bidError', (data) => {
  switch (data.message) {
    case 'Bid amount is too low':
      showMinimumBidError();
      break;
    case 'Auction has ended':
      redirectToAuctionResults();
      break;
    default:
      showGenericError(data.message);
  }
});
```

### 3. Connection Management
```javascript
// Implement connection health monitoring
let pingInterval;

socket.on('connected', () => {
  pingInterval = setInterval(() => {
    if (socket.readyState === WebSocket.OPEN) {
      socket.ping();
    }
  }, 30000); // Ping every 30 seconds
});

socket.on('disconnected', () => {
  clearInterval(pingInterval);
});
```

### 4. State Management
```javascript
// Use state management for complex applications
// Redux example
const auctionSlice = createSlice({
  name: 'auction',
  initialState: {
    connectionStatus: 'disconnected',
    currentBid: null,
    bidHistory: [],
    error: null
  },
  reducers: {
    setConnectionStatus: (state, action) => {
      state.connectionStatus = action.payload;
    },
    addBid: (state, action) => {
      state.bidHistory.push(action.payload);
      state.currentBid = action.payload;
    },
    setError: (state, action) => {
      state.error = action.payload;
    }
  }
});
```

## Testing

### Unit Testing Example
```javascript
// Jest + WebSocket testing
import WS from 'jest-websocket-mock';

describe('GoBidWebSocket', () => {
  let server;

  beforeEach(() => {
    server = new WS('ws://localhost:3080/api/v1/products/ws/subscribe/test-id');
  });

  afterEach(() => {
    WS.clean();
  });

  test('should connect and place bid', async () => {
    const socket = new GoBidWebSocket('test-id', 'test-token');
    socket.connect();

    await server.connected;

    socket.placeBid(100);

    await expect(server).toReceiveMessage(JSON.stringify({
      kind: 0,
      amount: 100
    }));
  });
});
```

## Security Considerations

1. **Token Management**: Store JWT tokens securely (httpOnly cookies preferred)
2. **Input Validation**: Validate bid amounts on client-side before sending
3. **Rate Limiting**: Implement client-side rate limiting for bid requests
4. **Error Information**: Don't expose sensitive server information in error messages
5. **HTTPS/WSS**: Use secure WebSocket connections in production

## Performance Optimization

1. **Connection Pooling**: Reuse connections when possible
2. **Message Batching**: Batch multiple operations when applicable
3. **Lazy Loading**: Connect only when auction room is visible
4. **Memory Management**: Clean up event listeners and references
5. **Debouncing**: Debounce rapid bid submissions