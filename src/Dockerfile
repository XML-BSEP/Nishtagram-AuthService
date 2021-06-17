# FROM golang:alpine

# # Install git
# RUN apk update && apk add --no-cache git

# # Where our files will be in the docker container 

# # Copy the source from the current directory to the working Directory inside the container 
# # Source also contains go.mod and go.sum which are dependency files
# WORKDIR /src

# # Get Dependency
# # COPY go.* .
# # RUN go mod download


# COPY . .

# RUN go build -o main .

# # Install Air for hot reload
# RUN go get -u github.com/cosmtrek/air

# # The ENTRYPOINT defines the command that will be ran when the container starts up
# # In this case air command for hot reload go apps on file changes
# ENTRYPOINT air

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

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/configurations ./src/configurations
COPY --from=builder /app/certificate ./src/certificate
COPY --from=builder /app/main .
# Command to run the executable
CMD ["./main"]