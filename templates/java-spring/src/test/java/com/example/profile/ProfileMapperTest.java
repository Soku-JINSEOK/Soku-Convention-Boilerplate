package com.example.profile;

import static org.junit.jupiter.api.Assertions.assertEquals;

import org.junit.jupiter.api.Test;

class ProfileMapperTest {

    @Test
    void toUserResponseReturnsOnlyPublicProfileFields() {
        User user = new User("user-123", "reader@example.com", "Reader Example", true);

        UserResponse response = ProfileMapper.toUserResponse(user);

        assertEquals(new UserResponse("user-123", "reader@example.com", "Reader Example"), response);
    }
}
