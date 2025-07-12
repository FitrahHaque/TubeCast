#!/usr/bin/env bash
# usage: ./add-video.sh --title="A" --description="Show description" --video-url="https://youtu.be/xvFZjo5PgG0?si=-BV8fIKLdQDzdBJO"
docker compose run --rm tubecast -add-video "$@"