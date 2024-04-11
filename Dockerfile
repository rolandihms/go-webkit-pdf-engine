# Builder to copy wkhtmltopdf binaries for alpine for latest version 0.12.6
FROM surnet/alpine-wkhtmltopdf:3.16.2-0.12.6-full as wkhtmltopdf
# wkhtmltopdf install dependencies
RUN apk add --no-cache \
    libstdc++ \
    libx11 \
    libxrender \
    libxext \
    libssl1.1 \
    ca-certificates \
    fontconfig \
    freetype \
    ttf-droid \
    ttf-freefont \
    ttf-liberation \
    # more fonts
    ;

# Use Build stage for compiling the go app
FROM golang:1.22-alpine AS BuildStage

# Install wkhtmltopdf
ENV DIR=/usr/local/bin/

#Change directory so that our commands run inside this new directory
WORKDIR $DIR
#Copy wkhtmltopdf binaries
COPY --from=wkhtmltopdf /bin/wkhtmltopdf /bin/libwkhtmltox.so /bin/

RUN apk add build-base

#CHange back to main Dir
WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .
# Build the Go application
RUN CGO_ENABLED=1 go build -o pdf-service .

# Deploy the application
# Must use 3.18.5, as since 3.19 libssl is giving issues
FROM alpine:3.18.5

WORKDIR /app

COPY --from=BuildStage /bin/wkhtmltopdf /bin/libwkhtmltox.so /bin/
COPY --from=BuildStage /app/ .

# Set the timezone and install CA certificates
RUN apk --no-cache add ca-certificates \
    tzdata \
    libstdc++ \
    libx11 \
    libxrender \
    libxext \
    ca-certificates \
    fontconfig \
    freetype \
    ttf-droid \
    ttf-freefont \
    ttf-liberation  \
    libssl1.1

EXPOSE 3000
#USER noroot:noroot

#Start the container with the application
ENTRYPOINT ["/app/pdf-service"]
