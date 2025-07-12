#!/usr/bin/env bash
# usage: ./remove-show.sh --title="A"
docker compose run --rm tubecast -remove-show "$@"