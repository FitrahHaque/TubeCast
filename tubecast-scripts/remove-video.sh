#!/usr/bin/env bash
# usage: ./remove-video.sh --title="A" --video-url="https://youtu.be/xvFZjo5PgG0?si=-BV8fIKLdQDzdBJO"
docker compose run --rm tubecast -remove-video "$@"