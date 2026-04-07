#!/bin/sh

echo "Fetching latest tags..."
git fetch --tags

LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo "Current version: $LATEST_TAG"
echo ""

echo "Select version bump type:"
echo "  1) major (breaking changes)"
echo "  2) minor (new features, backward compatible)"
echo "  3) patch (bug fixes)"
read -p "Enter choice (1/2/3): " BUMP_TYPE
echo ""

VERSION=$(echo $LATEST_TAG | sed 's/v//')
MAJOR=$(echo $VERSION | cut -d. -f1)
MINOR=$(echo $VERSION | cut -d. -f2)
PATCH=$(echo $VERSION | cut -d. -f3)

case $BUMP_TYPE in
	1)
		MAJOR=$((MAJOR + 1))
		MINOR=0
		PATCH=0
		NEW_TAG="v$MAJOR.$MINOR.$PATCH"
		echo "WARNING: Major version bump indicates BREAKING CHANGES!"
		echo ""
		;;
	2)
		MINOR=$((MINOR + 1))
		PATCH=0
		NEW_TAG="v$MAJOR.$MINOR.$PATCH"
		;;
	3)
		PATCH=$((PATCH + 1))
		NEW_TAG="v$MAJOR.$MINOR.$PATCH"
		;;
	*)
		echo "Invalid choice. Exiting."
		exit 1
		;;
esac

echo "New version will be: $NEW_TAG"
echo ""
read -p "Create and push tag $NEW_TAG? (yes/no): " CONFIRM

case $CONFIRM in
	yes|y|Y|YES)
		echo "Creating tag $NEW_TAG..."
		git tag -a $NEW_TAG -m "Release $NEW_TAG"
		echo "Pushing tag to GitHub..."
		git push origin $NEW_TAG
		echo "Tag $NEW_TAG created and pushed successfully!"
		;;
	*)
		echo "Tag creation cancelled."
		exit 1
		;;
esac
