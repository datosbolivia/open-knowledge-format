# ESTRUCTURA CAPITULAR DE PROYECTO DE GRADO
## Carrera de Ingeniería de Sistemas — Universidad de Aquino Bolivia (UDABOL)

**Título del Proyecto de Grado:**  
# "Diseño e Implementación del Formato de Conocimiento Abierto Orientado a Datos"
*(Subtítulo / Mención Técnica: Un Contenedor Semántico Agnóstico como Índice de Estándares para el Intercambio Estático y Dinámico de Datos con Contexto)*

**Modalidad:** Proyecto de Grado  
**Grado Académico al que se Postula:** Licenciatura en Ingeniería de Sistemas  
**Área de Especialidad:** Arquitectura de Software Distribuido, Estandarización de Formatos y Web Semántica  
**Postulante:** Andres Humberto Chirinos Lizondo  
**Ecosistema del Proyecto:**
- **Repositorio Matriz (Especificación, Contenedor y Abstracción de Estándares):** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)
- **Repositorio Dependiente (Implementación del SDK y Motor Hexagonal):** [`datamesh-sdk`](file:///home/andreschirinos/Proyectos/datamesh-sdk)

---

### PRESENTACIÓN INSTITUCIONAL Y PRELIMINARES
- **Portada Oficial UDABOL:** Logotipo institucional, Facultad de Ciencia y Tecnología, Carrera de Ingeniería de Sistemas, Título del Proyecto de Grado, Autor, Tutor/Relator Metodológico, La Paz - Bolivia.
- **Páginas Preliminares:**
  - Acta de Aprobación de Perfil y Defensa de Grado.
  - Dedicatoria y Agradecimientos.
  - Resumen Ejecutivo (Español) y Abstract (Inglés) con Palabras Clave / Keywords.
  - Índice General Paginado.
  - Índice de Figuras, Diagramas de Arquitectura y Flujos de Contenedor.
  - Índice de Tablas, Cuadros Técnicos y Esquemas de Metadatos.
  - Glosario de Términos del Lenguaje Ubicuo (DDD) y Abreviaturas.

---

### CAPÍTULO I: MARCO INTRODUCTORIO Y DE REFERENCIA
*Este capítulo establece el contexto del problema, la necesidad del contenedor abierto como índice universal de estándares, formula el problema central y delimita los alcances.*

- **1.1 Introducción**
  - La evolución de la gestión analítica: desde los Data Lakes cerrados hacia el intercambio federado de conocimiento abierto.
  - El Formato de Conocimiento Abierto (OKF/ODKF v0.2) como contenedor semántico universal que indexa protocolos y estándares preexistentes.
- **1.2 Antecedentes**
  - *1.2.1 Antecedentes Internacionales:* Paradigma Data Mesh (Zhamak Dehghani), espacios europeos de datos (Gaia-X, IDSA), iniciativas de catálogos semánticos y abiertos (Our World in Data, Frictionless Data, Wikidata).
  - *1.2.2 Antecedentes Nacionales (Bolivia):* Diagnóstico de datos en el Estado boliviano (guías AGETIC) y en la academia (UMSA, UDABOL); dispersión en silos no interoperables, falta de metadatos de linaje y pérdida de contexto institucional.
- **1.3 Planteamiento del Problema**
  - *1.3.1 Diagnóstico de la Situación Actual:* Árbol de problemas (causas raíz: dependencia de formatos propietarios, dispersión de estándares sin un contenedor unificador, fragilidad en flujos dinámicos y estáticos).
  - *1.3.2 Formulación del Problema de Ingeniería:*  
    *¿De qué manera el diseño e implementación de un Formato de Conocimiento Abierto Orientado a Datos, estructurado como un contenedor agnóstico que indexe estándares preexistentes y preserve el contexto tanto en modalidades estáticas como dinámicas, permite resolver la fragmentación, acoplamiento y falta de soberanía en la gestión distribuida de datos?*
- **1.4 Objetivos de la Investigación**
  - *1.4.1 Objetivo General:*  
    Diseñar e implementar el Formato de Conocimiento Abierto Orientado a Datos (OKF/ODKF) como una especificación formal y un contenedor semántico agnóstico que indexe estándares preexistentes (W3C DCAT, PROV-O, SKOS, Frictionless, Arrow, CWL) para el intercambio de datos con contexto (estático y dinámico), sustentado por una implementación técnica de referencia mediante un SDK desacoplado basado en arquitectura hexagonal (`datamesh-sdk`).
  - *1.4.2 Objetivos Específicos:*
    1. Sistematizar el estado del arte y compilar los estándares internacionales de datos tabulares, catálogos, procedencia y flujos analíticos.
    2. Diseñar la arquitectura del contenedor universal OKF en `open-knowledge-format`, definiendo claves especiales (`types`, `trust`, `provenance`, `computation`, `lifecycle`) para el intercambio contextualizado de datos.
    3. Formalizar el soporte dual para ofertas de datos: estáticas (archivos inmutables Parquet/CSV con contratos y hashes) y dinámicas (servicios, vistas SQL y cómputos atestados).
    4. Implementar en `datamesh-sdk` el motor de software de referencia bajo arquitectura hexagonal, consumiendo los lineamientos del repositorio de estandarización.
    5. Evaluar y verificar rigurosamente la conformidad del formato mediante suites de validación sintáctica, pruebas herméticas y casos de uso con datasets reales de Bolivia.
- **1.5 Justificación de la Investigación**
  - *1.5.1 Justificación Técnica:* Neutralidad e interoperabilidad mediante un contenedor meta-índice; desacoplamiento de storage y computación.
  - *1.5.2 Justificación Económica:* Supresión de licencias cerradas y optimización de recursos mediante cómputo local vectorizado.
  - *1.5.3 Justificación Social y Académica:* Soberanía tecnológica boliviana, fomento de la ciencia abierta y reproducibilidad FAIR para universidades (UDABOL/UMSA).
  - *1.5.4 Justificación Metodológica:* Rigor en Design Science Research Methodology (DSRM) y documentación con ADRs.
- **1.6 Alcance y Delimitaciones**
  - *1.6.1 Alcance del Formato (open-knowledge-format):* Especificación normativa `SPEC.md`, esquemas JSON formales (`okf-concept.schema.json`), bundles de dominios agnósticos.
  - *1.6.2 Alcance del SDK (datamesh-sdk):* Implementación técnica en Python, adaptadores de E/S, CLI y motor analítico embebido.
  - *1.6.3 Alcance Geográfico y Temporal:* Aplicación en Bolivia, periodo 2025–2026.
  - *1.6.4 Límites del Proyecto:* No administra centros de datos ni impone infraestructura propietaria.

---

### CAPÍTULO II: MARCO TEÓRICO Y ESTADO DEL ARTE
*Este capítulo compila los fundamentos teóricos, los estándares que el contenedor indexa y la literatura científica de respaldo.*

- **2.1 El Paradigma Data Mesh y la Filosofía de Conocimiento Abierto**
  - Principios de Zhamak Dehghani: Dominio soberano, Datos como Producto, Plataforma de autoservicio y Gobernanza federada.
  - De los datos crudos al "Conocimiento con Contexto": vinculación de microdatos con explicaciones conceptuales legibles por humanos y máquinas.
- **2.2 Catálogo e Índice de Estándares Preexistentes Indexados por el Formato**
  - *Estándares de Esquema y Datos Tabulares:* Frictionless Data Package, Table Schema y Apache Arrow IPC / Parquet.
  - *Estándares de Catálogo y Distribución:* W3C DCAT v2 (`dcat:Dataset`, `dcat:Distribution`, `dcat:accessURL`).
  - *Estándares de Linaje y Procedencia:* W3C PROV-O (`prov:Entity`, `prov:Activity`, `prov:Agent`, `prov:wasGeneratedBy`).
  - *Estándares de Taxonomía y Semántica:* W3C SKOS (`skos:Concept`) y ontologías OWL (`owl:sameAs` hacia Wikidata/DBpedia).
  - *Estándares de Pipelines y Cómputo:* Common Workflow Language (CWL v1.2) y especificaciones OpenAPI.
- **2.3 Contratos de Datos (Data Contracts)**
  - Sanderson, Freeman y Schmidt (2025): El contrato como interfaz explícita entre productores y consumidores de datos.
- **2.4 Patrones Arquitectónicos de Aislamiento**
  - Arquitectura Hexagonal o Puertos y Adaptadores (Dr. Alistair Cockburn, Garrido de Paz).
  - Domain-Driven Design (DDD) Estratégico y Táctico (Eric Evans, Vaughn Vernon).
  - Principios SOLID y Clean Architecture (Robert C. Martin).
- **2.5 Tecnologías de Procesamiento Analítico Embebido y Cero-Copia**
  - DuckDB: Arquitectura de base de datos OLAP embebida (Raasveldt & Mühleisen).
  - Apache Arrow: Modelo de memoria columnar interoperable.
- **2.6 Teoría de Sistemas Distribuidos, Redes P2P y Nombres Descentralizados**
  - Redes direccionables por contenido (Content-Addressed Storage - CAS vs. Location-Addressed).
  - Tablas de Hash Distribuidas (DHT Kademlia): métrica XOR y tolerancia bizantina a fallos.
  - Sistemas de nombres descentralizados y soberanos frente a DNS centralizado (ICANN).
  - Protocolos de difusión epidémica (GossipSub) y técnicas de travesía de NAT (STUN, ICE, Relay v2).

---

### CAPÍTULO III: MARCO METODOLÓGICO Y DE INGENIERÍA
*Este capítulo fundamenta el método científico DSRM y la relación de dependencia entre repositorios.*

- **3.1 Metodología de Investigación: Design Science Research Methodology (DSRM)**
  - Justificación de DSRM para proyectos de grado en ingeniería (Peffers et al., 2008; Hevner et al., 2004).
  - Clasificación de los 4 Artefactos (March & Smith, 1995) distribuidos en los dos repositorios:
    - *Constructos y Modelos:* Especificación del contenedor y esquemas en `open-knowledge-format`.
    - *Métodos e Instanciaciones:* Algoritmos de validación, motor y SDK en `datamesh-sdk`.
  - Las 6 Fases de Peffers et al. aplicadas al ciclo de vida del formato y del motor.
- **3.2 Gobernanza Metodológica y Relación de Dependencia**
  - Principio rector: *El SDK depende estrictamente de los lineamientos del repositorio de estandarización*.
  - Protocolo de evolución de la especificación (versionado semántico de OKF).
  - Registro de Decisiones de Arquitectura (ADRs) para preservar la trazabilidad histórica de diseño.
- **3.3 Técnicas e Instrumentos de Ingeniería de Software**
  - Métricas de calidad de código y análisis estático (Ruff, Mypy, Bandit).
  - Pirámide de pruebas herméticas (Pytest, validadores en Go/Python).

---

### CAPÍTULO IV: DISEÑO DEL CONTENEDOR E INGENIERÍA DEL SDK
*Este capítulo constituye el corazón de la tesis: la especificación detallada del contenedor (open-knowledge-format) y su realización técnica (datamesh-sdk).*

- **4.1 Especificación Formal del Contenedor de Conocimiento Abierto (`open-knowledge-format`)**
  - Filosofía de diseño: *Zero-Microdata Principle* y neutralidad de agentes y vendors.
  - Estructura canónica del Bundle OKF:
    - Directorio raíz del paquete, directivas de nombres reservados y archivos de índice (`index.md`).
    - Documentos de Concepto (`concepts/*.md`): cuerpo en Markdown y frontmatter YAML estructurado.
  - Familias de Claves Especiales y Tipos en el Frontmatter:
    - *Familia de Identidad:* `id`, `title`, `type`, `domain`.
    - *Familia de Semántica W3C:* `skos_concept`, `dcat_dataset`, `wikidata_id`.
    - *Familia de Contratos Sintácticos:* `contract_type`, `contract_path`, `schema_hash`.
    - *Familia de Procedencia y Linaje:* `sources`, `prov_activity`, `prov_generated_by`.
    - *Familia de Confianza (Trust):* `generated`, `verified`, `trust_tier`, `signed_by`, `data_hash_sha256`.
    - *Familia de Ciclo de Vida:* `status` (draft, active, deprecated), `stale_after`.
  - Oferta de Datos:
    - *Oferta Estática:* Distribuciones físicas inmutables (Parquet, CSV) validadas contra Table Schema y hash SHA-256.
    - *Oferta Dinámica:* Conceptos de computación atestada (*Attested Computations*), endpoints y consultas parametrizadas.
  - Direccionamiento Canónico mediante Tríadas: `[catalog:dataset:resource]`.
  - Invariantes Normativos de Conformidad:
    - Invariantes del Productor: PE1 a PE5.
    - Invariantes del Consumidor: OE1 a OE5.
  - Esquema de Validación Formal JSON (`okf-concept.schema.json`).
- **4.2 Protocolo de Red P2P y Resolución Soberana ("DNS de Datos")**
  - Identidad criptográfica de nodos (`NodeID` y pares de claves Ed25519).
  - Gramática y sintaxis de resolución: `dataset.recurso@catalogo` (con utilidad de diagnóstico `okf-dig`).
  - Catálogo de registros soberanos tipados: `SOA` (autoridad), `CAT` (catálogo), `RES` (recurso y multiaddresses), `SCH` (contrato), `LIN` (linaje W3C PROV-O) y `SRV` (servicios streaming).
  - Pila de transporte multiprotocolo: QUIC/UDP, TCP+TLS, WebSockets y WebRTC.
  - Topología de red: autodescubrimiento local vía mDNS y difusión WAN con GossipSub.
- **4.3 Implementación Técnica del Motor de Ejecución (`datamesh-sdk`)**
  - Arquitectura Hexagonal del SDK:
    - *Núcleo del Dominio (`domain/`):* Bounded Contexts puros, Value Objects inmutables (`frozen=True`).
    - *Puertos de Entrada (`ports/inbound/`):* Casos de uso de carga, validación y consulta.
    - *Puertos de Salida (`ports/outbound/`):* Abstracciones de almacenamiento, catálogo y verificación.
    - *Adaptadores (`adapters/`):* Adaptador local, remoto HTTPS, Kaggle, S3, y motor embebido DuckDB.
  - Fachada unificada de 1 línea: `import datamesh as dm; df = dm.read(...)`.
  - Interfaz de Línea de Comandos (CLI `dm`).
  - Motor de Verificación Criptográfica: validación de firmas SSH/SHA-256 contra listas blancas institucionales.

---

### CAPÍTULO V: EVALUACIÓN, PRUEBAS Y RESULTADOS
*Este capítulo evalúa cuantitativa y cualitativamente tanto el contenedor como la herramienta de software.*

- **5.1 Estrategia de Evaluación y Verificación de Conformidad**
  - Pruebas de conformidad del contenedor: validador agnóstico en Go/Python contra esquemas JSON.
  - Pruebas unitarias de dominio del SDK (100% de éxito, herméticas).
  - Pruebas de contrato y verificación de integridad de referencias.
- **5.2 Demostración de Casos de Estudio Reales (Contexto Bolivia)**
  - Caso 1: Empaquetamiento de datos de Calidad del Aire de Bolivia como bundle OKF estático con linaje PROV-O.
  - Caso 2: Oferta dinámica de variables socioeconómicas mediante vistas SQL y transformaciones reproducibles.
  - Caso 3: Consulta federada entre dominios soberanos mediante DuckDB sin descarga redundante.
- **5.3 Benchmarks de Desempeño y Evaluación ISO/IEC 25010**
  - Eficiencia en tiempo de validación de bundles frente a catálogos tradicionales (CKAN).
  - Latencia de consulta y huella de memoria: DuckDB/Apache Arrow vs. Pandas.
  - Evaluación de completitud semántica y cumplimiento de principios FAIR.
- **5.4 Discusión de Resultados y Ventajas del Contenedor como Meta-Índice**

---

### CAPÍTULO VI: CONCLUSIONES Y RECOMENDACIONES
- **6.1 Conclusiones**
  - Logro de objetivos y validación del contenedor como índice abierto de estándares.
  - Viabilidad de la separación conceptual entre estándar agnóstico (`open-knowledge-format`) e implementación (`datamesh-sdk`).
  - Impacto de la investigación como aporte científico-técnico para la soberanía de datos en Bolivia.
- **6.2 Recomendaciones y Trabajo Futuro**
  - Adopción institucional en la Universidad de Aquino Bolivia (UDABOL).
  - Extensión hacia redes descentralizadas peer-to-peer y agentes autónomos de gobernanza.

---

### REFERENCIAS Y ANEXOS
- **Referencias Bibliográficas:** Normas APA (7ma Edición) e IEEE.
- **Anexo A:** Texto íntegro de la Especificación Normativa OKF v0.2 (`SPEC.md`).
- **Anexo B:** Esquemas JSON formales (`okf-concept.schema.json`).
- **Anexo C:** Catálogo de ADRs (Architecture Decision Records) de la Arquitectura.
- **Anexo D:** Manual de Referencia Técnica del SDK y CLI.
