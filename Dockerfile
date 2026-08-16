FROM golang:1.26-alpine AS builder

# install the C compiler tools required for Go's race detector
RUN apk add --no-cache build-base

# set the working directory inside the container
WORKDIR /app

# download Go modules (caches dependencies to speed up future builds)
COPY go.mod go.sum ./
RUN go mod download

# copy the rest of the source code
COPY . .

# run all tests
RUN CGO_ENABLED=1 go test -race -v ./...

# build the binary statically so it can run without C dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o packing-server ./cmd/server

# ==========================================

# create a tiny production image

FROM alpine:latest

# add root certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# copy only the compiled binary from the builder stage
COPY --from=builder /app/packing-server .
COPY --from=builder /app/openapi.yaml .
COPY --from=builder /app/swagger.html .

# expose the API port
EXPOSE 8080

# run the binary
CMD ["./packing-server"]