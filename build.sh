
#!/bin/bash

cd $(dirname $0)

source ./.env
make frontend
make build_release
make build_packages
# built under ./bin
