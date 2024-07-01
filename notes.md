## Build Process for Paxd Daemon

This document outlines the steps to compile the `paxd` daemon using Ignite. The process includes updating modules, setting tags for new builds, and the actual compilation.

### Updating Modules

If there's a new module update available, pull the latest version using the following command:

```bash
go get github.com/unigrid-project/cosmos-gridnode@latest
```

This command fetches the latest version of the specified module.

### Setting Tags for a New Build

Before compiling a new build, it's important to tag the version. This ensures that the compiled daemon is correctly versioned. Use the following commands to tag your build:

```bash
git tag v0.0.1
git push origin v0.0.1
```

Replace `v0.0.1` with the appropriate version number for your new build.

### Dev server

Bring up a local node with a test account containing tokens

This is just designed for local testing/CI - do not use these scripts in production.
Very likely you will assign tokens to accounts whose mnemonics are public on github.

### Build the container with latest code
```sh
docker build --no-cache . -t unigrid/paxd:latest
```
Remove the volume to reset data.
```sh
docker volume rm -f paxd_data

# pass password (one time) as env variable for setup, so we don't need to keep typing it
# HEDGEHOG_URL change this if you are running hedgehog on a different port 
# or would like to point it to a different location 
# testnet hedgehog is https://149.102.147.45:39886
# add some addresses that you have private keys for (locally) to give them genesis funds
docker run --rm -it \
    --name paxd \
    -p 26657:26657 -p 26656:26656 -p 1317:1317 \
    -e PASSWORD=xxxxxxxx \
    -e HEDGEHOG_URL=https://127.0.0.1:40005 \
    --mount type=volume,source=paxd_data,target=/root \
    unigrid/paxd:latest /opt/setup_and_run.sh unigrid1pkptre7fdkl6gfrzlesjjvhxhlc3r4gmmk8rs6

# only perform setup
docker run --rm -it \
    -e PASSWORD=xxxxxxxx \
    -e HEDGEHOG_URL=https://127.0.0.1:40005 \
    --mount type=volume,source=paxd_data,target=/root \
    unigrid/paxd:latest /opt/setup_paxd.sh unigrid1pkptre7fdkl6gfrzlesjjvhxhlc3r4gmmk8rs6

# This will start both paxd and rest-server, both are logged
docker run --rm -it -p 26657:26657 -p 26656:26656 -p 1317:1317 \
    --name paxd \
    -e HEDGEHOG_URL=https://127.0.0.1:40005 \
    --mount type=volume,source=paxd_data,target=/root \
    unigrid/paxd:latest /opt/run_paxd.sh
```

## Copy over the paxd daemon

```bash
docker cp paxd:/usr/bin/paxd /path/to/local/directory
```

