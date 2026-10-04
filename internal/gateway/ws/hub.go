package ws

import (
	"sync"

	"github.com/coder/websocket"
)

// Hub menyimpan koneksi WebSocket aktif, dikelompokkan per room UUID.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*websocket.Conn]struct{})}
}

func (h *Hub) Add(roomUUID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[roomUUID] == nil {
		h.rooms[roomUUID] = make(map[*websocket.Conn]struct{})
	}
	h.rooms[roomUUID][conn] = struct{}{}
}

func (h *Hub) Remove(roomUUID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.rooms[roomUUID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.rooms, roomUUID)
		}
	}
}

// Connections mengembalikan salinan koneksi di sebuah room (aman untuk iterasi).
func (h *Hub) Connections(roomUUID string) []*websocket.Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns := h.rooms[roomUUID]
	out := make([]*websocket.Conn, 0, len(conns))
	for c := range conns {
		out = append(out, c)
	}
	return out
}
