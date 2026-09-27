package usuario

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UsuarioRepositorio interface {
	EncontarPorCorreo(ctx context.Context, correo string) (Usuario, error)
	EncontrarPorId(ctx context.Context, id string) (Usuario, error)
	CrearUsuario(ctx context.Context, u Usuario) (Usuario, error)
	ActualizarContrasenaHash(ctx context.Context, id, idAdmin string, nuevoHash string) error
	ActualizarRolUsuario(ctx context.Context, id, idAdmin, idDeposito string, rol Rol) error
	ActivarUsuario(ctx context.Context, id, idAdmin string, activo bool) error
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: coleccion}
}

// var _ Repository = (*RepositorioMongo)(nil) fuerza en tiempo de compilación
// que RepositorioMongo cumple la interfaz Repository (ver Clase 1).
var _ UsuarioRepositorio = (*RepositorioMongo)(nil)

// FindByEmail es la consulta que usa el login: el email es el identificador
// que ingresa el usuario, no el _id interno de Mongo.
func (r *RepositorioMongo) EncontarPorCorreo(ctx context.Context, correo string) (Usuario, error) {
	var u Usuario
	if err := r.coleccion.FindOne(ctx, bson.M{"correo": correo}).Decode(&u); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Usuario{}, errors.New("usuario no encontrado")
		}
		return Usuario{}, err
	}
	return u, nil
}

// FindByID es la consulta que usa el cambio de contraseña: el ID viaja en el
// claim "sub" del JWT ya validado por AuthMiddleware, no en el body.
func (r *RepositorioMongo) EncontrarPorId(ctx context.Context, id string) (Usuario, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Usuario{}, err
	}

	var u Usuario
	if err := r.coleccion.FindOne(ctx, bson.M{"_id": oid, "activado": true}).Decode(&u); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Usuario{}, errors.New("usuario no encontrado o desactivado")
		}
		return Usuario{}, err
	}
	return u, nil
}

// Create inserta un usuario nuevo. Quien llama (Service.Registrar) ya dejó
// PasswordHash con el resultado de bcrypt — el repository nunca hashea nada,
// solo persiste lo que recibe. Datos de auditoría también asignados en service
func (r *RepositorioMongo) CrearUsuario(ctx context.Context, u Usuario) (Usuario, error) {

	result, err := r.coleccion.InsertOne(ctx, u)
	if err != nil {
		return Usuario{}, err
	}

	u.ID = result.InsertedID.(bson.ObjectID)
	return u, nil
}

// ActualizarPasswordHash sobreescribe SOLO el campo password_hash — nunca
// se toca el email ni el _id desde este método, para no correr el riesgo de
// pisar otro campo del documento por accidente.
func (r *RepositorioMongo) ActualizarContrasenaHash(ctx context.Context, id, adminId string, nuevoHash string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	adminOID, err := bson.ObjectIDFromHex(adminId)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"contraseña_hasheada":      nuevoHash,
			"auditoria.modificado_por": adminOID,
			"auditoria.actualizado_en": time.Now(),
		}}
	result, err := r.coleccion.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("usuario no encontrado")
	}
	return nil
}

// Modifica el rol del usuario, sólo posible para administradores cuyos datos quedan en auditoria
func (r *RepositorioMongo) ActualizarRolUsuario(ctx context.Context, id, idAdmin, idDeposito string, rol Rol) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return err
	}

	// Campos obligatorios a actualizar
	setFields := bson.M{
		"rol":                      rol,
		"auditoria.modificado_por": oidAdmin,
		"auditoria.actualizado_en": time.Now(),
	}

	update := bson.M{"$set": setFields}

	// Manejo condicional del depósito ($set o $unset)
	if idDeposito != "" {
		oidDeposito, err := bson.ObjectIDFromHex(idDeposito)
		if err != nil {
			return err
		}
		setFields["deposito_asociado"] = oidDeposito
	} else {
		// Si no tiene depósito (ej. pasa a ser Admin/Auditor), lo elimina del documento
		update["$unset"] = bson.M{
			"deposito_asociado": "",
		}
	}

	//Ejecutar la actualización en MongoDB
	result, err := r.coleccion.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("usuario no encontrado")
	}

	return nil
}

// Activa o desactiva un usuario, sólo posible para administradores cuyos datos quedan en auditoria
func (r *RepositorioMongo) ActivarUsuario(ctx context.Context, id, idAdmin string, activo bool) error {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"activado":                 activo,
			"auditoria.modificado_por": oidAdmin,
			"auditoria.actualizado_en": time.Now(), // Fecha y hora actual
		},
	}

	result, err := r.coleccion.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("usuario no encontrado")
	}
	return nil
}
