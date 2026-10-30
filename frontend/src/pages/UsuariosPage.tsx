import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useAuth } from "../auth/AuthContext";
import {
  registrar,
  listarUsuarios,
  cambiarRolUsuario,
  cambiarContrasenia,
  activarUsuario,
  desactivarUsuario,
} from "../api/usuarios";
import { listarDepositos } from "../api/depositos";
import { ErrorAPI } from "../api/respuestas";
import type { DepositoDTO, UsuarioDTO, Rol } from "../api/types";
import { Rol as Roles } from "../api/types";

// Datos necesarios para registrar un usuario.
interface FormularioRegistro {
  correo: string;
  contrasenia: string;
  rol: Rol;
  deposito_asociado: string;
}

const REGISTRO_VACIO: FormularioRegistro = {
  correo: "",
  contrasenia: "",
  rol: Roles.Operario,
  deposito_asociado: "",
};

// Datos necesarios para cambiar el rol.
interface FormularioRol {
  rol: Rol;
  deposito_asociado: string;
}

export function UsuariosPage() {
  const { token, usuario: usuarioActual } = useAuth();

  const [usuarios, setUsuarios] = useState<UsuarioDTO[]>([]);
  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);

  const [cargando, setCargando] = useState(true);
  const [errorLista, setErrorLista] = useState<string | null>(null);

  const [registro, setRegistro] =
    useState<FormularioRegistro>(REGISTRO_VACIO);
  const [enviandoRegistro, setEnviandoRegistro] = useState(false);
  const [errorRegistro, setErrorRegistro] = useState<string | null>(null);

  const [usuarioRol, setUsuarioRol] = useState<UsuarioDTO | null>(null);
  const [formularioRol, setFormularioRol] = useState<FormularioRol>({
    rol: Roles.Operario,
    deposito_asociado: "",
  });
  const [enviandoRol, setEnviandoRol] = useState(false);
  const [errorRol, setErrorRol] = useState<string | null>(null);

  const [usuarioContrasenia, setUsuarioContrasenia] =
    useState<UsuarioDTO | null>(null);
  const [contraseniaActual, setContraseniaActual] = useState("");
  const [contraseniaNueva, setContraseniaNueva] = useState("");
  const [enviandoContrasenia, setEnviandoContrasenia] = useState(false);
  const [errorContrasenia, setErrorContrasenia] = useState<string | null>(
    null,
  );

  const [errorAccion, setErrorAccion] = useState<string | null>(null);

  const esAdministrador = usuarioActual?.rol === Roles.Administrador;

  // Carga los usuarios y los depósitos que se pueden asociar a cada uno.
  const cargarDatos = useCallback(async () => {
    if (!token) {
      setCargando(false);
      return;
    }

    setCargando(true);
    setErrorLista(null);

    try {
      const [datosUsuarios, datosDepositos] = await Promise.all([
        listarUsuarios(token),
        listarDepositos(token),
      ]);

      setUsuarios(datosUsuarios);
      setDepositos(datosDepositos);
    } catch (e) {
      setErrorLista(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar los usuarios",
      );
    } finally {
      setCargando(false);
    }
  }, [token]);

  useEffect(() => {
    cargarDatos();
  }, [cargarDatos]);

  // Determina qué roles necesitan tener un depósito asociado.
  function requiereDeposito(rol: Rol) {
    return rol === Roles.Gerente || rol === Roles.Operario;
  }

  // Limpia los formularios de registro y de edición.
  function limpiarRegistro() {
    setRegistro(REGISTRO_VACIO);
    setErrorRegistro(null);
  }

  function cancelarCambioRol() {
    setUsuarioRol(null);
    setErrorRol(null);
  }

  function cancelarCambioContrasenia() {
    setUsuarioContrasenia(null);
    setContraseniaActual("");
    setContraseniaNueva("");
    setErrorContrasenia(null);
  }

  // Registra un usuario y actualiza la tabla.
  async function manejarRegistro(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (!token) return;

    setErrorRegistro(null);
    setEnviandoRegistro(true);

    try {
      const dto = {
        correo: registro.correo.trim(),
        contrasenia: registro.contrasenia,
        rol: registro.rol,
        ...(requiereDeposito(registro.rol)
          ? { deposito_asociado: registro.deposito_asociado }
          : {}),
      };

      await registrar(dto, token);

      limpiarRegistro();
      await cargarDatos();
    } catch (e) {
      setErrorRegistro(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo registrar el usuario",
      );
    } finally {
      setEnviandoRegistro(false);
    }
  }

  // Abre el formulario de cambio de rol con los datos actuales.
  function empezarCambioRol(usuario: UsuarioDTO) {
    setUsuarioRol(usuario);
    setFormularioRol({
      rol: usuario.rol as Rol,
      deposito_asociado: usuario.deposito_asociado ?? "",
    });
    setErrorRol(null);
    setErrorAccion(null);
    cancelarCambioContrasenia();
  }

  // Guarda el nuevo rol y el depósito cuando corresponda.
  async function manejarCambioRol(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (!token || !usuarioRol) return;

    setErrorRol(null);
    setEnviandoRol(true);

    try {
      const dto = {
        rol: formularioRol.rol,
        ...(requiereDeposito(formularioRol.rol)
          ? { deposito_asociado: formularioRol.deposito_asociado }
          : {}),
      };

      await cambiarRolUsuario(usuarioRol.id, dto, token);

      cancelarCambioRol();
      await cargarDatos();
    } catch (e) {
      setErrorRol(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo cambiar el rol",
      );
    } finally {
      setEnviandoRol(false);
    }
  }

  // Abre el formulario de cambio de contraseña para el usuario seleccionado.
  function empezarCambioContrasenia(usuario: UsuarioDTO) {
    setUsuarioContrasenia(usuario);
    setContraseniaActual("");
    setContraseniaNueva("");
    setErrorContrasenia(null);
    setErrorAccion(null);
    cancelarCambioRol();
  }

  // El backend exige la contraseña actual de la cuenta seleccionada.
  async function manejarCambioContrasenia(
    e: FormEvent<HTMLFormElement>,
  ) {
    e.preventDefault();

    if (!token || !usuarioContrasenia) return;

    setErrorContrasenia(null);
    setEnviandoContrasenia(true);

    try {
      await cambiarContrasenia(
        usuarioContrasenia.id,
        {
          contrasenia: contraseniaActual,
          contrasenia_nueva: contraseniaNueva,
        },
        token,
      );

      cancelarCambioContrasenia();
      setErrorAccion("La contraseña se cambió correctamente.");
    } catch (e) {
      setErrorContrasenia(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo cambiar la contraseña",
      );
    } finally {
      setEnviandoContrasenia(false);
    }
  }

  // Activa o desactiva una cuenta según su estado actual.
  async function manejarEstado(usuario: UsuarioDTO) {
    if (!token) return;

    const accion = usuario.activado ? "desactivar" : "activar";

    if (
      !window.confirm(
        `¿Querés ${accion} la cuenta de ${usuario.correo}?`,
      )
    ) {
      return;
    }

    setErrorAccion(null);
    setErrorLista(null);

    try {
      if (usuario.activado) {
        await desactivarUsuario(usuario.id, token);
      } else {
        await activarUsuario(usuario.id, token);
      }

      await cargarDatos();
    } catch (e) {
      setErrorAccion(
        e instanceof ErrorAPI
          ? e.message
          : `No se pudo ${accion} el usuario`,
      );
    }
  }

  if (!esAdministrador) {
    return (
      <div className="panel">
        <h1>Usuarios</h1>
        <p className="error mb-0">No tenés permisos para administrar usuarios.</p>
      </div>
    );
  }

  return (
    <div className="panel">
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <h1 className="mb-0">Usuarios</h1>
        <span className="badge text-bg-light border">
          {usuarios.length} {usuarios.length === 1 ? "usuario" : "usuarios"}
        </span>
      </div>

      <section className="filtros-barra mb-4">
        <h2 className="mb-2">Registrar usuario</h2>

        <form onSubmit={manejarRegistro}>
          <div className="row g-3">
            <div className="col-12 col-md-6 col-lg-3">
              <label htmlFor="correo" className="form-label">
                Correo electrónico
              </label>
              <input
                id="correo"
                type="email"
                className="form-control"
                placeholder="correo@deltainsumos.com"
                value={registro.correo}
                onChange={(e) =>
                  setRegistro({ ...registro, correo: e.target.value })
                }
                required
              />
            </div>

            <div className="col-12 col-md-6 col-lg-3">
              <label htmlFor="contrasenia" className="form-label">
                Contraseña
              </label>
              <input
                id="contrasenia"
                type="password"
                className="form-control"
                placeholder="••••••••"
                value={registro.contrasenia}
                onChange={(e) =>
                  setRegistro({ ...registro, contrasenia: e.target.value })
                }
                required
              />
            </div>

            <div className="col-12 col-md-6 col-lg-3">
              <label htmlFor="rolRegistro" className="form-label">
                Rol
              </label>
              <select
                id="rolRegistro"
                className="form-select"
                value={registro.rol}
                onChange={(e) =>
                  setRegistro({
                    ...registro,
                    rol: e.target.value as Rol,
                    deposito_asociado: "",
                  })
                }
              >
                <option value={Roles.Administrador}>Administrador</option>
                <option value={Roles.Gerente}>Gerente</option>
                <option value={Roles.Operario}>Operario</option>
                <option value={Roles.Auditor}>Auditor</option>
              </select>
            </div>

            {requiereDeposito(registro.rol) && (
              <div className="col-12 col-md-6 col-lg-3">
                <label htmlFor="depositoRegistro" className="form-label">
                  Depósito asociado
                </label>
                <select
                  id="depositoRegistro"
                  className="form-select"
                  value={registro.deposito_asociado}
                  onChange={(e) =>
                    setRegistro({
                      ...registro,
                      deposito_asociado: e.target.value,
                    })
                  }
                  required
                >
                  <option value="">Seleccionar depósito</option>
                  {depositos.map((deposito) => (
                    <option key={deposito.id} value={deposito.id}>
                      {deposito.nombre} — {deposito.localidad}
                    </option>
                  ))}
                </select>
              </div>
            )}
          </div>

          {errorRegistro && (
            <p className="error mb-0" role="alert">
              {errorRegistro}
            </p>
          )}

          <div className="acciones mt-1">
            <button
              type="submit"
              className="btn btn-primary"
              disabled={enviandoRegistro}
            >
              {enviandoRegistro ? "Registrando..." : "Registrar usuario"}
            </button>
          </div>
        </form>
      </section>

      {usuarioRol && (
        <section className="filtros-barra mb-4 border-primary">
          <h2 className="mb-2">Cambiar rol: {usuarioRol.correo}</h2>

          <form onSubmit={manejarCambioRol}>
            <div className="row g-3">
              <div className="col-12 col-md-6">
                <label htmlFor="rolNuevo" className="form-label">
                  Nuevo rol
                </label>
                <select
                  id="rolNuevo"
                  className="form-select"
                  value={formularioRol.rol}
                  onChange={(e) =>
                    setFormularioRol({
                      ...formularioRol,
                      rol: e.target.value as Rol,
                      deposito_asociado: "",
                    })
                  }
                >
                  <option value={Roles.Administrador}>Administrador</option>
                  <option value={Roles.Gerente}>Gerente</option>
                  <option value={Roles.Operario}>Operario</option>
                  <option value={Roles.Auditor}>Auditor</option>
                </select>
              </div>

              {requiereDeposito(formularioRol.rol) && (
                <div className="col-12 col-md-6">
                  <label htmlFor="depositoRol" className="form-label">
                    Depósito asociado
                  </label>
                  <select
                    id="depositoRol"
                    className="form-select"
                    value={formularioRol.deposito_asociado}
                    onChange={(e) =>
                      setFormularioRol({
                        ...formularioRol,
                        deposito_asociado: e.target.value,
                      })
                    }
                    required
                  >
                    <option value="">Seleccionar depósito</option>
                    {depositos.map((deposito) => (
                      <option key={deposito.id} value={deposito.id}>
                        {deposito.nombre} — {deposito.localidad}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>

            {errorRol && (
              <p className="error mb-0" role="alert">
                {errorRol}
              </p>
            )}

            <div className="acciones mt-1">
              <button
                type="submit"
                className="btn btn-primary"
                disabled={enviandoRol}
              >
                {enviandoRol ? "Guardando..." : "Guardar rol"}
              </button>
              <button
                type="button"
                className="btn btn-outline-secondary"
                onClick={cancelarCambioRol}
                disabled={enviandoRol}
              >
                Cancelar
              </button>
            </div>
          </form>
        </section>
      )}

      {usuarioContrasenia && (
        <section className="filtros-barra mb-4 border-primary">
          <h2 className="mb-2">
            Cambiar contraseña: {usuarioContrasenia.correo}
          </h2>

          <form onSubmit={manejarCambioContrasenia}>
            <div className="row g-3">
              <div className="col-12 col-md-6">
                <label htmlFor="contraseniaActual" className="form-label">
                  Contraseña actual
                </label>
                <input
                  id="contraseniaActual"
                  type="password"
                  className="form-control"
                  value={contraseniaActual}
                  onChange={(e) => setContraseniaActual(e.target.value)}
                  required
                />
              </div>

              <div className="col-12 col-md-6">
                <label htmlFor="contraseniaNueva" className="form-label">
                  Nueva contraseña
                </label>
                <input
                  id="contraseniaNueva"
                  type="password"
                  className="form-control"
                  value={contraseniaNueva}
                  onChange={(e) => setContraseniaNueva(e.target.value)}
                  required
                />
              </div>
            </div>

            {errorContrasenia && (
              <p className="error mb-0" role="alert">
                {errorContrasenia}
              </p>
            )}

            <div className="acciones mt-1">
              <button
                type="submit"
                className="btn btn-primary"
                disabled={enviandoContrasenia}
              >
                {enviandoContrasenia
                  ? "Cambiando..."
                  : "Cambiar contraseña"}
              </button>
              <button
                type="button"
                className="btn btn-outline-secondary"
                onClick={cancelarCambioContrasenia}
                disabled={enviandoContrasenia}
              >
                Cancelar
              </button>
            </div>
          </form>
        </section>
      )}

      {errorAccion && (
        <p className="error" role="alert">
          {errorAccion}
        </p>
      )}

      <section className="mt-3 pt-0 border-0">
        <h2>Usuarios registrados</h2>

        {cargando && (
          <div className="d-flex align-items-center gap-2 text-muted my-3">
            <span className="spinner-border spinner-border-sm" aria-hidden="true" />
            <p className="mb-0">Cargando usuarios...</p>
          </div>
        )}

        {errorLista && (
          <p className="error" role="alert">
            {errorLista}
          </p>
        )}

        {!errorLista && !cargando && usuarios.length === 0 && (
          <p className="text-muted mb-0">No hay usuarios registrados.</p>
        )}

        {!errorLista && usuarios.length > 0 && (
          <div
            className="tabla-wrapper table-responsive"
            style={{ opacity: cargando ? 0.6 : 1 }}
          >
            <table className="table table-hover align-middle mb-0">
              <thead>
                <tr>
                  <th>Correo</th>
                  <th>Rol</th>
                  <th>Depósito asociado</th>
                  <th>Estado</th>
                  <th className="text-end">Acciones</th>
                </tr>
              </thead>

              <tbody>
                {usuarios.map((usuario) => {
                  const deposito = depositos.find(
                    (d) => d.id === usuario.deposito_asociado,
                  );

                  return (
                    <tr key={usuario.id}>
                      <td className="fw-medium">{usuario.correo}</td>
                      <td>
                        <span className="badge text-bg-light border text-capitalize">
                          {usuario.rol}
                        </span>
                      </td>
                      <td>{deposito?.nombre ?? "—"}</td>
                      <td>
                        <span
                          className={`badge rounded-pill ${
                            usuario.activado
                              ? "text-bg-success"
                              : "text-bg-secondary"
                          }`}
                        >
                          {usuario.activado ? "Activo" : "Inactivo"}
                        </span>
                      </td>
                      <td>
                        <div className="acciones justify-content-end">
                          <button
                            type="button"
                            className="btn btn-outline-primary btn-sm"
                            onClick={() => empezarCambioRol(usuario)}
                          >
                            Cambiar rol
                          </button>

                          <button
                            type="button"
                            className="btn btn-outline-secondary btn-sm"
                            onClick={() =>
                              empezarCambioContrasenia(usuario)
                            }
                          >
                            Cambiar contraseña
                          </button>

                          <button
                            type="button"
                            className={`btn btn-sm ${
                              usuario.activado
                                ? "btn-outline-danger"
                                : "btn-outline-success"
                            }`}
                            onClick={() => manejarEstado(usuario)}
                          >
                            {usuario.activado ? "Desactivar" : "Activar"}
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  );
}
