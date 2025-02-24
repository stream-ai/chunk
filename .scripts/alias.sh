#!/bin/zsh

eval "$(dircolors -b)"
export LS_OPTIONS='--color=auto'
alias ls='ls $LS_OPTIONS'
alias l='ls $LS_OPTIONS -lA'
alias ll='ls $LS_OPTIONS -laFh '

alias gotest='$(git rev-parse --show-toplevel)/.scripts/gotest.sh "$@"'
alias gorun='while inotifywait -r -e close_write /workspace/monorepo ; do go run . | jq '.'; done'
alias xclip='socat - tcp:host.docker.internal:8121'
alias srctree='tree -d -a -I "cdk.out" -I "node_modules" -I ".git" -I ".angular" -I "dist"'
alias rt='git rev-parse --show-toplevel'
alias lgd='unbuffer aws logs tail "/aws/lambda/huddle-websocket-default" --follow --format short |  unbuffer -p cut -d " " -f 2- | loggo stream -t /workspace/loggo.fmt'

