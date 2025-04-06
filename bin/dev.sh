#!/usr/bin/bash
GOODPATH="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd $GOODPATH
cd ../

# Screenshot
go mod tidy
service gophre stop
go run cmd/main.go serve
echo "Compiled."

# Service
boum .
service gophre start

message="$1"
if [[ $# -eq "" ]]
then
	echo "Done."
else
	gigit $message
	echo "Updated."
fi

