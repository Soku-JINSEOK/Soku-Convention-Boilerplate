#!/usr/bin/env bash
# scripts/setup-git-signing.sh
# Diagnostic and helper tool to configure Git Commit & Tag signing (Verified status).
set -euo pipefail

echo "============================================="
echo "⚙️ Git Commit & Tag Signing Setup Helper"
echo "============================================="

# 1. Check current setup
current_key=$(git config --get user.signingkey || true)
current_format=$(git config --get gpg.format || true)
gpg_sign_enabled=$(git config --get commit.gpgsign || true)

echo "Current Configuration:"
if [[ -n "$current_key" ]]; then
  echo " - Signing Key: $current_key"
  echo " - GPG Format: ${current_format:-gpg (default)}"
  echo " - Auto Sign Commits: ${gpg_sign_enabled:-false}"
else
  echo " - No signing key configured."
fi
echo "---------------------------------------------"

# 2. Find public keys
echo "Searching for local SSH public keys..."
ssh_keys=()
if [[ -d "$HOME/.ssh" ]]; then
  while IFS= read -r key; do
    ssh_keys+=("$key")
  done < <(find "$HOME/.ssh" -name "*.pub" -type f 2>/dev/null || true)
fi

if [[ ${#ssh_keys[@]} -gt 0 ]]; then
  echo "Found the following SSH public keys:"
  for i in "${!ssh_keys[@]}"; do
    echo "  [$i] ${ssh_keys[$i]}"
  done
else
  echo "No SSH public keys found in ~/.ssh/."
fi
echo "---------------------------------------------"

# 3. Setup Choice
echo "How would you like to set up Git Signing?"
echo "  [1] Set up SSH key signing (Recommended)"
echo "  [2] Set up GPG key signing"
echo "  [3] Exit"
read -rp "Enter choice [1-3]: " choice

case "$choice" in
  1)
    if [[ ${#ssh_keys[@]} -eq 0 ]]; then
      echo "Cannot configure SSH signing automatically because no public keys were found."
      echo "Please generate an SSH key first: ssh-keygen -t ed25519"
      exit 1
    fi
    read -rp "Select SSH key index to use [0-$(( ${#ssh_keys[@]} - 1 ))]: " key_idx
    if ! [[ "$key_idx" =~ ^[0-9]+$ ]] || [[ "$key_idx" -lt 0 ]] || [[ "$key_idx" -ge ${#ssh_keys[@]} ]]; then
      echo "Invalid index."
      exit 1
    fi
    selected_key="${ssh_keys[$key_idx]}"
    echo "Configuring Git to use: $selected_key"
    
    # Configure SSH format
    git config --global gpg.format ssh
    git config --global user.signingkey "$selected_key"
    git config --global commit.gpgsign true
    git config --global tag.gpgsign true
    
    echo "✅ Git globally configured for SSH signing!"
    echo "👉 IMPORTANT: Ensure this public key is added to your GitHub account"
    echo "   under Settings -> SSH and GPG keys -> New SSH Key (Key type: 'Signing Key')."
    echo "   Key content:"
    cat "$selected_key"
    ;;
  2)
    echo "To configure GPG signing:"
    echo "1. Run: gpg --full-generate-key"
    echo "2. Find Key ID: gpg --list-secret-keys --keyid-format=long"
    echo "3. Run: git config --global user.signingkey <KEY_ID>"
    echo "4. Run: git config --global commit.gpgsign true"
    echo "5. Run: git config --global tag.gpgsign true"
    echo "6. Export key to GitHub: gpg --armor --export <KEY_ID>"
    ;;
  *)
    echo "Exiting."
    ;;
esac
