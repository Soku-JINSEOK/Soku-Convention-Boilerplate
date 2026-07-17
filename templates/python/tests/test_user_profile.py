from user_profile import User, UserResponse, to_user_response


def test_to_user_response_returns_only_public_profile_fields() -> None:
    user = User(
        id="user-123",
        email="reader@example.com",
        display_name="Reader Example",
        is_active=True,
    )

    assert to_user_response(user) == UserResponse(
        id="user-123",
        email="reader@example.com",
        display_name="Reader Example",
    )
