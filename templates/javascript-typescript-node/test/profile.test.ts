import {describe, expect, it} from 'vitest';

import {toUserResponse, type User} from '../src/profile.js';

describe('toUserResponse', () => {
  it('returns only public profile fields', () => {
    const user: User = {
      id: 'user-123',
      email: 'reader@example.com',
      displayName: 'Reader Example',
      isActive: true,
    };

    expect(toUserResponse(user)).toEqual({
      id: 'user-123',
      email: 'reader@example.com',
      displayName: 'Reader Example',
    });
  });
});
