#!/usr/bin/env bash
# scripts/create-release-tag.sh
# Prompts for a release version, verifies git state and signing, and creates a signed release tag.
set -euo pipefail

echo "============================================="
echo "🏷️ Verified Release Tag Creator"
echo "============================================="

# 1. Check branch
current_branch=$(git branch --show-current)
if [[ "$current_branch" != "main" && "$current_branch" != "master" ]]; then
  echo "⚠️ Warning: You are not on 'main' or 'master' branch. Current branch: $current_branch"
  read -rp "Do you want to continue anyway? (y/N): " confirm_branch
  if [[ "$confirm_branch" != "y" && "$confirm_branch" != "Y" ]]; then
    echo "Aborted."
    exit 1
  fi
fi

# 2. Check clean working tree
if ! git diff-index --quiet HEAD --; then
  echo "❌ Error: Your working tree has uncommitted changes. Please commit or stash them first."
  exit 1
fi

# 3. Check signing configuration
signing_key=$(git config --get user.signingkey || true)
if [[ -z "$signing_key" ]]; then
  echo "❌ Error: No Git signing key configured (user.signingkey)."
  echo "   To get a 'Verified' badge on GitHub, you must sign release tags."
  echo "   Please run: scripts/setup-git-signing.sh to configure signing first."
  exit 1
fi

# 4. Display recent tags
echo "Recent release tags:"
git tag -l "v*" --sort=-v:refname | head -n 5 || echo "No existing release tags found."
echo "---------------------------------------------"

# 5. Prompt for next version
read -rp "Enter next release version (e.g. v1.2.3): " next_version

# Validate tag format (vMAJOR.MINOR.PATCH)
if [[ ! "$next_version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "❌ Error: Version must follow the format 'vMAJOR.MINOR.PATCH' (e.g., v1.0.0)."
  exit 1
fi

# Check if tag already exists
if git rev-parse "$next_version" >/dev/null 2>&1; then
  echo "❌ Error: Tag $next_version already exists."
  exit 1
fi

# 6. Create signed tag
echo "Creating signed tag: $next_version"
if git tag -s -m "Release $next_version" "$next_version"; then
  echo "✅ Successfully created signed release tag: $next_version"
  echo "🚀 To push the tag to GitHub, run:"
  echo "   git push origin $next_version"
else
  echo "❌ Failed to create signed tag. Please check your GPG/SSH signing configuration."
  exit 1
fi
