# Use the official Golang image as the base for building the application
FROM golang:1.23.3-alpine3.20 AS builder

# Argument for the environment to be passed during the build
ARG ENV

# Update and upgrade the Alpine packages, then install 'make'
RUN apk update && apk upgrade --available && \
    apk add make && \
    # Create a new user 'auth' with specific parameters
    adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "10001" \
    "auth"

WORKDIR /opt/app/

# Copy the go.mod and go.sum first to install dependencies
COPY go.mod go.sum ./

# Download the Go module dependencies and verify them
RUN go mod download && go mod verify

# Copy the entire application code into the working directory
COPY . .

# Build the application using the 'make' command, passing the environment as a variable
RUN make build-app ENV=${ENV}

# Create a logs directory for the application (if needed at runtime)
RUN mkdir logs/

###########
# 2 stage #
###########
# Use a minimal base image to run the application
FROM scratch

# Argument for configuration file path to be passed during the build
ARG CONFIG

# Set the working directory in the new image
WORKDIR /opt/app/

# Copy the passwd and group files from the builder stage for the user 'auth'
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the compiled binary and configuration file from the builder stage
# Ensure the ownership is set to the 'auth' user and group
COPY --from=builder --chown=auth:auth /opt/app/bin/main .
COPY --from=builder --chown=auth:auth /opt/app/${CONFIG} ./config

# Set the user and group for running the application
USER auth:auth

# Command to run the application with the specified configuration file
CMD ["./main", "-config=./config"]
