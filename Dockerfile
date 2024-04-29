# docker build . -t unigrid/paxd:latest
# docker run --name paxd --rm -it unigrid/paxd:latest /bin/sh
# docker cp paxd:/usr/bin/paxd /path/to/local/directory

# Use the Go alpine image as the base for the builder stage
FROM golang:1.21-alpine AS go-builder

SHELL ["/bin/sh", "-ecuxo", "pipefail"]

# Install necessary build dependencies
RUN apk add --no-cache ca-certificates build-base git
WORKDIR /code

# Copy the entire current directory to /code
COPY . /code/

# Display detailed list of all files and directories copied
RUN ls -la /code

# Specifically check for the .git directory
RUN if [ -d "/code/.git" ]; then \
        echo ".git directory exists"; \
        ls -la /code/.git; \
    else \
        echo ".git directory does NOT exist"; \
    fi

# Add go.mod and go.sum to download and cache dependencies
ADD go.mod go.sum ./

# Dynamically determine the version of CosmWasm libwasmvm and download it
RUN set -eux; \
    ARCH=$(uname -m); \
    WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm | awk '{print $2}'); \
    if [ ! -z "${WASM_VERSION}" ]; then \
      echo "Downloading libwasmvm_muslc for ARCH=${ARCH} and WASM_VERSION=${WASM_VERSION}"; \
      DOWNLOAD_URL="https://github.com/CosmWasm/wasmvm/releases/download/${WASM_VERSION}/libwasmvm_muslc.${ARCH}.a"; \
      echo "Download URL: ${DOWNLOAD_URL}"; \
      wget -O /lib/libwasmvm_muslc.a ${DOWNLOAD_URL}; \
    fi; \
    go mod download

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

# Copy the necessary scripts and configuration files
COPY docker/* /opt/
RUN chmod +x /opt/*.sh

COPY contracts/artifacts /opt/artifacts/

WORKDIR /opt

# Expose necessary ports (adjust these according to your application's needs)
EXPOSE 1317 26656 26657 9090

# Define the default command
CMD ["/usr/bin/paxd", "version"]
