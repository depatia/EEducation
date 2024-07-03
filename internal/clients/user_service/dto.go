package user

type RegisterReq struct {
	Email    string `json:"email"  bson:"email"`
	Password string `json:"password"  bson:"password"`
}

type LoginReq struct {
	Email    string `json:"email"  bson:"email"`
	Password string `json:"password"  bson:"password"`
}

type UpdatePasswordReq struct {
	Email       string `json:"email"  bson:"email"`
	OldPassword string `json:"old_password"  bson:"old_password"`
	NewPassword string `json:"new_password"  bson:"new_password"`
	Token       string `json:"token"  bson:"token"`
}

type SetPermissionLevelReq struct {
	UserID          int64 `json:"user_id"  bson:"user_id"`
	PermissionLevel int64 `json:"permission_level"  bson:"permission_level"`
	InitiatorID     int64 `json:"initiator_id"  bson:"initiator_id"`
}

type GetPermissionLevelReq struct {
	UserID int64 `json:"user_id"  bson:"user_id"`
}

type FillUserProfileReq struct {
	Name        string `json:"name"  bson:"name"`
	Lastname    string `json:"lastname"  bson:"lastname"`
	MiddleName  string `json:"middle_name"  bson:"middle_name"`
	DateOfBirth string `json:"date_of_birth"  bson:"date_of_birth"`
	Classname   string `json:"classname"  bson:"classname"`
	UserID      int64  `json:"user_id"  bson:"user_id"`
}

type ChangeUserStatusReq struct {
	UserID      int64 `json:"user_id"  bson:"user_id"`
	InitiatorID int64 `json:"initiator_id"  bson:"initiator_id"`
	Active      bool  `json:"active"  bson:"active"`
}

type IsUserActiveReq struct {
	UserID int64 `json:"user_id"  bson:"user_id"`
}
