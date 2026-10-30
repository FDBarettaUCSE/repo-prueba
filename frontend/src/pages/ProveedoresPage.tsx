import {
  useCallback,
  useEffect,
  useState,
  type FormEvent,
} from "react";
import { useAuth } from "../auth/AuthContext";
import {
  listarProveedores,
  crearProveedor,
  actualizarProveedor,
  eliminarProveedor,
} from "../api/proveedores";
import { ErrorAPI } from "../api/respuestas";
import type {
  ProveedorDTO,
  CrearProveedorDTO,
  ActualizarProveedorDTO,
} from "../api/types";

interface FormularioProveedor {
  nombre: string;
  cuit: string;
  email: string;
  telefono: string;
}

const FORMULARIO_VACIO: FormularioProveedor = {
  nombre: "",
  cuit: "",
  email: "",
  telefono: "",
};

export function ProveedoresPage() {
  const { token, usuario } = useAuth();

  const puedeEditar =
    usuario?.rol === "administrador" ||
    usuario?.rol === "gerente";

  const [proveedores, setProveedores] = useState<ProveedorDTO[]>([]);
  const [formulario, setFormulario] =
    useState<FormularioProveedor>(FORMULARIO_VACIO);
  const [proveedorEditando, setProveedorEditando] =
    useState<string | null>(null);

  const [cargando, setCargando] = useState(true);
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [mensaje, setMensaje] = useState<string | null>(null);

  const cargarProveedores = useCallback(async () => {
    if (!token) return;

    setCargando(true);
    setError(null);

    try {
      const datos = await listarProveedores(token);
      setProveedores(datos);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar los proveedores.",
      );
    } finally {
      setCargando(false);
    }
  }, [token]);

  useEffect(() => {
    void cargarProveedores();
  }, [cargarProveedores]);

  function cancelarEdicion() {
    setFormulario(FORMULARIO_VACIO);
    setProveedorEditando(null);
    setError(null);
  }

  function editarProveedor(proveedor: ProveedorDTO) {
    setFormulario({
      nombre: proveedor.nombre,
      cuit: proveedor.cuit,
      email: proveedor.email,
      telefono: proveedor.telefono,
    });

    setProveedorEditando(proveedor.id);
    setError(null);
    setMensaje(null);
  }

  async function manejarEnvio(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (!token || !puedeEditar) return;

    setEnviando(true);
    setError(null);
    setMensaje(null);

    const datos = {
      nombre: formulario.nombre.trim(),
      cuit: formulario.cuit.trim(),
      email: formulario.email.trim(),
      telefono: formulario.telefono.trim(),
    };

    try {
      if (proveedorEditando) {
        const dto: ActualizarProveedorDTO = datos;

        await actualizarProveedor(proveedorEditando, dto, token);
        setMensaje("Proveedor actualizado correctamente.");
      } else {
        const dto: CrearProveedorDTO = datos;

        await crearProveedor(dto, token);
        setMensaje("Proveedor creado correctamente.");
      }

      cancelarEdicion();
      await cargarProveedores();
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo guardar el proveedor.",
      );
    } finally {
      setEnviando(false);
    }
  }

  async function manejarEliminar(proveedor: ProveedorDTO) {
    if (!token || !puedeEditar) return;

    if (
      !window.confirm(
        `¿Seguro que querés eliminar el proveedor "${proveedor.nombre}"?`,
      )
    ) {
      return;
    }

    setError(null);
    setMensaje(null);

    try {
      await eliminarProveedor(proveedor.id, token);

      if (proveedorEditando === proveedor.id) {
        cancelarEdicion();
      }

      setMensaje("Proveedor eliminado correctamente.");
      await cargarProveedores();
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo eliminar el proveedor.",
      );
    }
  }

  return (
    <div className="panel">
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <h1 className="mb-0">Proveedores</h1>
        <span className="badge text-bg-light border">
          {proveedores.length}{" "}
          {proveedores.length === 1 ? "proveedor" : "proveedores"}
        </span>
      </div>

      {mensaje && <p className="mensaje">{mensaje}</p>}
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}

      {puedeEditar && (
        <section className="filtros-barra mb-4">
          <h2>
            {proveedorEditando ? "Editar proveedor" : "Nuevo proveedor"}
          </h2>

          <form onSubmit={manejarEnvio}>
            <div className="row g-3">
              <div className="col-12 col-md-6 col-lg-3">
                <label className="form-label mb-0">
                  Nombre / Razón social
                  <input
                    className="form-control"
                    placeholder="Ej.: Insumos del Litoral S.R.L."
                    value={formulario.nombre}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        nombre: e.target.value,
                      })
                    }
                    required
                  />
                </label>
              </div>

              <div className="col-12 col-md-6 col-lg-3">
                <label className="form-label mb-0">
                  CUIT
                  <input
                    className="form-control"
                    placeholder="30-71234567-9"
                    value={formulario.cuit}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        cuit: e.target.value,
                      })
                    }
                    required
                  />
                </label>
              </div>

              <div className="col-12 col-md-6 col-lg-3">
                <label className="form-label mb-0">
                  Email
                  <input
                    type="email"
                    className="form-control"
                    placeholder="ventas@proveedor.com"
                    value={formulario.email}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        email: e.target.value,
                      })
                    }
                    required
                  />
                </label>
              </div>

              <div className="col-12 col-md-6 col-lg-3">
                <label className="form-label mb-0">
                  Teléfono
                  <input
                    className="form-control"
                    placeholder="+54 3492 400000"
                    value={formulario.telefono}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        telefono: e.target.value,
                      })
                    }
                    required
                  />
                </label>
              </div>
            </div>

            <div className="acciones mt-1">
              <button
                type="submit"
                className="btn btn-primary"
                disabled={enviando}
              >
                {enviando
                  ? "Guardando..."
                  : proveedorEditando
                    ? "Guardar cambios"
                    : "Crear proveedor"}
              </button>

              {proveedorEditando && (
                <button
                  type="button"
                  className="btn btn-outline-secondary"
                  onClick={cancelarEdicion}
                  disabled={enviando}
                >
                  Cancelar
                </button>
              )}
            </div>
          </form>
        </section>
      )}

      <section className="mt-3 pt-0 border-0">
        <h2>Listado de proveedores</h2>

        {cargando && (
          <div className="d-flex align-items-center gap-2 text-muted my-3">
            <span className="spinner-border spinner-border-sm" aria-hidden="true" />
            <p className="mb-0">Cargando proveedores...</p>
          </div>
        )}

        {!cargando && (
          <div className="tabla-wrapper table-responsive">
            <table className="table table-hover align-middle mb-0">
              <thead>
                <tr>
                  <th>Nombre</th>
                  <th>CUIT</th>
                  <th>Email</th>
                  <th>Teléfono</th>
                  {puedeEditar && <th className="text-end">Acciones</th>}
                </tr>
              </thead>

              <tbody>
                {proveedores.length === 0 ? (
                  <tr>
                    <td
                      colSpan={puedeEditar ? 5 : 4}
                      className="text-center text-muted py-4"
                    >
                      No hay proveedores para mostrar.
                    </td>
                  </tr>
                ) : (
                  proveedores.map((proveedor) => (
                    <tr key={proveedor.id}>
                      <td className="fw-medium">{proveedor.nombre}</td>
                      <td className="font-monospace small">{proveedor.cuit}</td>
                      <td>{proveedor.email}</td>
                      <td>{proveedor.telefono}</td>

                      {puedeEditar && (
                        <td>
                          <div className="acciones justify-content-end">
                            <button
                              type="button"
                              className="btn btn-outline-primary btn-sm"
                              onClick={() => editarProveedor(proveedor)}
                            >
                              Editar
                            </button>

                            <button
                              type="button"
                              className="btn btn-outline-danger btn-sm"
                              onClick={() =>
                                void manejarEliminar(proveedor)
                              }
                            >
                              Eliminar
                            </button>
                          </div>
                        </td>
                      )}
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  );
}
