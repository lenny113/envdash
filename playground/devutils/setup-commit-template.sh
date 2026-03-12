#!/usr/bin/env sh

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT" || exit 1

git config commit.template playground/devutils/gitmessage.txt

echo "Commit template configured:"
git config --get commit.template