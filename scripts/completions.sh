#!/bin/sh
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
  go run ./cmd/kh completion "$sh" > "completions/kh.$sh"
done
