#!/bin/bash

# Drunken Dolphin Personal Assistant Examples
# This script demonstrates various ways to use the personal CLI assistant

echo "🐬 Drunken Dolphin Personal Assistant Examples"
echo "=============================================="
echo

echo "📦 Installation:"
echo "   cargo install --path ."
echo "   # Don't forget to add ~/.cargo/bin to your PATH!"
echo "   # Fish: echo 'set -gx PATH \$HOME/.cargo/bin \$PATH' >> ~/.config/fish/config.fish"
echo "   # Zsh:  echo 'export PATH=\"\$HOME/.cargo/bin:\$PATH\"' >> ~/.zshrc"
echo

echo "1. Record push-ups for today:"
echo "   ./target/release/drunken-dolphin fitness pushups 25"
echo

echo "2. Record push-ups for yesterday:"
echo "   ./target/release/drunken-dolphin fitness pushups 20 --date yesterday"
echo

echo "3. Record push-ups for a specific date:"
echo "   ./target/release/drunken-dolphin fitness pushups 35 --date 2024-01-15"
echo

echo "4. Record sit-ups for today:"
echo "   ./target/release/drunken-dolphin fitness situps 50"
echo

echo "5. Record sit-ups for yesterday:"
echo "   ./target/release/drunken-dolphin fitness situps 40 --date yesterday"
echo

echo "6. Record sit-ups for a specific date:"
echo "   ./target/release/drunken-dolphin fitness situps 55 --date 2024-01-15"
echo

echo "7. Quick shortcut - record both for today:"
echo "   ./target/release/drunken-dolphin fitness today 25 50"
echo

echo "8. Quick shortcut - record both for yesterday:"
echo "   ./target/release/drunken-dolphin fitness yesterday 30 45"
echo

echo "9. Chore - add a task (coming soon):"
echo "    ./target/release/drunken-dolphin chore add \"Clean the garage\" --priority high"
echo

echo "10. Chore - mark as complete (coming soon):"
echo "    ./target/release/drunken-dolphin chore complete \"Clean the garage\""
echo

echo "11. Check CLI version and help:"
echo "    ./target/release/drunken-dolphin --version"
echo "    ./target/release/drunken-dolphin --help"
echo "    ./target/release/drunken-dolphin fitness --help"
echo "    ./target/release/drunken-dolphin chore --help"
echo

echo "12. Use custom data file location:"
echo "    ./target/release/drunken-dolphin --data-file ~/my_personal_data.json fitness pushups 25"
echo

echo "Note: Data is automatically saved to personal_data.json (or your custom file)"
echo "      Dates can be: 'today', 'yesterday', or YYYY-MM-DD format"
echo "      Use 'fitness today' and 'fitness yesterday' commands for quick recording of both exercises"
echo "      Chore commands are coming soon!"
echo
echo "💡 Installation Tip: After 'cargo install --path .', add ~/.cargo/bin to your PATH"
echo "   Fish: source ~/.config/fish/config.fish"
echo "   Zsh:  source ~/.zshrc"
