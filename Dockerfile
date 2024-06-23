FROM golang:1.22.4 AS builder

WORKDIR /go/src/tweet-bypass-api

# Copy only the necessary files and folders excluding those listed in .dockerignore
COPY . .

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /tweet-bypass-api

# Stage 2: Create a minimal runtime image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy only the necessary files and the db folder from the builder stage
COPY --from=builder /tweet-bypass-api .

RUN apk add dumb-init
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

EXPOSE 3000

CMD ["./tweet-bypass-api"]