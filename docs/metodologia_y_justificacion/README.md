# FUNDAMENTACIÓN METODOLÓGICA Y JUSTIFICACIÓN DEL PROYECTO DE GRADO
## Especificación y Diseño del Formato de Conocimiento Abierto Orientado a Datos (OKF/ODKF v0.2)
### Carrera de Ingeniería de Sistemas — Universidad de Aquino Bolivia (UDABOL)

**Título del Proyecto de Grado:**  
# "Diseño e Implementación del Formato de Conocimiento Abierto Orientado a Datos"
*(Subtítulo / Mención Técnica: Un Contenedor Semántico Agnóstico como Índice de Estándares, Red P2P y Resolución Descentralizada para el Intercambio Estático y Dinámico de Datos con Contexto)*

**Postulante:** Andres Humberto Chirinos Lizondo  
**Institución:** Universidad de Aquino Bolivia (UDABOL)  
**Fecha:** Octubre 2026

---

## 1. Naturaleza de este Repositorio: Fuente Única de Verdad Normativa

Este repositorio ([`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)) es la **sede conceptual, metodológica y normativa** del proyecto de grado. Aquí reside:
1. **La Especificación del Contenedor Universal:** El *Open Knowledge Bundle*, que desacopla la información de cualquier tecnología, proveedor de nube o framework específico.
2. **El Meta-Índice de Estándares Preexistentes:** La unificación armónica de Frictionless Data Table Schema, W3C DCAT v2, W3C PROV-O, W3C SKOS, Apache Arrow, Parquet y Common Workflow Language (CWL).
3. **El Modelo Dual de Oferta de Datos:**
   - *Estático:* Archivos tabulares inmutables referenciados por URIs con validación de hashes SHA-256.
   - *Dinámico:* Servicios, consultas analíticas y cómputos atestados (*Attested Computations*) con trazabilidad en W3C PROV-O.
4. **La Red P2P y el Modelo de "DNS Soberano":**
   - Resolución distribuida sin ICANN ni servidores raíz centralizados mediante la sintaxis canónica:  
     `dataset.recurso@catalogo` (con utilidad de diagnóstico y consulta `okf-dig`).
   - Registros soberanos tipados (`SOA`, `CAT`, `RES`, `SCH`, `LIN`, `SRV`).
   - Protocolo de transporte agnóstico (QUIC, TCP+TLS, WebSockets, WebRTC) y difusión por **GossipSub** sobre Kademlia DHT.
5. **La Fundamentación y Justificación de la Tesis:** Todo el marco epistemológico, el análisis de problemas, la justificación de ingeniería y la estructura de capítulos para la UDABOL.

---

## 2. Jerarquía de Dependencia: `open-knowledge-format` gobierna a `datamesh-sdk`

```
  +-------------------------------------------------------------------------+
  |              ESPECIFICACIÓN NORMATIVA Y ESTÁNDAR AGNÓSTICO              |
  |             /home/andreschirinos/Proyectos/open-knowledge-format        |
  |  (Aquí reside la tesis, justificación, contenedor, red P2P y DNS)      |
  +-------------------------------------------------------------------------+
  | * SPEC.md (§1 a §15): Especificación universal del contenedor y P2P     |
  | * Protocolo P2P: Resolución dataset.recurso@catalogo y registros DNS    |
  | * schemas/: Esquemas JSON formales (okf-concept.schema.json)            |
  | * bundles/: Ejemplos agnósticos de bundles con contexto semántico       |
  | * docs/metodologia_y_justificacion/: Documentación formal UDABOL        |
  +-------------------------------------------------------------------------+
                                     |
                                     | [Gobierna mediante contratos normativos]
                                     v
  +-------------------------------------------------------------------------+
  |               SDK DE INFRAESTRUCTURA Y MOTOR DE EJECUCIÓN               |
  |                 /home/andreschirinos/Proyectos/datamesh-sdk             |
  |     (Implementación técnica en software de lo acordado arriba)          |
  +-------------------------------------------------------------------------+
  | * Arquitectura Hexagonal (Ports & Adapters) en Python                   |
  | * Motor analítico embebido DuckDB y Apache Arrow (cero-copia)           |
  | * Adaptador de red y resolvedor P2P / okf-dig implementado en código    |
  | * Verificación criptográfica de confianza (SHA-256 / llaves SSH)        |
  | * Fachada de 1 línea (`import datamesh as dm; df = dm.read(...)`)        |
  | * Interfaz de línea de comandos (CLI `dm init`, `dm query`, `dm dig`)    |
  +-------------------------------------------------------------------------+
```

---

## 3. Catálogo de Documentos Académicos en esta Carpeta

| Documento | Enfoque y Contenido Principal |
| :--- | :--- |
| **[`01_ESTRUCTURA_CAPITULAR_TESIS_UDABOL.md`](01_ESTRUCTURA_CAPITULAR_TESIS_UDABOL.md)** | **Capitulación Oficial UDABOL:** Los 6 capítulos formales de la tesis estructurados para defender el diseño del contenedor abierto como índice universal, la red P2P/DNS y su implementación técnica en el SDK. |
| **[`02_JUSTIFICACION_Y_DISENO_DEL_CONTENEDOR.md`](02_JUSTIFICACION_Y_DISENO_DEL_CONTENEDOR.md)** | **Justificación Integral:** Diagnóstico de silos, justificación técnica del contenedor como índice de estándares, soporte dual (estático y dinámico), justificación económica, social para Bolivia y rigor DSRM. |
| **[`03_ESPECIFICACION_DEL_CONTENEDOR_Y_CLAVES_ESPECIALES.md`](03_ESPECIFICACION_DEL_CONTENEDOR_Y_CLAVES_ESPECIALES.md)** | **Especificación del Contenedor:** Catálogo exhaustivo de claves especiales en el frontmatter (`types`, `trust`, `provenance`, `computation`, `lifecycle`, `contracts`, `semantic_anchors`), direccionamiento canónico por tríadas e invariantes PE1-PE5 / OE1-OE5. |
| **[`04_MARCO_METODOLOGICO_DSRM_Y_BIBLIOGRAFIA.md`](04_MARCO_METODOLOGICO_DSRM_Y_BIBLIOGRAFIA.md)** | **DSRM y Bibliografía:** Fundamentación según Peffers et al. y March & Smith (distribución de los 4 artefactos) y bibliografía completa en normas APA 7ma / IEEE. |
| **[`05_ARQUITECTURA_P2P_Y_RESOLUCION_DISTRIBUIDA.md`](05_ARQUITECTURA_P2P_Y_RESOLUCION_DISTRIBUIDA.md)** | **Red P2P y Resolución Soberana ("DNS de Datos"):** Principios de redes distribuidas (Kademlia, Content-Addressing), sintaxis `dataset.recurso@catalogo`, comando `okf-dig`, registros tipados (`SOA`, `CAT`, `RES`, `SCH`, `LIN`, `SRV`), transporte multiprotocolo (QUIC/TCP/WebRTC), travesía de NAT y difusión por GossipSub. |

---

## 4. Claves de Defensa ante el Tribunal Calificador (UDABOL)

1. **¿Por qué un nuevo contenedor y no simplemente usar CSV o Parquet directamente?**  
   *Respuesta:* Los archivos CSV o Parquet son contenedores de microdatos ciegos: no expresan la procedencia, el linaje, la identidad criptográfica del autor ni el significado ontológico de las columnas. El formato OKF envuelve estos formatos tabulares como un meta-índice, asegurando que los datos viajen siempre con su contexto explicativo y contratos sintácticos garantizados.
2. **¿Cómo soporta el contenedor datos tanto estáticos como dinámicos?**  
   *Respuesta:* Para datos estáticos, utiliza distribuciones físicas validadas con huellas SHA-256 e invariantes de Table Schema. Para datos dinámicos, introduce el concepto de **Attested Computations** y servicios, especificando la consulta SQL o el script generador, los parámetros de entrada y el linaje formal en W3C PROV-O, permitiendo computar vistas actualizadas bajo demanda sin perder trazabilidad.
3. **¿Cómo resuelve la red los nombres sin un servidor DNS centralizado?**  
   *Respuesta:* Implementa un **DNS Soberano P2P** basado en tablas de hash distribuidas (Kademlia DHT). Al consultar `dataset.recurso@catalogo` (mediante `okf-dig`), el sistema ubica el PeerID del nodo soberano firmante, verifica su registro de autoridad `SOA` con claves Ed25519 y resuelve los registros `RES` (ubicación física), `SCH` (esquema) y `LIN` (linaje), logrando tolerancia total a fallos y resistencia a la censura.
4. **¿Cuál es la relación entre este repositorio y `datamesh-sdk`?**  
   *Respuesta:* Este repositorio (`open-knowledge-format`) es la **especificación normativa universal** (independiente de cualquier lenguaje o tecnología). `datamesh-sdk` es una **instanciación técnica de referencia** que implementa la Arquitectura Hexagonal en Python para consumir y producir estos bundles de forma eficiente y soberana.
