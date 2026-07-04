package profile

import "testing"

func TestToUserResponseReturnsOnlyPublicProfileFields(t *testing.T) {
	user := User{
		ID:          "user-123",
		Email:       "reader@example.com",
		DisplayName: "Reader Example",
		IsActive:    true,
	}

	want := UserResponse{
		ID:          "user-123",
		Email:       "reader@example.com",
		DisplayName: "Reader Example",
	}

	got := ToUserResponse(user)
	if got != want {
		t.Errorf("ToUserResponse(%+v) = %+v, want %+v", user, got, want)
	}
}
