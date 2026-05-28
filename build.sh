
#!/bin/bash

cd $(dirname $0)

source ./.env
make frontend
make build_packages
# built under ./bin
