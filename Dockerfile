# ---- build stage ----
FROM golang:1.25-alpine AS build
WORKDIR /app
# faster builds
ENV CGO_ENABLED=0 GO111MODULE=on

# download deps first
COPY go.mod go.sum ./
RUN go mod download

# copy sources
COPY . .

# build
RUN go build -o /out/issue-service ./cmd/issue-service

# ---- run stage ----
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /out/issue-service /app/issue-service
EXPOSE 8081
USER nonroot:nonroot
ENTRYPOINT ["/app/issue-service"]


