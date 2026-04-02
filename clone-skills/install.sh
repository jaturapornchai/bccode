#!/bin/bash
# Jead Skill Library — Auto Install
# Copy skills + references to multiple AI tool directories

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGETS=(".claude" ".agents" ".kilocode" ".kiro" ".agent")

echo "=== Jead Skill Library Installer ==="
echo "Source: $SCRIPT_DIR"
echo ""

for target in "${TARGETS[@]}"; do
    TARGET_DIR="$HOME/$target/skills"
    echo "Installing to $TARGET_DIR..."
    mkdir -p "$TARGET_DIR"

    # Copy skills
    if [ -d "$SCRIPT_DIR/skills" ]; then
        cp -r "$SCRIPT_DIR/skills/"* "$TARGET_DIR/" 2>/dev/null
    fi

    # Copy references
    REF_DIR="$HOME/$target/references"
    if [ -d "$SCRIPT_DIR/references" ]; then
        mkdir -p "$REF_DIR"
        cp -r "$SCRIPT_DIR/references/"* "$REF_DIR/" 2>/dev/null
    fi

    echo "  Done."
done

echo ""
echo "=== Installation complete ==="
echo "Skills installed to: ${TARGETS[*]}"
