#!/bin/bash

GOLANG_GEN_DIR=$1
PROTOS_DIR=$2

for VersionDir in "$PROTOS_DIR"/*; do
    API_VERSION=$(basename "${VersionDir}")

    for file in "$VersionDir"/*.proto; do

        SERVICE=$(basename "${file%.*}" .proto)

        dir="$GOLANG_GEN_DIR/$API_VERSION/$SERVICE"
        if ! [[ -d $dir ]]; then
            mkdir -p $dir
        fi

        protoc -I $PROTOS_DIR/$API_VERSION "$PROTOS_DIR/$API_VERSION/$SERVICE.proto" --go_out=$dir --go_opt=paths=source_relative --go-grpc_out=$dir --go-grpc_opt=paths=source_relative
        echo "Generating SERVICE: $SERVICE version: $API_VERSION Success"

    done
done
