#!/bin/sh
set -ex

DOCKERFILE=Dockerfile
IMAGE=ssegpt
TAG=0.0.1
CONTEXT_DIR=$(dirname "$DOCKERFILE")
PROJECT=ssgpt

# Wait for the Docker daemon to be available.
until podman ps
do sleep 3
done

echo "Project: "$PROJECT
cd $CONTEXT_DIR
echo "Building image $IMAGE:$TAG"
podman build . -t $IMAGE:$TAG
podman tag $IMAGE:$TAG $IMAGE:latest
echo "Push image $IMAGE:$TAG"
podman push $IMAGE:$TAG docker://docker.io/megafyk/$IMAGE:$TAG
podman push $IMAGE:latest docker://docker.io/megafyk/$IMAGE:latest
echo "Finished building image"