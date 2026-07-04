package com.example.profile;

public class User {
    private final String id;
    private final String email;
    private final String displayName;
    private final boolean isActive;

    public User(String id, String email, String displayName, boolean isActive) {
        this.id = id;
        this.email = email;
        this.displayName = displayName;
        this.isActive = isActive;
    }

    public String getId() {
        return id;
    }

    public String getEmail() {
        return email;
    }

    public String getDisplayName() {
        return displayName;
    }

    public boolean isActive() {
        return isActive;
    }
}
