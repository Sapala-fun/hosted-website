#!/bin/bash
# Theme Switcher - Toggle between classic and modern templates
# Usage: ./scripts/switch-theme.sh [classic|modern]

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

THEME_DIR="$ROOT_DIR/frontend/src/styles"
CONFIG_FILE="$ROOT_DIR/frontend/src/config/theme.ts"
THEME_FILE="$THEME_DIR/theme.css"

if [ $# -eq 0 ]; then
  # Show current theme
  CURRENT=$(grep "THEME = '" "$CONFIG_FILE" | sed "s/.*THEME = '\(.*\)'.*/\1/")
  echo "Current theme: $CURRENT"
  echo "Available themes: classic, modern"
  echo ""
  echo "Usage: $0 [classic|modern]"
  exit 0
fi

TARGET="$1"

if [[ "$TARGET" != "classic" && "$TARGET" != "modern" ]]; then
  echo "Error: Theme must be 'classic' or 'modern'"
  exit 1
fi

# Update config file
sed -i '' "s/THEME = '[^']*'/THEME = '$TARGET'/" "$CONFIG_FILE"

# Copy the target theme's CSS to theme.css (works everywhere, no symlink issues)
cp "$THEME_DIR/${TARGET}.css" "$THEME_FILE"

echo "✓ Switched to '$TARGET' theme"
echo "  Config: $CONFIG_FILE → THEME = '$TARGET'"
echo "  Styles: $THEME_FILE updated from ${TARGET}.css"
