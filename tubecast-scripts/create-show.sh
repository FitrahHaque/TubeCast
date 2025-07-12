#!/usr/bin/env bash
# usage: ./create-show.sh --title="A" --description="B"
docker compose run --rm tubecast -create-show "$@"