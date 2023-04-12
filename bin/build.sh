#!/usr/bin/bash
GOODPATH="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd $GOODPATH
cd ../

# Screenshot
cat cmd/main.go | snippit --syntax go --out README.png

# Build
go mod tidy
go build -ldflags "-s -w" -o bin/gophre cmd/main.go

# Service
boum .
service gophre stop
service gophre start

message="$1"
if [[ $# -eq 0 ]]
then
	echo "Done."
else
	gigit $message
	echo "Updated."
fi

