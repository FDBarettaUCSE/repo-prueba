# breaking-bad-requests
Baretta Franco David,
Bonino Agustín José,
Krause Guillermo Andrés.

## Cómo levantar el proyecto

1. Copiar `.env.example` a `.env` y completar `JWT_SECRET` con un valor
   propio (por ejemplo, `openssl rand -base64 48`). Sin esto la API no
   arranca a propósito — ver `api/internal/auth/jwt.go`.
   Completar también `INITIAL_ADMIN_EMAIL` e `INITIAL_ADMIN_PASSWORD` (entre
   8 y 64 caracteres): con la base vacía son obligatorias para crear el
   primer administrador y no tienen valor por defecto — ver
   `api/internal/usuario/service.go` (`InicializarAdmin`).
2. `docker compose up --build`

   Esto levanta MongoDB (con un volumen para persistir los datos) y la API
   en `http://localhost:8080`.

