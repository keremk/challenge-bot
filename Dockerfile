# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.12 as builder

# Copy local code to the container image.
WORKDIR /go/src/github.com/keremk/challenge-bot
COPY . .

# Build the command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -v -o challenge

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