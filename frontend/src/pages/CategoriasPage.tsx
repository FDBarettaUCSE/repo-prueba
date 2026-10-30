import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useAuth } from "../auth/AuthContext";
import {
  listarCategorias,
  crearCategoria,
  actualizarCategoria,
  eliminarCategoria,
} from "../api/categorias";
import { ErrorAPI } from "../api/respuestas";
import type { CategoriaDTO } from "../api/types";

// Estado del formulario: solamente necesitamos el nombre de la categoría.
interface FormularioCategoria {
  nombre: string;
}

const FORMULARIO_VACIO: FormularioCategoria = {
  nombre: "",
};

export function CategoriasPage() {
  const { token, usuario } = useAuth();

  // Estado de la lista de categorías.
  const [categorias, setCategorias] = useState<CategoriaDTO[]>([]);
  const [cargando, setCargando] = useState(true);
  const [errorLista, setErrorLista] = useState<string | null>(null);

  // Estado del formulario: crear una categoría o editar una existente.
  const [formulario, setFormulario] = useState<FormularioCategoria>(FORMULARIO_VACIO);
  const [editandoId, setEditandoId] = useState<string | null>(null);
  const [enviando, setEnviando] = useState(false);
  const [errorFormulario, setErrorFormulario] = useState<string | null>(null);

  const esAdministrador = usuario?.rol === "administrador";

  // Carga las categorías desde el backend. Se reutiliza después de cada
  // operación para actualizar la tabla sin recargar toda la página.
  const cargarCategorias = useCallback(async () => {
    if (!token) return;

    setCargando(true);
    setErrorLista(null);

    try {
      const datos = await listarCategorias(token);
      setCategorias(datos);
    } catch (e) {
      setErrorLista(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar las categorías",
      );
    } finally {
      setCargando(false);
    }
  }, [token]);

  // Carga inicial de la lista.
  useEffect(() => {
    cargarCategorias();
  }, [cargarCategorias]);

  // Carga los datos de una categoría en el formulario para editarla.
  function empezarEdicion(categoria: CategoriaDTO) {
    setFormulario({
      nombre: categoria.nombre,
    });
    setEditandoId(categoria.id);
    setErrorFormulario(null);
  }

  // Limpia el formulario y vuelve al modo de creación.
  function cancelarEdicion() {
    setFormulario(FORMULARIO_VACIO);
    setEditandoId(null);
    setErrorFormulario(null);
  }

  // Una única función para crear y actualizar categorías.
  async function manejarEnvio(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (!token) return;

    setErrorFormulario(null);
    setEnviando(true);

    try {
      const dto = {
        nombre: formulario.nombre.trim(),
      };

      if (editandoId) {
        await actualizarCategoria(editandoId, dto, token);
      } else {
        await crearCategoria(dto, token);
      }

      cancelarEdicion();
      await cargarCategorias();
    } catch (e) {
      setErrorFormulario(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo guardar la categoría",
      );
    } finally {
      setEnviando(false);
    }
  }

  // Pide confirmación antes de eliminar una categoría.
  async function manejarEliminar(id: string) {
    if (!token) return;

    if (!window.confirm("¿Eliminar esta categoría?")) return;

    setErrorLista(null);

    try {
      await eliminarCategoria(id, token);

      // Si se elimina la categoría que se estaba editando, se cancela
      // la edición para no conservar un formulario de un registro eliminado.
      if (editandoId === id) {
        cancelarEdicion();
      }

      await cargarCategorias();
    } catch (e) {
      setErrorLista(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo eliminar la categoría",
      );
    }
  }

  return (
    <div className="panel">
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <h1 className="mb-0">Categorías</h1>
        <span className="badge text-bg-light border">
          {categorias.length} {categorias.length === 1 ? "categoría" : "categorías"}
        </span>
      </div>

      {esAdministrador && (
        <form onSubmit={manejarEnvio} className="filtros-barra mb-4">
          <div className="row g-3 align-items-end">
            <div className="col-12 col-md-7">
              <label className="form-label mb-0">
                Nombre de la categoría
                <input
                  className="form-control"
                  placeholder="Ej.: Herramientas eléctricas, Seguridad industrial..."
                  value={formulario.nombre}
                  onChange={(e) =>
                    setFormulario({
                      ...formulario,
                      nombre: e.target.value,
                    })
                  }
                  maxLength={100}
                  required
                />
              </label>
            </div>

            <div className="col-12 col-md-5">
              <div className="acciones">
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={enviando}
                >
                  {enviando
                    ? "Guardando..."
                    : editandoId
                      ? "Guardar cambios"
                      : "Crear categoría"}
                </button>

                {editandoId && (
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
            </div>
          </div>

          {errorFormulario && (
            <p className="error mb-0" role="alert">
              {errorFormulario}
            </p>
          )}
        </form>
      )}

      {cargando && categorias.length === 0 && (
        <div className="d-flex align-items-center gap-2 text-muted my-3">
          <span className="spinner-border spinner-border-sm" aria-hidden="true" />
          <p className="mb-0">Cargando categorías...</p>
        </div>
      )}

      {errorLista && (
        <p className="error" role="alert">
          {errorLista}
        </p>
      )}

      {!errorLista && (categorias.length > 0 || !cargando) && (
        <div
          className="tabla-wrapper table-responsive"
          style={{ opacity: cargando ? 0.6 : 1 }}
        >
          <table className="table table-hover align-middle mb-0">
            <thead>
              <tr>
                <th>Nombre</th>
                {esAdministrador && <th className="text-end">Acciones</th>}
              </tr>
            </thead>

            <tbody>
              {categorias.map((categoria) => (
                <tr key={categoria.id}>
                  <td className="fw-medium">{categoria.nombre}</td>

                  {esAdministrador && (
                    <td>
                      <div className="acciones justify-content-end">
                        <button
                          type="button"
                          className="btn btn-outline-primary btn-sm"
                          onClick={() => empezarEdicion(categoria)}
                        >
                          Editar
                        </button>

                        <button
                          type="button"
                          className="btn btn-outline-danger btn-sm"
                          onClick={() => manejarEliminar(categoria.id)}
                        >
                          Eliminar
                        </button>
                      </div>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!cargando && !errorLista && categorias.length === 0 && (
        <p className="text-muted mt-3 mb-0">No hay categorías registradas.</p>
      )}
    </div>
  );
}
