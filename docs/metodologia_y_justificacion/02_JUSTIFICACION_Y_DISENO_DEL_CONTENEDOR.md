# JUSTIFICACIÓN INTEGRAL Y DISEÑO CONCEPTUAL DEL CONTENEDOR
## El Formato de Conocimiento Abierto como Meta-Índice de Estándares para el Intercambio Estático y Dinámico de Datos con Contexto

**Título del Proyecto de Grado:**  
# "Diseño e Implementación del Formato de Conocimiento Abierto Orientado a Datos"

**Postulante:** Andres Humberto Chirinos Lizondo  
**Carrera:** Ingeniería de Sistemas — Universidad de Aquino Bolivia (UDABOL)  
**Repositorio Normativo Matriz:** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)  
**Implementación Dependiente:** [`datamesh-sdk`](file:///home/andreschirinos/Proyectos/datamesh-sdk)

---

## 1. Planteamiento del Problema: La Fragmentación de los Estándares de Datos

En el ecosistema analítico moderno existen múltiples estándares de alta calidad técnica para resolver problemas puntuales:
- Para describir esquemas tabulares: **Frictionless Data Table Schema**.
- Para catalogar datasets y distribuciones: **W3C DCAT v2**.
- Para auditar el linaje y procedencia de ejecución: **W3C PROV-O**.
- Para estructurar conceptos y taxonomías: **W3C SKOS**.
- Para pipelines reproducibles: **Common Workflow Language (CWL)**.
- Para almacenamiento columnar y memoria eficiente: **Apache Parquet y Apache Arrow**.

Sin embargo, en el mundo real (y de manera especialmente crítica en instituciones públicas y académicas de Bolivia), **estos estándares viven completamente desarticulados**. La industria fuerza a las organizaciones a elegir entre dos extremos ineficientes:
1. **Adopción fragmentada y manual:** Los equipos de datos definen el esquema en un archivo, el linaje en otra herramienta (o en ningún lado), la documentación en un Wiki estático desactualizado y los datos en un bucket de nube sin contexto semántico. Cuando una columna cambia, los flujos analíticos se rompen en silencio (*silent pipeline failure*).
2. **Cautiverio tecnológico en plataformas propietarias (Vendor Lock-in):** Plataformas corporativas (Collibra, Dataplex, Snowflake, Unity Catalog) prometen unificar catálogo, esquema y gobernanza, pero a costa de forzar a las organizaciones a almacenar sus datos bajo formatos y sistemas cerrados, con costos anuales exorbitantes en moneda extranjera.

### La Pregunta Central de Investigación
> *¿Cómo diseñar un contenedor de conocimiento abierto, universal y agnóstico a plataformas, que actúe como un meta-índice unificado de los estándares existentes, permitiendo empaquetar e intercambiar datos acompañados de su contexto semántico, linaje auditable y contratos sintácticos, tanto para modalidades de acceso estático como dinámico?*

---

## 2. La Solución Conceptual: El Contenedor como Meta-Índice de Estándares

El **Formato de Conocimiento Abierto (OKF/ODKF v0.2)** no reinventa la rueda de los formatos tabulares ni inventa una ontología incompatible. Su innovación radica en operar como un **contenedor semántico universal y meta-índice de protocolos preexistentes**:

```
                       OPEN KNOWLEDGE CONTAINER (BUNDLE)
  +-------------------------------------------------------------------------+
  |  METADATOS GLOBALES & IDENTIDAD DEL DOMINIO                             |
  |  * metadata.yaml: Identidad de dominio, autores, licencias, versiones   |
  +-------------------------------------------------------------------------+
                                      |
         +----------------------------+----------------------------+
         v                                                         v
  +-----------------------------+           +------------------------------+
  |  CONTRATOS SINTÁCTICOS      |           |  SEMÁNTICA & TAXONOMÍAS      |
  |  (Índice de Formatos)       |           |  (Índice W3C)                |
  |  * Frictionless DataPackage |           |  * W3C DCAT v2 (Catálogos)   |
  |  * Table Schema             |           |  * W3C SKOS (Conceptos)      |
  |  * Arrow Schema / Parquet   |           |  * RDF/OWL (owl:sameAs)      |
  +-----------------------------+           +------------------------------+
         |                                                         |
         +----------------------------+----------------------------+
                                      |
         +----------------------------+----------------------------+
         v                                                         v
  +-----------------------------+           +------------------------------+
  |  PROCEDENCIA & LINAJE       |           |  CONFIANZA CRIPTOGRÁFICA     |
  |  (Índice de Linaje)         |           |  (Gobernanza Computacional)  |
  |  * W3C PROV-O (Activities)  |           |  * Hash SHA-256 inmutable    |
  |  * CWL v1.2 (Transformación)|           |  * Firma SSH (Ed25519)       |
  |  * Hash de transformaciones |           |  * Trust Tiers (Niveles)     |
  +-----------------------------+           +------------------------------+
                                      |
  +-------------------------------------------------------------------------+
  |                   MODALIDADES DE OFERTA DE DATOS                        |
  |  [ESTÁTICA]: Archivos Parquet/CSV inmutables direccionados por URIs     |
  |  [DINÁMICA]: Servicios, cómputos atestados, vistas SQL parametrizadas   |
  +-------------------------------------------------------------------------+
```

---

## 3. Oferta de Datos: Dualidad Estática y Dinámica con Contexto

El contenedor unifica dos formas esenciales de compartir datos que antes estaban divorciadas:

### 3.1. Oferta de Datos Estática (Static Data Offering)
- **Mecanismo:** El bundle referencia archivos inmutables (ej. distribuciones en Parquet, GeoParquet o CSV comprimido) alojados en almacenamiento distribuido (GitHub Releases, Kaggle, Google Drive, AWS S3 o servidores HTTP locales).
- **Contexto Semántico:** Cada archivo se asocia a un documento de concepto (`concepts/*.md`) que explica en lenguaje natural para qué sirve el dato, enlazado a un esquema estricto de **Table Schema** (`datapackage.json`).
- **Garantía de Inmutabilidad:** Se calcula y estipula la huella digital **SHA-256** (`data_hash_sha256`) en el frontmatter. Si un solo byte cambia en la fuente remota, el consumidor detecta la alteración antes de cargar el dato en memoria.

### 3.2. Oferta de Datos Dinámica (Dynamic Data Offering & Attested Computations)
- **Mecanismo:** Los datos no siempre son estáticos; a menudo se generan a través de consultas analíticas sobre almacenes transaccionales, modelos de inferencia o transformaciones complejas bajo demanda.
- **Cómputos Atestados (*Attested Computations*):** El contenedor especifica la consulta o script generador (ej. consulta DuckDB SQL, comando CLI o llamada a API), los parámetros de entrada y el hash del código transformador.
- **Contexto de Ejecución:** Utiliza la ontología **W3C PROV-O** para modelar la actividad dinámica (`prov_activity`), el software ejecutor (`prov_generated_by`) y el agente que autorizó el cómputo.
- **Vistas Parametrizadas y Ensayos Reproducibles:** El usuario puede invocar dinámicamente el dato sabiendo exactamente qué parámetros produjeron el cálculo y bajo qué condiciones de frescura (`stale_after`).

---

---

## 4. El Principio de Legibilidad Tripartita y el Catálogo Ultraligero (Estilo llms.txt)

Una justificación metodológica y de ingeniería crucial para la adopción del patrón **Open Knowledge Format (Google OKF)** en esta tesis de grado es el **Principio de Legibilidad Tripartita (Human / Machine / AI-Readable)**:

1. **Legibilidad Humana (Human-Readable):** Redactado en Markdown estándar. Cualquier ciudadano, estudiante o funcionario público puede leer el contenido en un editor de texto plano o navegador web sin instalar software costoso.
2. **Legibilidad por Máquinas (Machine-Readable):** El encabezado YAML frontmatter y los contratos en `datapackage.json` permiten que parsers y validadores tipados verifiquen esquemas y hashes SHA-256 de forma determinista.
3. **Legibilidad por Agentes de IA (AI-Readable):** Los modelos de lenguaje (LLM) consumen el archivo directamente en su ventana de contexto. Las referencias explícitas (`refs/`, vocabularios SKOS, identificadores de Wikidata) eliminan las alucinaciones al interactuar con datos.

### 4.1. Creación Ultraligera de Catálogos (Demostrado en datosbolivia.github.io/llms.txt)
Frente a la complejidad prohibitiva de portales pesados como CKAN o Socrata, OKF formaliza el **Perfil de Catálogo de Archivo Único** (análogo a `https://datosbolivia.github.io/llms.txt`):
- Una universidad (como la UDABOL), un ministerio o un municipio puede publicar su catálogo soberano en **GitHub Pages en menos de 2 minutos**.
- Solo requiere un archivo `llms.txt` o `index.md` con enlaces directos a los bundles de conceptos.
- El SDK `datamesh-sdk` es capaz de consumir e indexar este catálogo estático en una sola línea de código: `catalog = dm.load_catalog("https://datosbolivia.github.io/llms.txt")`.

---

## 5. Claves Especiales y Familias de Metadatos del Contenedor

Para que cualquier herramienta (sea un SDK en Python, una biblioteca en Go o un agente LLM) pueda intercambiar datos sin perder el contexto, el contenedor estandariza las siguientes familias de claves en el YAML frontmatter de los conceptos:

1. **Familia de Identidad y Tipos (`types`):**
   - `id`: Identificador canónico único en el dominio.
   - `type`: Clasificación del artefacto (ej. `concept`, `dataset`, `service`, `computation`).
   - `domain`: Bounded Context propietario del conocimiento.
2. **Familia de Semántica W3C (`semantic_anchors`):**
   - `dcat_dataset`: Enlace formal a la clase `dcat:Dataset`.
   - `skos_concept`: Término ontológico formal (`skos:Concept`) para desambiguar variables.
   - `wikidata_id`: Identificador persistente en la Web Semántica global (ej. `Q750` para Bolivia).
3. **Familia de Contratos Sintácticos (`contracts`):**
   - `contract_type`: Estándar utilizado (`frictionless_table_schema`, `json_schema`, `arrow_schema`).
   - `contract_path`: Ruta relativa al archivo de esquema.
   - `schema_hash`: Huella digital del esquema para prevenir derivas no autorizadas.
4. **Familia de Procedencia y Linaje (`provenance`):**
   - `sources`: Lista de fuentes primarias y documentos de referencia.
   - `prov_activity`: URI de la actividad de transformación según W3C PROV-O.
   - `prov_generated_by`: Identificador del script o pipeline (compatible con CWL).
5. **Familia de Confianza y Autoría Criptográfica (`trust`):**
   - `data_hash_sha256`: Hash del activo físico de datos.
   - `signed_by`: Identidad criptográfica (llave pública SSH o GitHub).
   - `trust_tier`: Nivel de verificación formal (`tier_1_verified`, `tier_2_community`, `tier_3_raw`).
6. **Familia de Ciclo de Vida (`lifecycle`):**
   - `status`: Estado del concepto (`draft`, `active`, `deprecated`).
   - `stale_after`: Fecha límite de vigencia para datos de alta rotación.

---

## 6. Jerarquía de Dependencia: `open-knowledge-format` gobierna a `datamesh-sdk`

Una regla metodológica fundamental del proyecto de grado es la estricta **jerarquía de dependencia**:

```
  ===========================================================================
  CAPA NORMATIVA (open-knowledge-format) — Fuente Única de Verdad
  ===========================================================================
  * Define el Contenedor y los Manifiestos.
  * Indexa los Estándares (W3C DCAT/PROV/SKOS, Frictionless, Arrow).
  * Establece los Invariantes PE1-PE5 (Productor) y OE1-OE5 (Consumidor).
  * Diseña los esquemas JSON de validación agnósticos.
                                     |
                                     | [Gobierna mediante contratos normativos]
                                     v
  ===========================================================================
  CAPA DE IMPLEMENTACIÓN (datamesh-sdk) — Motor Tecnológico
  ===========================================================================
  * Implementa la Arquitectura Hexagonal en Python.
  * Provee adaptadores de E/S físicos (Kaggle, Drive, S3, Local).
  * Ejecuta consultas analíticas con DuckDB y Apache Arrow.
  * Expone la fachada de 1 línea (`dm.read()`) y el CLI (`dm`).
```

Cualquier cambio conceptual, nueva familia de claves o nuevo estándar soportado DEBE nacer y formalizarse primero en `open-knowledge-format`. El repositorio `datamesh-sdk` no formula estándares de forma aislada; es un cliente e implementador disciplinado de la especificación.

---

## 7. Justificaciones Académicas para la Universidad de Aquino Bolivia (UDABOL)

### 6.1. Justificación Técnica
La separación entre estándar agnóstico e implementación en software garantiza la longevidad del sistema. Si en el futuro se desea construir un SDK en Rust o una interfaz web en WebAssembly, la especificación en `open-knowledge-format` permanecerá intacta y universal.

### 6.2. Justificación Económica
El enfoque del contenedor como meta-índice permite a universidades y entidades bolivianas federar datos sin adquirir licencias corporativas privativas millonarias ni depender de un proveedor de nube monopolístico.

### 6.3. Justificación Social y Académica
Permite a investigadores de la UDABOL, de la UMSA y de instituciones públicas compartir investigaciones científicas con contexto semántico garantizado, resolviendo el problema crónico de los datasets huérfanos que nadie puede reutilizar por falta de documentación de procedencia.

### 6.4. Justificación Metodológica (DSRM)
Al formular el contenedor como **Constructo** y **Modelo** de investigación científica formal bajo DSRM (March & Smith, 1995; Peffers et al., 2008), el proyecto cumple con los más altos estándares de rigor científico exigidos por los tribunales de titulación de Ingeniería de Sistemas.
