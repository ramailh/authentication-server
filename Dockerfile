################################
# STEP 1 build executable binary
################################
FROM golang:alpine AS builder

RUN apk add --update --no-cache ca-certificates git
ENV GO111MODULE=on

# Set work directory before compile
RUN  mkdir  /api
WORKDIR /api/
COPY go.mod .
COPY go.sum .

# Fetch dependencies.
RUN go mod download

# Copy all Environtment
COPY . .

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /api/auth-server

############################
# STEP 2 build a small image
############################
FROM alpine

RUN apk add --update --no-cache ca-certificates git

# Copy our static executable.
COPY --from=builder /api/auth-server /api/auth-server
# Run the binary.
EXPOSE 9901
ENTRYPOINT ["/api/auth-server"]
