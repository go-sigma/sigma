---
title: Apptainer
description: Push apptainer sif to sigma
---

# Apptainer push to sigma

### Run apptainer container

Apptainer just run on linux system, so we can run it in container like this:

``` bash
docker run -it --rm ghcr.io/apptainer/apptainer:1.3.0-rc.1 bash
```

### Pull a image and save it as a sif file.

``` bash
apptainer build alpine.sif docker://alpine
```

### Login and push sif file

``` bash
apptainer registry login -u sigma docker://192.168.31.112:3000
apptainer push alpine.sif oras://192.168.31.112:3000/library/alpine:tosone
```
