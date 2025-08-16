#!/bin/bash
set -e

SESSION="tdl"

# === CHECK FOR EXISTING SESSION ===
if tmux has-session -t "$SESSION" 2>/dev/null; then
	echo "Session $SESSION already exists. Attaching..."
	tmux attach-session -t "$SESSION"
	exit 0
fi

# === BACKUP START ===
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="$PARENT_DIR/tdl-backup"

echo "Creating backup at $BACKUP_DIR..."

[ -d "$BACKUP_DIR" ] && rm -rf "$BACKUP_DIR"
cp -r "$SCRIPT_DIR" "$BACKUP_DIR"

echo "Backup complete."
# === BACKUP END ===

# === TMUX SESSION SETUP ===
tmux new-session -d -s "$SESSION" -n root

# Add windows
tmux new-window -t "$SESSION" -n samples
tmux new-window -t "$SESSION" -n main
tmux new-window -t "$SESSION" -n codes
tmux new-window -t "$SESSION" -n build
tmux new-window -t "$SESSION" -n git

# Preload commands
tmux send-keys -t "$SESSION:main" "nvim main.go" C-m
tmux send-keys -t "$SESSION:codes" "cd codes" C-m
tmux send-keys -t "$SESSION:codes" "nvim *" C-m
tmux send-keys -t "$SESSION:samples" "cd samples" C-m
tmux send-keys -t "$SESSION:samples" "nvim *" C-m
tmux send-keys -t "$SESSION:build" "clear && go build . && ./tdl | less"

# Focus on root
tmux select-window -t "$SESSION:root"

# Attach to session
tmux attach-session -t "$SESSION"

