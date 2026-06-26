export type User = {
  id: string;
  email: string;
  displayName: string;
  isActive: boolean;
};

export type UserResponse = {
  id: string;
  email: string;
  displayName: string;
};

export function toUserResponse(user: User): UserResponse {
  return {
    id: user.id,
    email: user.email,
    displayName: user.displayName,
  };
}
