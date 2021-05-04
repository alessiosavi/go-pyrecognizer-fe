#!/bin/bash
podman-compose down
podman build -t go-pyrecognizer-fe .
podman-compose up
