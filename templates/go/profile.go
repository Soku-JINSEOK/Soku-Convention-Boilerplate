package profile

type User struct {
	ID          string
	Email       string
	DisplayName string
	IsActive    bool
}

type UserResponse struct {
	ID          string
	Email       string
	DisplayName string
}

func ToUserResponse(user User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}
}
