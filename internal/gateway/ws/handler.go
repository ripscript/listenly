package ws

import (
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/labstack/echo/v5"

	authv1 "listenly-backend/gen/go/auth/v1"
)

type Handler struct {
	hub        *Hub
	authClient authv1.AuthServiceClient
}

func NewHandler(hub *Hub, authClient authv1.AuthServiceClient) *Handler {
	return &Handler{hub: hub, authClient: authClient}
}

// Connect menangani upgrade HTTP->WS untuk sebuah room.
// Auth lewat query param ?token=... karena browser WebSocket API
// tidak bisa mengirim header Authorization custom.
func (h *Handler) Connect(c *echo.Context) error {
	roomUUID := c.Param("uuid")

	token := c.QueryParam("token")
	if token == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
	}

	// verifikasi token ke auth-service (pola sama seperti JWT middleware)
	result, err := h.authClient.VerifyToken(c.Request().Context(), &authv1.VerifyTokenRequest{
		AccessToken: token,
	})
	if err != nil || !result.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
	}

	// upgrade ke WebSocket
	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{
		// untuk dev lokal; di produksi batasi OriginPatterns ke domain frontend Anda
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}

	h.hub.Add(roomUUID, conn)
	defer func() {
		h.hub.Remove(roomUUID, conn)
		conn.Close(websocket.StatusNormalClosure, "")
	}()

	// kirim pesan sambutan agar client tahu koneksi siap
	ctx := c.Request().Context()
	_ = wsjson.Write(ctx, conn, map[string]any{
		"type":      "connected",
		"room_uuid": roomUUID,
	})

	// loop baca: kita tidak mengharapkan pesan dari client (ini kanal push 1 arah),
	// tapi loop ini WAJIB ada untuk mendeteksi disconnect & merespons ping/pong.
	for {
		if _, _, err := conn.Read(ctx); err != nil {
			return nil // client disconnect -> keluar, defer membersihkan dari Hub
		}
	}
}
