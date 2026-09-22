package models

type User struct {
	BaseModel
	Email    string  `gorm:"type:varchar(255);uniqueIndex:idx_users_email;not null"`
	Password string  `gorm:"type:varchar(255);not null"`
	FullName *string `gorm:"type:varchar(100)"`

	RefreshTokens []RefreshToken `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (User) TableName() string {
	return "users"
}
