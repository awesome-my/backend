FROM golang:1.22-alpine as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download \
    && go mod verify
COPY . .
RUN CGO_ENABLED=0 go build -v -ldflags="-s -w" -o /app/awesome-my cmd/awesome-my/*.go

FROM gcr.io/distroless/static-debian11:latest
COPY --from=build /app/awesome-my /


ENTRYPOINT ["/awesome-my", "serve"]