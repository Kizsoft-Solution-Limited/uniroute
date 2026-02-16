#!/bin/bash

# UniRoute Release Script
# Creates a new release tag and triggers GitHub Actions workflow

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}UniRoute Release Script${NC}"
echo ""

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}Error: Not in a git repository${NC}"
    exit 1
fi

# Check if there are uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${YELLOW}Warning: You have uncommitted changes${NC}"
    echo "It's recommended to commit or stash changes before creating a release."
    read -p "Continue anyway? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Get current version from git tags
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo -e "${BLUE}Current version: ${CURRENT_VERSION}${NC}"
echo ""

# Parse version number
if [[ $CURRENT_VERSION =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    MAJOR=${BASH_REMATCH[1]}
    MINOR=${BASH_REMATCH[2]}
    PATCH=${BASH_REMATCH[3]}
else
    MAJOR=0
    MINOR=0
    PATCH=0
fi

# Ask for release type
echo "What type of release?"
echo "1) Patch (${MAJOR}.${MINOR}.$((PATCH + 1))) - Bug fixes, small changes"
echo "2) Minor (${MAJOR}.$((MINOR + 1)).0) - New features, backward compatible"
echo "3) Major ($((MAJOR + 1)).0.0) - Breaking changes"
echo "4) Custom version"
echo ""
read -p "Choose option (1-4): " -n 1 -r
echo ""

case $REPLY in
    1)
        NEW_VERSION="v${MAJOR}.${MINOR}.$((PATCH + 1))"
        ;;
    2)
        NEW_VERSION="v${MAJOR}.$((MINOR + 1)).0"
        ;;
    3)
        NEW_VERSION="v$((MAJOR + 1)).0.0"
        ;;
    4)
        read -p "Enter version (e.g., v1.2.3): " NEW_VERSION
        if [[ ! $NEW_VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo -e "${RED}Error: Invalid version format. Use vX.Y.Z (e.g., v1.2.3)${NC}"
            exit 1
        fi
        ;;
    *)
        echo -e "${RED}Invalid option${NC}"
        exit 1
        ;;
esac

# Ask for release message
echo ""
read -p "Release message (optional): " RELEASE_MESSAGE
if [ -z "$RELEASE_MESSAGE" ]; then
    RELEASE_MESSAGE="Release ${NEW_VERSION}"
fi

# Confirm
echo ""
echo -e "${YELLOW}Ready to create release:${NC}"
echo -e "  Version: ${GREEN}${NEW_VERSION}${NC}"
echo -e "  Message: ${RELEASE_MESSAGE}"
echo ""
read -p "Create and push tag? (y/n): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled"
    exit 0
fi

# Create tag
echo ""
echo -e "${BLUE}Creating tag ${NEW_VERSION}...${NC}"
git tag -a "${NEW_VERSION}" -m "${RELEASE_MESSAGE}"

# Push tag
echo -e "${BLUE}Pushing tag to remote...${NC}"
git push origin "${NEW_VERSION}"

echo ""
echo -e "${GREEN}Release ${NEW_VERSION} created and pushed${NC}"
echo ""
echo "The GitHub Actions workflow will now:"
echo "  1. Build CLI binaries for all platforms"
echo "  2. Create GitHub release ${NEW_VERSION}"
echo "  3. Attach binaries to the release"
echo ""
echo "Monitor progress at:"
echo "  https://github.com/Kizsoft-Solution-Limited/uniroute/actions"
echo ""
echo "Once complete, download links will be available at:"
echo "  https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest"

