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

### Compiling the Daemon

To compile the `paxd` daemon, use the following command:

```bash
ignite chain build
```

This command will compile the daemon with the latest updates and the specified version tag.

After compiling you can check the version was correctly added.
```bash
paxd version
```

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
# add some addresses that you have private keys for (locally) to give them genesis funds
docker run --rm -it \
    -e PASSWORD=xxxxxxxx \
    --mount type=volume,source=paxd_data,target=/root \
    unigrid/paxd:latest /opt/setup_and_run.sh unigrid1pkptre7fdkl6gfrzlesjjvhxhlc3r4gmmk8rs6

# only perform setup
docker run --rm -it \
    -e PASSWORD=xxxxxxxx \
    --mount type=volume,source=paxd_data,target=/root \
    unigrid/paxd:latest /opt/setup_paxd.sh unigrid1pkptre7fdkl6gfrzlesjjvhxhlc3r4gmmk8rs6

# This will start both paxd and rest-server, both are logged
docker run --rm -it -p 26657:26657 -p 26656:26656 -p 1317:1317 \
    --mount type=volume,source=paxd_data,target=/root \
    unigrid/paxd:latest /opt/run_paxd.sh
```
