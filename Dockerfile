# Make a dockerfile for the golang-ms-template
FROM --platform=linux/amd64 golang:1.22.3 AS builder

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /build


# New stage
FROM --platform=linux/amd64 scratch

# Copy the binary from the build stage
COPY --from=builder /build /build

# Optional:
EXPOSE 3000

# Run
CMD ["/build"]