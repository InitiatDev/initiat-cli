#!/bin/bash

# Setup git hooks for Initiat CLI development

set -e

echo "🔧 Setting up git hooks..."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository"
    exit 1
fi

# Get the git hooks directory
HOOKS_DIR="$(git rev-parse --git-dir)/hooks"

# Create pre-commit hook
echo "📝 Creating pre-commit hook..."
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash
# Auto-generated pre-commit hook for Initiat CLI
exec ./scripts/pre-commit.sh
EOF

# Make the hook executable
chmod +x "$HOOKS_DIR/pre-commit"

echo "✅ Git hooks installed successfully!"
echo ""
echo "The pre-commit hook will now run automatically before each commit."
echo "It will:"
echo "  • Format your code"
echo "  • Run the linter"
echo "  • Run tests"
echo "  • Check for security issues"
echo "  • Verify the build works"
echo ""
echo "To skip the hook for a specific commit, use: git commit --no-verify"
