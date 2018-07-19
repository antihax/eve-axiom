#!/bin/bash
set +e
CGO_ENABLED=1 GOOS=linux go build -a --installsuffix cgo  --ldflags '-linkmode external -extldflags "-static"'  -o bin/eve-axiom ./cmd/
docker build -t antihax/eve-axiom .
