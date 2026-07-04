from dataclasses import dataclass


@dataclass
class User:
    id: str
    email: str
    display_name: str
    is_active: bool


@dataclass
class UserResponse:
    id: str
    email: str
    display_name: str


def to_user_response(user: User) -> UserResponse:
    return UserResponse(
        id=user.id,
        email=user.email,
        display_name=user.display_name,
    )
