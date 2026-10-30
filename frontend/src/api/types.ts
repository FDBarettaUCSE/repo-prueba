// Types 1: Usuarios

export interface RegistrarDTO {
  correo: string;
  contrasenia: string;
  rol: string;
  deposito_asociado?: string;
}

export interface CambiarContraseniaDTO {
  contrasenia: string;
  contrasenia_nueva: string;
}

export interface CambiarRolDTO {
  rol: string;
  deposito_asociado?: string;
}

export interface IngresarDTO {
  correo: string;
  contrasenia: string;
}

export interface EventoAuditoriaUsuarioDTO {
  accion: string;
  realizado_por: string;
  fecha_hora: string;
  detalle?: string;
}

export interface UsuarioDTO {
  id: string;
  correo: string;
  rol: string;
  activado: boolean;
  deposito_asociado?: string;
  historial_auditoria?: EventoAuditoriaUsuarioDTO[];
}

export interface UsuariosPaginadosDTO {
  items: UsuarioDTO[];
  total: number;
  pagina: number;
  tamanioPagina: number;
}

export interface CambiarEstadoUsuarioDTO {
  estado: boolean;
}

export interface RespuestaIngresar {
  token: string;
}

export interface ErrorResponseDTO {
  error: string;
}

// Esto no es un type, son solo los valores.
export const Rol = {
  Administrador: "administrador",
  Gerente: "gerente",
  Operario: "operario",
  Auditor: "auditor",
} as const;

// La idea es que el rol sea un type para que vaya como parametro si es necesario.
export type Rol = (typeof Rol)[keyof typeof Rol];


// Types 2: Depositos.

export interface DepositoDTO {
  id: string;
  nombre: string;
  localidad: string;
  provincia: string;
  responsable_id?: string;
}

export interface CrearDepositoDTO {
  nombre: string;
  localidad: string;
  provincia: string;
  responsable_id?: string;
}

export interface ActualizarDepositoDTO {
  nombre: string;
  localidad: string;
  provincia: string;
  responsable_id?: string;
}

// Types 3: Productos

export const UnidadMedida = {
  Unidad: "unidad",
  Kilogramo: "kilogramo",
  Litro: "litro",
  Metro: "metro",
} as const;

export type UnidadMedida =
  (typeof UnidadMedida)[keyof typeof UnidadMedida];

export interface ProductoDTO {
  id: string;
  nombre: string;
  descripcion: string;
  categoriaId: string;
  categoriaNombre: string;
  unidadMedida: UnidadMedida;
  costoUnitario: number;
  stockMinimo: number;
  fechaVencimiento?: string;
}

export interface CrearProductoDTO {
  nombre: string;
  descripcion: string;
  categoriaId: string;
  unidadMedida: UnidadMedida;
  costoUnitario: number;
  stockMinimo: number;
  fechaVencimiento?: string;
}

export interface ActualizarProductoDTO {
  nombre: string;
  descripcion: string;
  categoriaId: string;
  unidadMedida: UnidadMedida;
  costoUnitario: number;
  fechaVencimiento?: string;
}

export interface StockMinimoDTO {
  stockMinimo: number;
}

export interface FiltroProductosDTO {
  categoria_id?: string;
  pagina?: number;
  tamanioPagina?: number;
  todos?: string;
}

export interface ProductosPaginadosDTO {
  items: ProductoDTO[];
  total: number;
  pagina: number;
  tamanioPagina: number;
}

// Stocks de depositos (depositoProducto)

// Stock de un producto dentro de un depósito
export interface StockDepositoDTO {
  depositoId: string;
  depositoNombre: string;
  productoId: string;
  productoNombre: string;
  categoriaId: string;
  categoriaNombre: string;
  unidadMedida: string;
  stock: number;
  stockMinimo: number;
  stockMinimoOverride?: number;
  costoUnitario?: number;
  valorizado?: number;
  bajoMinimo: boolean;
}

// Configuración del stock mínimo particular de un depósito
export interface ConfigurarStockMinimoDTO {
  stockMinimo: number;
}

// Filtros para consultar el stock de un depósito
export interface ListarStockFiltro {
  categoriaId?: string;
  bajoMinimo?: boolean;
  pagina?: number;
  tamanioPagina?: number;
}

// Respuesta paginada del stock de un depósito
export interface StockDepositoPaginadoDTO {
  items: StockDepositoDTO[];
  total: number;
  pagina: number;
  tamanioPagina: number;
}

// Proveedores

export interface ProveedorDTO {
  id: string;
  nombre: string;
  cuit: string;
  email: string;
  telefono: string;
}

export interface CrearProveedorDTO {
  nombre: string;
  cuit: string;
  email: string;
  telefono: string;
}

export interface ActualizarProveedorDTO {
  nombre: string;
  cuit: string;
  email: string;
  telefono: string;
}

// Productos que suministra un proveedor
export interface AgregarProductoSuministradoDTO {
  productoId: string;
  costo: number;
  tiempoEntregaDias: number;
}

export interface ActualizarProductoSuministradoDTO {
  costo: number;
  tiempoEntregaDias: number;
}

export interface ProductoSuministradoDTO {
  productoId: string;
  productoNombre: string;
  costo: number;
  tiempoEntregaDias: number;
}

// Proveedores que suministran un producto
export interface ProveedorDeProductoDTO {
  proveedorId: string;
  proveedorNombre: string;
  costo: number;
  tiempoEntregaDias: number;
}

// Movimientos de stock:


export type TipoMovimiento = "ingreso" | "egreso" | "ajuste" | "transferencia";

// Datos necesarios para registrar un movimiento de stock
export interface RegistrarMovimientoDTO {
  producto_id: string;
  deposito_id: string;
  cantidad: number;
  tipo: TipoMovimiento;
  motivo: string;
}

// Representa un movimiento de stock registrado
export interface MovimientoStockDTO {
  id: string;
  producto_id: string;
  producto_nombre: string;
  deposito_id: string;
  cantidad: number;
  tipo: TipoMovimiento;
  usuario_responsable_id: string;
  fecha_hora: string;
  motivo: string;
}


// Parámetros opcionales para filtrar el historial
export interface FiltroMovimientosDTO {
  deposito_id?: string;
  producto_id?: string;
  tipo?: TipoMovimiento;
  fecha_desde?: string;
  fecha_hasta?: string;
  pagina?: number;
  tamanioPagina?: number;
}

// Respuesta paginada del historial de movimientos
export interface MovimientosPaginadosDTO {
  items: MovimientoStockDTO[];
  total: number;
  pagina: number;
  tamanioPagina: number;
}

// Transferencias entre depósitos


export type EstadoTransferencia =
  | "solicitada"
  | "aprobada"
  | "rechazada"
  | "completada";

// Datos necesarios para solicitar una transferencia
export interface SolicitarTransferenciaDTO {
  deposito_origen_id: string;
  deposito_destino_id: string;
  producto_id: string;
  cantidad: number;
  motivo: string;
}

// Datos para resolver una transferencia
export interface ResolverTransferenciaDTO {
  accion: "aprobar" | "rechazar";
  motivo?: string;
}

// Datos opcionales al completar una transferencia
export interface CompletarTransferenciaDTO {
  observaciones?: string;
}

// Representa una transferencia
export interface TransferenciaDTO {
  id: string;
  deposito_origen_id: string;
  deposito_destino_id: string;
  producto_id: string;
  cantidad: number;
  estado: EstadoTransferencia;
  usuario_solicitante_id: string;
  usuario_resolucion_id?: string;
  fecha_solicitud: string;
  fecha_resolucion?: string;
  motivo?: string;
}

// Parámetros para filtrar transferencias
export interface FiltroTransferenciasDTO {
  deposito_id?: string;
  estado?: EstadoTransferencia;
  pagina?: number;
  tamanioPagina?: number;
}

// Respuesta paginada del listado de transferencias
export interface TransferenciasPaginadasDTO {
  items: TransferenciaDTO[];
  total: number;
  pagina: number;
  tamanioPagina: number;
}


// Alertas de stock

export interface AlertaStockDTO {
  id: string;
  tipo: TipoAlerta;
  producto_id: string;
  producto_nombre: string;
  deposito_id: string;
  deposito_nombre: string;
  stock_minimo: number;
  stock_actual: number;
  fecha_vencimiento?: string;
  resuelta: boolean;
  fecha_creacion: string;
  fecha_resolucion?: string;
  modificado_por?: string;
}

export interface AlertasPaginadasDTO {
  items: AlertaStockDTO[];
  total: number;
  pendientes: number;
  pagina: number;
  tamanioPagina: number;
}

export interface FiltrosAlertasStock {
  deposito_id?: string;
  solo_pendientes?: boolean;
  pagina?: number;
  tamanioPagina?: number;
}

export interface ParametrosEscaneoAlertasStock {
  deposito_id?: string;
  dias_vencimiento?: number;
}

export interface RespuestaEscaneoAlertasStockDTO {
  mensaje: string;
  nuevas_alertas: AlertaStockDTO[];
  cantidad: number;
}

// Novedades depósito

export interface RespuestaNovedadDTO {
  texto: string;
  usuario_id: string;
  fecha_respuesta: string;
}

export interface NovedadDTO {
  id: string;
  deposito_id: string;
  texto: string;
  fecha_envio: string;
  leido: boolean;
  leido_por?: string[];
  respuesta?: RespuestaNovedadDTO;
  respuestas?: RespuestaNovedadDTO[];
}

export interface NovedadesPaginadoDTO {
  items: NovedadDTO[];
  total: number;
  no_leidas: number;
  pagina: number;
  tamanioPagina: number;
}

export interface FiltrosNovedades {
  solo_no_leidas?: boolean;
  pagina?: number;
  tamanioPagina?: number;
}

export interface EnviarNovedadDTO {
  texto: string;
}

export interface EnviarNovedadDifusionDTO {
  texto: string;
  depositos_ids?: string[];
  todos?: boolean;
}

export interface ResponderNovedadDTO {
  texto: string;
}

// Reportes

// ---------- Paginación ----------

export interface PaginacionDTO {
  pagina: number;
  tamanioPagina: number;
  total: number;
}

// ---------- Inventario ----------

export interface FiltroInventario {
  deposito_id?: string;
  categoria_id?: string;
  pagina?: number;
}

export interface ValorizacionDepositoDTO {
  deposito_id: string;
  deposito_nombre: string;
  stock: number;
  valorizado: number;
}

export interface ValorizacionCategoriaDTO {
  categoria_id: string;
  categoria_nombre: string;
  stock: number;
  valorizado: number;
}

export interface ReporteInventarioDTO {
  por_deposito: ValorizacionDepositoDTO[];
  paginacion_deposito: PaginacionDTO;
  por_categoria: ValorizacionCategoriaDTO[];
  paginacion_categoria: PaginacionDTO;
  stock_total: number;
  valorizado_total: number;
}

// ---------- Movimientos ----------

export type TipoMovimientoReporte =
  | "ingreso"
  | "egreso"
  | "ajuste"
  | "transferencia";

export interface FiltroReporteMovimientos {
  deposito_id?: string;
  tipo?: TipoMovimientoReporte;
  fecha_desde?: string;
  fecha_hasta?: string;
  pagina?: number;
}

export interface ResumenMovimientoDTO {
  deposito_id: string;
  deposito_nombre: string;
  fecha: string;
  tipo: string;
  cantidad_eventos: number;
  unidades_netas: number;
}

export interface ReporteMovimientosDTO {
  resumen: ResumenMovimientoDTO[];
  total_movimientos: number;
  paginacion: PaginacionDTO;
}

// Alertas ----------

export interface FiltroAlertas {
  deposito_id?: string;
  dias_vencimiento?: number;
  pagina?: number;
}

export interface ProductoBajoMinimoDTO {
  deposito_id: string;
  deposito_nombre: string;
  producto_id: string;
  producto_nombre: string;
  stock: number;
  stock_minimo: number;
}

export interface ProductoPorVencerDTO {
  deposito_id: string;
  deposito_nombre: string;
  producto_id: string;
  producto_nombre: string;
  stock: number;
  fecha_vencimiento: string;
  dias_para_vencer: number;
}

export interface ReporteAlertasDTO {
  bajo_minimo: ProductoBajoMinimoDTO[];
  paginacion_minimo: PaginacionDTO;
  por_vencer: ProductoPorVencerDTO[];
  paginacion_vencimiento: PaginacionDTO;
}

// ---------- Pendientes ----------

export interface FiltroPendientes {
  deposito_id?: string;
  pagina?: number;
}

export interface EstadoOrdenesDepositoDTO {
  deposito_id: string;
  deposito_nombre: string;
  borrador: number;
  confirmada: number;
  parcial: number;
  completada: number;
}

export interface EstadoTransferenciasDepositoDTO {
  deposito_id: string;
  deposito_nombre: string;
  solicitadas_como_origen: number;
  pendientes_de_aprobar: number;
  aprobadas_por_recibir: number;
}

export interface ReportePendientesDTO {
  ordenes: EstadoOrdenesDepositoDTO[];
  paginacion_ordenes: PaginacionDTO;
  transferencias: EstadoTransferenciasDepositoDTO[];
  paginacion_transferencias: PaginacionDTO;
}


// Órdenes de compra

export type EstadoOrdenCompra =
  | "borrador"
  | "confirmada"
  | "parcial"
  | "completada"
  | "descartada";

export interface OrdenCompraItemDTO {
  id: string;
  ordenCompraId: string;
  productoId: string;
  cantidadSolicitada: number;
  cantidadRecibida: number;
  costoUnitario: number;
}

export interface CrearOrdenCompraItemDTO {
  productoId: string;
  cantidadSolicitada: number;
}

export interface ActualizarOrdenCompraItemDTO {
  productoId: string;
  cantidadSolicitada: number;
}

export interface OrdenCompraDTO {
  id: string;
  proveedorId: string;
  depositoId: string;
  estado: EstadoOrdenCompra;
  items: OrdenCompraItemDTO[];
}

export interface CrearOrdenCompraDTO {
  proveedorId: string;
  depositoId: string;
}

export interface ActualizarOrdenCompraDTO {
  proveedorId: string;
  depositoId: string;
}

export interface FiltroOrdenesCompraDTO {
  deposito_id?: string;
  estado?: EstadoOrdenCompra;
  proveedor_id?: string;
  pagina?: number;
  tamanioPagina?: number;
}

export interface OrdenesCompraPaginadasDTO {
  items: OrdenCompraDTO[];
  total: number;
  pagina: number;
  tamanioPagina: number;
}

export interface RecepcionItemDTO {
  itemId: string;
  cantidad: number;
}

export interface RegistrarRecepcionDTO {
  items: RecepcionItemDTO[];
}

export interface SugerenciasOrdenCompraDTO {
  mensaje: string;
  ordenes_generadas: OrdenCompraDTO[];
  cantidad: number;
}


 // Categorías

export interface CategoriaDTO {
  id: string;
  nombre: string;
}

export interface CrearCategoriaDTO {
  nombre: string;
}

export interface ActualizarCategoriaDTO {
  nombre: string;
}

// Alertas pero de reportes 

// Alertas ----------

export interface FiltroAlertas {
  deposito_id?: string;
  dias_vencimiento?: number;
  pagina?: number;
}

export interface ProductoBajoMinimoDTO {
  deposito_id: string;
  deposito_nombre: string;
  producto_id: string;
  producto_nombre: string;
  stock: number;
  stock_minimo: number;
}

export interface ProductoPorVencerDTO {
  deposito_id: string;
  deposito_nombre: string;
  producto_id: string;
  producto_nombre: string;
  stock: number;
  fecha_vencimiento: string;
  dias_para_vencer: number;
}

export interface ReporteAlertasDTO {
  bajo_minimo: ProductoBajoMinimoDTO[];
  paginacion_minimo: PaginacionDTO;
  por_vencer: ProductoPorVencerDTO[];
  paginacion_vencimiento: PaginacionDTO;
}


// ---------- Alertas de stock ----------

export type TipoAlerta = "stock_bajo" | "vencimiento";

export interface AlertaStockDTO {
  id: string;
  tipo: TipoAlerta;
  producto_id: string;
  producto_nombre: string;
  deposito_id: string;
  deposito_nombre: string;
  stock_minimo: number;
  stock_actual: number;
  fecha_vencimiento?: string;
  resuelta: boolean;
  fecha_creacion: string;
  fecha_resolucion?: string;
  modificado_por?: string;
}

export interface AlertasPaginadasDTO {
  items: AlertaStockDTO[];
  total: number;
  pendientes: number;
  pagina: number;
  tamanioPagina: number;
}

export interface FiltrosAlertasStock {
  deposito_id?: string;
  solo_pendientes?: boolean;
  pagina?: number;
  tamanioPagina?: number;
}

