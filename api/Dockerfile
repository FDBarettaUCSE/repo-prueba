# Build: compila el binario estático de la API. Usa golang:1.26 porque
# go.mod declara "go 1.26.2" — una imagen más vieja rechaza el build.
FROM golang:1.26 AS build
WORKDIR /src

# Copiamos primero solo los archivos de dependencias para que Docker cachee
# la descarga de módulos mientras no cambien go.mod/go.sum (ver Clase 2 —
# "capas de Docker").
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produce un binario estático: no depende de libc del sistema,
# así que puede correr sobre una imagen final sin ninguna herramienta de
# compilación (distroless).
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# Runtime: distroless no tiene shell, package manager ni nada más que el
# binario y sus certificados TLS — la superficie de ataque de la imagen
# final es mínima. La variante ":nonroot" corre como un usuario sin
# privilegios en vez de root.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
