
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "./auth/AuthContext";
import { RutaProtegida } from "./components/RutaProtegida";
import { Layout } from "./components/Layout";
import { IniciarSesionPage } from "./pages/InicioSesionPage";
import { InicioPage } from "./pages/InicioPage";
import { CategoriasPage } from "./pages/CategoriasPage";
import { UsuariosPage } from "./pages/UsuariosPage";
import { DepositosPage } from "./pages/DepositosPage";
import { ProductosPage } from "./pages/ProductosPage";
import { ProveedoresPage } from "./pages/ProveedoresPage";
import { StockPage } from "./pages/StockPage";
import { ProductosSuministradosPage } from "./pages/ProductosSuministradosPage";
import { MovimientosPage } from "./pages/MovimientosPage";
import { TransferenciasPage } from "./pages/TransferenciasPage";
import { NovedadesPage } from "./pages/NovedadesPage";
import { AlertasStockPage } from "./pages/AlertasStockPage";
import { OrdenesCompraPage } from "./pages/OrdenesCompraPage";
import { ReportesPage } from "./pages/ReportesPage";

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Inicio de sesión: no requiere autenticación */}
          <Route
            path="/login"
            element={<IniciarSesionPage />}
          />

          {/* Inicio: requiere una sesión válida */}
          <Route
            path="/"
            element={
              <RutaProtegida>
                <Layout>
                  <InicioPage />
                </Layout>
              </RutaProtegida>
            }
          />

          {/* Categorías de productos */}

          <Route
            path="/categorias"
            element={
              <RutaProtegida>
                <Layout>
                  <CategoriasPage />
                </Layout>
              </RutaProtegida>
            }
          />

          {/* Modulo usuarios */}
          <Route
            path="/usuarios"
            element={
              <RutaProtegida>
                <Layout>
                  <UsuariosPage />
                </Layout>
              </RutaProtegida>
            }
          />

          {/* Modulo depositos */}
          <Route
            path="/depositos"
            element={
              <RutaProtegida>
                <Layout>
                  <DepositosPage />
                </Layout>
              </RutaProtegida>
            }
          />

          { /* Modulo productos */}
          <Route
            path="/productos"
            element={
              <RutaProtegida>
                <Layout>
                  <ProductosPage />
                </Layout>
              </RutaProtegida>
            }
          />

          { /* Modulo stock (depositoProducto) */}
          <Route
            path="/stock"
            element={
              <RutaProtegida>
                <Layout>
                  <StockPage />
                </Layout>
              </RutaProtegida>
            }
          />

          {/* Modulo proveedores */}
          <Route
            path="/proveedores"
            element={
              <RutaProtegida>
                <Layout>
                  <ProveedoresPage />
                </Layout>
              </RutaProtegida>
            }
          />

          {/* Productos suministrados por los proveedores */}
          <Route
            path="/productos-suministrados"
            element={
              <RutaProtegida>
                <Layout>
                  <ProductosSuministradosPage />
                </Layout>
              </RutaProtegida>
            }
          />

          {/* Movimientos de stock */}
          <Route 
            path="/movimientos" 
            element={
                <RutaProtegida>
                  <Layout>
                    <MovimientosPage />
                  </Layout>
              </RutaProtegida>
            } 
          />

          {/* Movimientos de stock */}
          <Route 
            path="/transferencias" 
            element={
                <RutaProtegida>
                  <Layout>
                    <TransferenciasPage />
                  </Layout>
              </RutaProtegida>
            } 
          />

          {/* Novedades deposito */}
          <Route
            path="/novedades"
            element={
              <RutaProtegida>
                <Layout>
                  <NovedadesPage />
                </Layout>
              </RutaProtegida>
            }
          />

          {/* Alertas de stock */}
          <Route
            path="/alertas-stock"
            element={
              <RutaProtegida>
                <Layout>
                  <AlertasStockPage />
                </Layout>
              </RutaProtegida>
            }
          />

          {/* Órdenes de compra */}
          <Route
            path="/ordenes-compra"
            element={
              <RutaProtegida>
                <Layout>
                  <OrdenesCompraPage />
                </Layout>
              </RutaProtegida>
            }
          />

          {/* Reportes */}
          <Route
            path="/reportes"
            element={
              <RutaProtegida>
                <Layout>
                  <ReportesPage />
                </Layout>
              </RutaProtegida>
            }
          />

          {/* Cualquier ruta desconocida vuelve al inicio */}
          <Route
            path="*"
            element={<Navigate to="/" replace />}
          />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
