#!/bin/bash
# Stop and remove nexus-pod
echo "Stopping and removing Podman pod 'nexus-pod'..."
podman pod stop nexus-pod 2>/dev/null || true
podman pod rm -f nexus-pod 2>/dev/null || true
echo "Nexus pod stopped and cleaned up."
