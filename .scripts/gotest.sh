#!/bin/zsh

root=$(git rev-parse --show-toplevel)
go_dirs="${root}/cdk ${root}/cmd ${root}/pkg ${root}/internal"

div="\n************************************************************\n"
while inotifywait -r -e modify,create,delete $root ; do echo $div && go test ./... "$@"; done
