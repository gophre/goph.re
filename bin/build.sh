#!/usr/bin/bash
GOODPATH="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd $GOODPATH
cd ../

# Screenshot
go mod tidy
go build -ldflags "-s -w" -o bin/gophre cmd/main.go
echo "Compiled."

# Service
boum .
service gophre stop
service gophre start

message="$1"
if [[ $# -eq "" ]]
then
	echo "Done."
else
	gigit "$message"
	echo "Updated."
fi

