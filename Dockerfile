# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.12 as builder

# Copy local code to the container image.
WORKDIR /go/src/github.com/keremk/challenge-bot
COPY go.mod .
COPY go.sum .

# Force the go compiler to use modules
ENV GO111MODULE=on

# Get dependencies - will be cached if we don't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -v -o challenge

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine
RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /go/src/github.com/keremk/challenge-bot/challenge /challenge
COPY --from=builder /go/src/github.com/keremk/challenge-bot/challenge-db-key.json /challenge-db-key.json
COPY --from=builder /go/src/github.com/keremk/challenge-bot/static/. /static/.

# Run the web service on container startup.
CMD ["/challenge"]