#!/usr/bin/env bash
# usage: ./sync-channel.sh --title="A" --channel-id="ThePrimeTimeagen"
docker compose run --rm tubecast -sync-channel "$@"