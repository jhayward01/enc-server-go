# syntax=docker/dockerfile:1

FROM golang:1.24

# Set destination for COPY
WORKDIR /app

# Copy repo
COPY ../.. ./

# Build
RUN make install-servers
