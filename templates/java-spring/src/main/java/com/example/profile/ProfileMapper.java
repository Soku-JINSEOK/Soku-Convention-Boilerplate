package com.example.profile;

public final class ProfileMapper {

    private ProfileMapper() {
    }

    public static UserResponse toUserResponse(User user) {
        return new UserResponse(user.getId(), user.getEmail(), user.getDisplayName());
    }
}
