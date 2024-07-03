package models

type User struct {
	ID              int64  `json:"id"`
	Email           string `json:"email"`
	PassHash        string `json:"password"`
	PermissionLevel int64  `json:"permission_level"`
}

type UserInfo struct {
	ID          int64
	Name        string
	Lastname    string
	Middlename  string
	DateOfBirth string
	Classname   string
}
