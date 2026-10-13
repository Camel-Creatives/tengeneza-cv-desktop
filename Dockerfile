FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/tengenezacv-api ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache chromium ttf-dejavu ca-certificates
COPY --from=build /out/tengenezacv-api /tengenezacv-api
ENV CHROMIUM_EXECUTABLE=/usr/bin/chromium-browser
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/tengenezacv-api"]
