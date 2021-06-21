FROM golang:latest as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
COPY Nishtagram-Logger/ ../../Nishtagram-Logger/
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy everything from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

######## Start a new stage from scratch #######
FROM alpine:latest  

WORKDIR /root/

COPY --from=builder /app/configurations ./src/configurations
COPY --from=builder /app/certificate ./src/certificate
COPY --from=builder /app/http ./src/http
COPY --from=builder /app/grpc ./src/grpc
COPY --from=builder /app/assets ./src/assets
COPY --from=builder /app/domain ./src/domain
COPY --from=builder /app/gateway ./src/gateway
COPY --from=builder /app/helper ./src/helper
COPY --from=builder /app/infrastructure ./src/infrastructure
COPY --from=builder /app/interactor ./src/interactor
COPY --from=builder /app/usecase ./src/usecase
COPY --from=builder /app/repository ./src/repository


COPY --from=builder /app/main .
# Command to run the executable
CMD ["./main"]