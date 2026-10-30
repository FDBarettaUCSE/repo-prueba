
import type { ReactNode } from "react";
import { NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";

export function Layout({
  children,
}: {
  children: ReactNode;
}) {
  const { usuario, logout } = useAuth();
  const navigate = useNavigate();

  const rol = usuario?.rol;

  const esAdministrador = rol === "administrador";
  const esGerente = rol === "gerente";
  const esOperario = rol === "operario";
  const esAuditor = rol === "auditor";

  const puedeGestionarCatalogos = esAdministrador || esGerente;
  const puedeVerDepositos = esAdministrador || esAuditor;
  const puedeVerStock = esAdministrador || esGerente || esOperario;
  const puedeVerMovimientos = esAdministrador || esGerente || esOperario || esAuditor;
  const puedeVerTransferencias = esAdministrador || esGerente || esOperario;
  const puedeVerOrdenesCompra = esAdministrador || esGerente || esAuditor;
  const puedeVerNovedades = esGerente || esOperario || esAuditor;
  const puedeVerAlertas = esAdministrador || esGerente || esOperario || esAuditor;
  const puedeVerReportes = esAdministrador || esGerente || esAuditor;

  function manejarLogout() {
    logout();
    navigate("/login");
  }

  return (
    <div className="layout">
      <header className="header shadow-sm">
        <div className="container-fluid d-flex flex-wrap align-items-center justify-content-between gap-3 px-2 px-md-3">
          <span className="header-marca d-flex align-items-center gap-2">
            <span className="marca-icono" aria-hidden="true">📦</span>
            <span>GeStock</span>
          </span>

          <nav className="header-nav">
            <NavLink
              to="/"
              end
              className={({ isActive }) =>
                isActive ? "nav-item-link activo" : "nav-item-link"
              }
            >
              Inicio
            </NavLink>

            {esAdministrador && (
              <NavLink
                to="/usuarios"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Usuarios
              </NavLink>
            )}

            {puedeVerDepositos && (
              <NavLink
                to="/depositos"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Depósitos
              </NavLink>
            )}

            {esAdministrador && (
              <NavLink
                to="/categorias"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Categorías
              </NavLink>
            )}

            {puedeGestionarCatalogos && (
              <NavLink
                to="/productos"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Productos
              </NavLink>
            )}

            {puedeGestionarCatalogos && (
              <NavLink
                to="/proveedores"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Proveedores
              </NavLink>
            )}

            {puedeGestionarCatalogos && (
              <NavLink
                to="/productos-suministrados"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Suministros
              </NavLink>
            )}

            {puedeVerStock && (
              <NavLink
                to="/stock"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Stock
              </NavLink>
            )}

            {puedeVerMovimientos && (
              <NavLink
                to="/movimientos"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Movimientos
              </NavLink>
            )}

            {puedeVerTransferencias && (
              <NavLink
                to="/transferencias"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Transferencias
              </NavLink>
            )}

            {puedeVerOrdenesCompra && (
              <NavLink
                to="/ordenes-compra"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Órdenes de compra
              </NavLink>
            )}

            {puedeVerNovedades && (
              <NavLink
                to="/novedades"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Novedades
              </NavLink>
            )}

            {puedeVerAlertas && (
              <NavLink
                to="/alertas-stock"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Alertas
              </NavLink>
            )}

            {puedeVerReportes && (
              <NavLink
                to="/reportes"
                className={({ isActive }) =>
                  isActive ? "nav-item-link activo" : "nav-item-link"
                }
              >
                Reportes
              </NavLink>
            )}
          </nav>

          <div className="header-usuario ms-auto">
            {usuario && (
              <div className="d-flex align-items-center gap-2">
                <span className="badge rounded-pill header-rol-badge text-capitalize">
                  {usuario.rol}
                </span>
                <span className="header-email">{usuario.correo}</span>
              </div>
            )}

            <button
              type="button"
              className="btn btn-outline-light btn-sm"
              onClick={manejarLogout}
            >
              Cerrar sesión
            </button>
          </div>
        </div>
      </header>

      <main className="contenido container-xl">{children}</main>
    </div>
  );
}
