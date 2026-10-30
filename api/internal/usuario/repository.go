package usuario

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UsuarioRepositorio interface {
	EncontarPorCorreo(ctx context.Context, correo string) (Usuario, error)
	ExisteCorreo(ctx context.Context, correo string) (bool, error)
	EncontrarPorId(ctx context.Context, id string) (Usuario, error)
	EncontrarTodos(ctx context.Context) ([]Usuario, error)
	Listar(ctx context.Context, pagina, tamanioPagina int) ([]Usuario, int64, error)
	CrearUsuario(ctx context.Context, u Usuario) (Usuario, error)
	ActualizarContrasenaHash(ctx context.Context, id, idAdmin string, nuevoHash string) error
	ActualizarRolUsuario(ctx context.Context, id, idAdmin, idDeposito string, rol Rol) error
	DesactivarUsuario(ctx context.Context, id string, idAdmin string) error
	ActivarUsuario(ctx context.Context, id string, idAdmin string) error
	EncontrarRolPorId(ctx context.Context, id string) (string, error)
	ContarUsuarios(ctx context.Context, roles ...Rol) (int64, error)
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

// EncontarPorCorreo es la consulta que usa el login: exige que el usuario
// esté activo.
func (r *RepositorioMongo) EncontarPorCorreo(ctx context.Context, correo string) (Usuario, error) {
	var u Usuario
	if err := r.coleccion.FindOne(ctx, bson.M{"correo": correo, "activado": true}).Decode(&u); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Usuario{}, errors.New("usuario no encontrado")
		}
		return Usuario{}, err
	}
	return u, nil
}

// ExisteCorreo verifica unicidad del email sobre toda la colección (incluidos
// los usuarios desactivados) para evitar duplicados al registrar.
func (r *RepositorioMongo) ExisteCorreo(ctx context.Context, correo string) (bool, error) {
	n, err := r.coleccion.CountDocuments(ctx, bson.M{"correo": correo})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// EncontrarPorId es la consulta que usa el cambio de contraseña y verificaciones.
func (r *RepositorioMongo) EncontrarPorId(ctx context.Context, id string) (Usuario, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Usuario{}, ErrUsuarioInexistente
	}

	var u Usuario
	if err := r.coleccion.FindOne(ctx, bson.M{"_id": oid}).Decode(&u); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Usuario{}, ErrUsuarioInexistente
		}
		return Usuario{}, err
	}
	return u, nil
}

// CrearUsuario inserta un usuario nuevo. Quien llama (Service.Registrar) ya dejó
// PasswordHash con el resultado de bcrypt — el repository nunca hashea nada,
// solo persiste lo que recibe. Datos de auditoría también asignados en service.
func (r *RepositorioMongo) CrearUsuario(ctx context.Context, u Usuario) (Usuario, error) {

	result, err := r.coleccion.InsertOne(ctx, u)
	if err != nil {
		return Usuario{}, err
	}

	u.ID = result.InsertedID.(bson.ObjectID)
	return u, nil
}

// ActualizarContrasenaHash sobreescribe el hash de contraseña y agrega una
// entrada inmutable en historial_auditoria.
func (r *RepositorioMongo) ActualizarContrasenaHash(ctx context.Context, id, adminId string, nuevoHash string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	adminOID, err := bson.ObjectIDFromHex(adminId)
	if err != nil {
		return err
	}

	ahora := time.Now()
	evento := EventoAuditoriaUsuario{
		Accion:       "cambio_contrasenia",
		RealizadoPor: adminOID,
		Fecha:        ahora,
	}

	update := bson.M{
		"$set": bson.M{
			"contrasenia_hasheada":     nuevoHash,
			"auditoria.modificado_por": adminOID,
			"auditoria.actualizado_en": ahora,
		},
		"$push": bson.M{
			"historial_auditoria": evento,
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

// ActualizarRolUsuario modifica el rol (y depósito asociado) del usuario y
// registra el cambio en historial_auditoria.
func (r *RepositorioMongo) ActualizarRolUsuario(ctx context.Context, id, idAdmin, idDeposito string, rol Rol) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return err
	}

	ahora := time.Now()
	detalle := "rol=" + string(rol)

	// Campos obligatorios a actualizar
	setFields := bson.M{
		"rol":                      rol,
		"auditoria.modificado_por": oidAdmin,
		"auditoria.actualizado_en": ahora,
	}

	update := bson.M{"$set": setFields}

	// Manejo condicional del depósito ($set o $unset)
	if idDeposito != "" {
		oidDeposito, err := bson.ObjectIDFromHex(idDeposito)
		if err != nil {
			return err
		}
		setFields["deposito_asociado"] = oidDeposito
		detalle += ", deposito=" + oidDeposito.Hex()
	} else {
		// Si no tiene depósito (ej. pasa a ser Admin/Auditor), lo elimina del documento
		update["$unset"] = bson.M{
			"deposito_asociado": "",
		}
	}

	update["$push"] = bson.M{
		"historial_auditoria": EventoAuditoriaUsuario{
			Accion:       "cambio_rol",
			Detalle:      detalle,
			RealizadoPor: oidAdmin,
			Fecha:        ahora,
		},
	}

	// Ejecutar la actualización en MongoDB
	result, err := r.coleccion.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("usuario no encontrado")
	}

	return nil
}

// DesactivarUsuario desactiva un usuario y deja constancia en historial_auditoria.
func (r *RepositorioMongo) DesactivarUsuario(
	ctx context.Context,
	id string,
	idAdmin string,
) error {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return err
	}

	ahora := time.Now()

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid, "activado": true},
		bson.M{
			"$set": bson.M{
				"activado":                 false,
				"auditoria.modificado_por": oidAdmin,
				"auditoria.actualizado_en": ahora,
			},
			"$push": bson.M{
				"historial_auditoria": EventoAuditoriaUsuario{
					Accion:       "desactivacion_usuario",
					RealizadoPor: oidAdmin,
					Fecha:        ahora,
				},
			},
		},
	)

	if err != nil {
		return err
	}

	if resultado.MatchedCount == 0 {
		return errors.New("usuario no encontrado o ya desactivado")
	}

	return nil
}

func (r *RepositorioMongo) EncontrarRolPorId(
	ctx context.Context,
	id string,
) (string, error) {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return "", err
	}

	var usuario Usuario

	err = r.coleccion.FindOne(
		ctx,
		bson.M{
			"_id":      oid,
			"activado": true,
		},
	).Decode(&usuario)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", errors.New("usuario no encontrado")
		}

		return "", err
	}

	return string(usuario.Rol), nil
}

func (r *RepositorioMongo) EncontrarTodos(ctx context.Context) ([]Usuario, error) {
	cur, err := r.coleccion.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var usuarios []Usuario
	if err := cur.All(ctx, &usuarios); err != nil {
		return nil, err
	}
	return usuarios, nil
}

// Listar devuelve los usuarios paginados con skip/limit en MongoDB.
func (r *RepositorioMongo) Listar(ctx context.Context, pagina, tamanioPagina int) ([]Usuario, int64, error) {
	if pagina < 1 {
		pagina = 1
	}
	if tamanioPagina < 1 {
		tamanioPagina = 20
	}
	if tamanioPagina > MaxTamanioPagina {
		tamanioPagina = MaxTamanioPagina
	}

	total, err := r.coleccion.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "auditoria.creado_en", Value: -1}, {Key: "_id", Value: 1}}).
		SetSkip(int64((pagina - 1) * tamanioPagina)).
		SetLimit(int64(tamanioPagina))

	cur, err := r.coleccion.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var usuarios []Usuario
	if err := cur.All(ctx, &usuarios); err != nil {
		return nil, 0, err
	}
	return usuarios, total, nil
}

// ActivarUsuario reactiva un usuario y deja constancia en historial_auditoria.
func (r *RepositorioMongo) ActivarUsuario(ctx context.Context, id, idAdmin string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return err
	}

	ahora := time.Now()
	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid, "activado": false},
		bson.M{
			"$set": bson.M{
				"activado":                 true,
				"auditoria.modificado_por": oidAdmin,
				"auditoria.actualizado_en": ahora,
			},
			"$push": bson.M{
				"historial_auditoria": EventoAuditoriaUsuario{
					Accion:       "activacion_usuario",
					RealizadoPor: oidAdmin,
					Fecha:        ahora,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	if resultado.MatchedCount == 0 {
		return errors.New("usuario no encontrado o ya activado")
	}

	return nil
}

func (r *RepositorioMongo) ContarUsuarios(
	ctx context.Context,
	roles ...Rol,
) (int64, error) {

	filtro := bson.M{
		"activado": true,
	}

	if len(roles) > 0 {
		filtro["rol"] = roles[0]
	}

	return r.coleccion.CountDocuments(ctx, filtro)
}
