package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// claveEnvJWTSecret es el nombre de la variable de entorno que tiene que
// traer el secreto con el que se firman los JWT. A propósito NO hay un
// valor hardcodeado de respaldo: un secreto conocido de antemano (o
// compartido entre todos los que clonan el repo) le permite a cualquiera
// forjar un token de administrador válido. Ver RequerirSecretoConfigurado,
// que main.go llama al arrancar para cortar ahí mismo si falta.
const claveEnvJWTSecret = "JWT_SECRET"

// RequerirSecretoConfigurado valida que JWT_SECRET esté seteado. Se llama
// UNA vez, al principio de main() — antes de levantar el servidor — para
// que una configuración faltante corte el arranque con un mensaje claro, en
// vez de fallar recién en el primer login (o, peor, si alguien reintrodujera
// un valor por defecto, firmar tokens con un secreto adivinable).
func RequerirSecretoConfigurado() error {
	if os.Getenv(claveEnvJWTSecret) == "" {
		return errors.New("falta la variable de entorno JWT_SECRET: la API no puede firmar JWT sin un secreto propio (ver .env.example)")
	}
	return nil
}

func getSecretKey() []byte {
	return []byte(os.Getenv(claveEnvJWTSecret))
}

// duracionToken es cuánto tiempo queda vigente un token desde que se emite.
// modificacion de duracion de tokens, pasaron de 2 horas a 30 minutos.
const duracionToken = 30 * time.Minute

// GenerarToken emite un JWT cuyo único claim de negocio es "sub" (subject):
// el ID (hex de bson.ObjectID) del usuario autenticado. A propósito NO viaja
// nada más sensible en el payload — el JWT solo está firmado, no encriptado,
// así que cualquiera que lo intercepte puede leerlo en base64 (ver Clase 3 —
// "Qué es un JWT").
func GenerarToken(usuarioID, rol, depositoID string) (string, error) {
	ahora := time.Now()
	claims := jwt.MapClaims{
		"sub":         usuarioID,
		"rol":         rol,
		"deposito_id": depositoID,
		"iat":         ahora.Unix(),
		"exp":         ahora.Add(duracionToken).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecretKey())
}

// ValidarToken decodifica el token, verifica la firma contra secretKey y
// chequea la expiración (todo eso lo hace jwt.ParseWithClaims por debajo). Si
// el token fue alterado, está mal formado, o está vencido, devuelve error.
func ValidarToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inesperado")
		}
		return getSecretKey(), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}
