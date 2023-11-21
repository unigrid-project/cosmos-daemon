## Build Process for Paxd Daemon

This document outlines the steps to compile the `paxd` daemon using Ignite. The process includes updating modules, setting tags for new builds, and the actual compilation.

### Updating Modules

If there's a new module update available, pull the latest version using the following command:

```bash
go get github.com/unigrid-project/cosmos-sdk-gridnode@latest
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


