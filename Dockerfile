# docker build . -t unigrid/paxd:latest
# docker run --name paxd --rm -it unigrid/paxd:latest /bin/sh
# docker cp pax:/usr/bin/paxd /path/to/local/directory

# Use the Go alpine image as the base for the builder stage
FROM golang:1.21-alpine AS go-builder

SHELL ["/bin/sh", "-ecuxo", "pipefail"]

# Install necessary build dependencies
RUN apk add --no-cache ca-certificates build-base git

WORKDIR /code

# Add go.mod and go.sum to download and cache dependencies
ADD go.mod go.sum ./

# Dynamically determine the version of CosmWasm libwasmvm and download it
RUN set -eux; \
    ARCH=$(uname -m); \
    WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm | awk '{print $2}'); \
    if [ ! -z "${WASM_VERSION}" ]; then \
      wget -O /lib/libwasmvm_muslc.a https://github.com/CosmWasm/wasmvm/releases/download/${WASM_VERSION}/libwasmvm_muslc.${ARCH}.a; \
    fi; \
    go mod download

# Copy over the source code
COPY . /code/

# Build the project with static linking using the Makefile
# Ensure that the binary is statically linked
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make build \
  && file /code/bin/paxd \
  && echo "Ensuring binary is statically linked ..." \
  && (file /code/bin/paxd | grep "statically linked")

# Start a new stage for the final image
FROM alpine:3.16

# Copy the built binary from the builder stage
COPY --from=go-builder /code/bin/paxd /usr/bin/paxd

# Optionally copy any additional files or scripts required by your application
# COPY docker/* /opt/
# RUN chmod +x /opt/*.sh

WORKDIR /opt

# Expose necessary ports (adjust these according to your application's needs)
EXPOSE 1317 26656 26657

# Define the default command
CMD ["/usr/bin/paxd", "version"]
