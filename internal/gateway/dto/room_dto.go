package dto

type CreateRoomHTTPRequest struct {
	Name       string `json:"name" validate:"required,min=3,max=100"`
	Visibility string `json:"visibility" validate:"required,oneof=public private"`
}

type JoinRoomByTokenHTTPRequest struct {
	InviteToken string `json:"invite_token" validate:"required"`
}

type RoomHTTPResponse struct {
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Visibility  string  `json:"visibility"`
	HostUUID    string  `json:"host_uuid"`
	InviteCode  *string `json:"invite_code,omitempty"`
	InviteLink  *string `json:"invite_link,omitempty"`
	MemberCount int     `json:"member_count"`
	OnlineCount int     `json:"online_count"`
}

type ListPublicRoomsHTTPResponse struct {
	Rooms      []RoomHTTPResponse `json:"rooms"`
	TotalItems int                `json:"total_items"`
}

type JoinPrivateRoomHTTPRequest struct {
	InviteCode  *string `json:"invite_code,omitempty"`
	InviteToken *string `json:"invite_token,omitempty"`
}

type UpdatePlaybackHTTPRequest struct {
	CurrentTrackID  string `json:"current_track_id"`
	PositionSeconds int    `json:"position_seconds"`
	IsPlaying       bool   `json:"is_playing"`
}

type PlaybackStateHTTPResponse struct {
	CurrentTrackID  string `json:"current_track_id"`
	PositionSeconds int    `json:"position_seconds"`
	IsPlaying       bool   `json:"is_playing"`
	UpdatedAt       string `json:"updated_at"`
}
