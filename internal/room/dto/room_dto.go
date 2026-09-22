package dto

type CreateRoomRequest struct {
	HostUUID   string
	Name       string
	Visibility string
}

type JoinRoomByCodeRequest struct {
	UserUUID   string
	InviteCode string
}

type JoinRoomByTokenRequest struct {
	UserUUID    string
	InviteToken string
}

type RoomResponse struct {
	UUID        string
	Name        string
	Visibility  string
	HostUUID    string
	InviteCode  *string
	InviteToken *string
	MemberCount int
	OnlineCount int
}

type JoinRoomRequest struct {
	RoomUUID string
	UserUUID string
}
