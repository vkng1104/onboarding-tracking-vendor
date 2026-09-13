FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /vendor-onboarding-api ./cmd/api

FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app
COPY --from=build /vendor-onboarding-api /usr/local/bin/vendor-onboarding-api
USER app
EXPOSE 8080
ENTRYPOINT ["vendor-onboarding-api"]
