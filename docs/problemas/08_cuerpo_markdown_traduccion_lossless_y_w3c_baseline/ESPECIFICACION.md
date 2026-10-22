# ESPECIFICACIÓN TÉCNICA Y PROBLEMÁTICA 08
## El Cuerpo Markdown como Interfaz Operativa Humano/IA, Traducción de Formatos con Pérdida Mínima y Núcleo Semántico W3C por Defecto con Extensiones Abiertas

**Área de Especialidad:** Ingeniería de Contexto, Interfaz Humano-Agente, Mapeo de Esquemas y Web Semántica  
**Fecha:** Octubre 2026  
**Autor:** Andres Humberto Chirinos Lizondo  

---

## 1. Declaración de la Problemática

En la interacción contemporánea entre humanos, agentes de inteligencia artificial y motores de bases de datos analíticas, se manifiestan tres fricciones críticas:

1. **La Reducción del Documento a Mero Contenedor de Metadatos:** Los sistemas tradicionales tratan el archivo de metadatos como una estructura rígida (XML, JSON, YAML puro) desprovista de contexto discursivo. Al ignorar el **cuerpo en lenguaje natural (Markdown body)**, los agentes LLM y los analistas humanos carecen de la guía cognitiva necesaria para entender el propósito del dato, sus sesgos metodológicos, sus advertencias de calidad y las instrucciones operativas para interactuar con las herramientas que resuelven los contratos sintácticos.
2. **Degradación y Pérdida de Metadatos en la Traducción (*Information Loss in Format Bridges*):** Cuando un producto de datos se exporta entre ecosistemas heterogéneos (ej. OKF $\leftrightarrow$ W3C DCAT-AP $\leftrightarrow$ Frictionless Data Package $\leftrightarrow$ CKAN $\leftrightarrow$ Apache Arrow), los traductores ingenuos descartan campos no reconocidos, destruyendo la proveniencia, las atestaciones y las restricciones de integridad sintáctica.
3. **Fragmentación y Conflicto de Diccionarios de Datos:** La ausencia de un vocabulario canónico por defecto genera islas semánticas; no obstante, forzar un esquema monolítico e inflexible impide que dominios especializados (salud, finanzas, censos) incorporen sus propios atributos terminológicos.

---

## 2. Estado del Arte y Fundamentación Teórica

### 2.1 Ingeniería de Contexto y Markdown como Interfaz Agente-Humano
* **Evidencia Técnica:** La evolución de la interacción humano-IA ha transitado del *Prompt Engineering* a la **Ingeniería de Contexto** (Anthropic, 2024; Cloudflare, 2024). Los Modelos de Lenguaje Grande (LLMs) procesan Markdown como su representación nativa óptima debido a su eficiencia de tokens y estructuración jerárquica clara (encabezados `#`, listas, bloques de código cercados y tablas).
* **El Cuerpo Markdown como Andamio Operativo:** Mientras que el bloque YAML frontmatter es analizado por los parsers computacionales para resolver contratos de bajo nivel (Frictionless, Parquet, DuckDB), el **cuerpo Markdown** actúa como la interfaz interactiva donde el agente y el humano comprenden *cómo usar* las herramientas:
  - Definición cualitativa de dimensiones y atributos.
  - Ejemplos ejecutables de consultas SQL y llamadas a herramientas MCP (`SELECT ...`, `datamesh_sql_query`).
  - Advertencias metodológicas (sesgos muestrales, truncamientos, cambios legislativos).

### 2.2 Mapeo de Esquemas y Métricas de Preservación de Información
* **Fundamentación en Preservación de Metadatos:** La literatura en integración de datos (Doan, Halevy & Ives, 2012) define que un mapeo entre esquemas $S_1 \to S_2$ es mínimamente disruptivo (*near-lossless*) si preserva:
  1. Identidad global (URIs y claves canónicas).
  2. Tipado primitivo e invariantes de dominio (restricciones de nulos, rangos, claves foráneas).
  3. Metadatos de proceso (paradatos, linaje y actores).
* **Solución W3C y Extensiones Abiertas:** W3C DCAT v3 y Frictionless Data Package establecen propiedades estándar básicas, pero contemplan extensiones no destructivas (`dcat:distribution`, `custom` attributes y bloques `extras` en CKAN).

---

## 3. Especificación en el Formato de Conocimiento Abierto (OKF v0.2)

### 3.1 Estructura Normativa del Cuerpo Markdown
En un documento de concepto OKF, el cuerpo bajo el delimitador de cierre `---` DEBE estructurarse siguiendo las siguientes secciones canónicas:

```markdown
---
# Frontmatter YAML (Contratos, Tipos, Proveniencia, Semántica)
type: dataset
title: Censo de Población y Vivienda 2024
contracts:
  - type: datapackage
    ref: "datapackage.json"
---

# Censo de Población y Vivienda 2024

## 1. Descripción y Propósito del Conocimiento
Explicación contextual en prosa humana del dominio, origen de los datos y relevancia analítica.

## 2. Metodología, Alcance y Advertencias de Calidad
- **Unidad de Análisis:** Vivienda particular y persona residente habitual.
- **Limitaciones Conocidas:** Cobertura geográfica en zonas no cartografiadas previamente.
- **Política de Privacidad:** Cumplimiento de regla Zero-Microdata (Habeas Data).

## 3. Interacción Operativa y Consultas de Ejemplo
Instrucciones explícitas para agentes y humanos sobre cómo consultar este producto usando herramientas analíticas (DuckDB / MCP):

\`\`\`sql
-- Consulta de población agregada por departamento
SELECT departamento, SUM(poblacion_total) as total 
FROM "bolivia:censo_2024:personas" 
GROUP BY departamento 
ORDER BY total DESC;
\`\`\`

## 4. Diccionario de Datos y Reglas de Negocio
Detalle textual de campos, unidades de medida y equivalencias ontológicas.
```

### 3.2 El Núcleo Semántico W3C por Defecto
El estándar OKF v0.2 adopta oficialmente la suite de recomendaciones **W3C** como su **diccionario mínimo recomendado por defecto**:

| Estándar W3C | Rol en OKF | Mapeo Canónico en Frontmatter |
| :--- | :--- | :--- |
| **W3C DCAT v3** | Vocabulario de Catálogos y Recursos | `dcat:Catalog`, `dcat:Dataset`, `dcat:Distribution`, `dcat:DataService`, `dcat:DatasetSeries`. |
| **W3C PROV-O** | Ontología de Proveniencia | `generated.by` $\to$ `prov:wasGeneratedBy`<br>`generated.from` $\to$ `prov:wasDerivedFrom`<br>`verified.by` $\to$ `prov:wasAttributedTo`. |
| **W3C SKOS** | Anclajes Conceptuales y Taxonomías | `skos: [...]` con URIs desreferenciables a tesauros. |
| **DCMI / Dublin Core** | Metadatos Nucleares Universales | `dcterms:title`, `dcterms:description`, `dcterms:license`, `dcterms:issued`, `dcterms:spatial`, `dcterms:temporal`. |

### 3.3 Extensibilidad Abierta (*Open Extensibility & Custom Extras*)
Para evitar la tiranía de un esquema rígido:
- Cualquier concepto OKF puede incluir propiedades adicionales arbitrarias bajo el bloque `extra:` (o campos clave/valor en el nivel raíz del frontmatter).
- El motor de validación Go y el runtime Python **no descartan ni rechazan campos no reconocidos**; los preservan en memoria en el mapa `Extra map[string]interface{}`.
- Durante la exportación a otros formatos:
  - En **CKAN**: Se empaquetan en el diccionario `extras: [{key, value}]`.
  - En **Frictionless**: Se colocan en propiedades de extensión a nivel de paquete o recurso.
  - En **W3C DCAT (JSON-LD / Turtle)**: Se serializan como predicados RDF personalizados conservando su namespace.

---

## 4. Matriz de Traducción de Pérdida Mínima (Cross-Standard Lossless Matrix)

| Atributo OKF / ODKF v0.2 | W3C DCAT v3 | Frictionless DataPackage | CKAN Portal API | Apache Parquet / Arrow |
| :--- | :--- | :--- | :--- | :--- |
| `title` | `dct:title` | `title` | `title` | Arrow Schema Metadata (`title`) |
| `description` | `dct:description` | `description` | `notes` | Arrow Schema Metadata (`description`) |
| `skos` (anclajes) | `dcat:theme` / `dcat:keyword` | `keywords` / `custom.skos` | `tags` / `extras.skos` | Parquet File Meta (`skos_uris`) |
| `contracts[datapackage]` | `dcat:distribution` | Root DataPackage | Resource schema | Native Arrow Table Schema |
| `contracts[openapi]` | `dcat:accessService` | `custom.api` | Resource URL (format API) | N/A (Dynamic Endpoint) |
| `lineage.sources` | `prov:wasDerivedFrom` | `sources[]` | `extras.lineage_sources` | Arrow Field Meta (`sources`) |
| `generated.by` | `prov:wasGeneratedBy` | `contributors[]` | `author` / `maintainer` | Arrow Meta (`author`) |
| `policy` (zero-microdata) | `odrl:hasPolicy` | `custom.policy` | `extras.policy` | Execution Guard Filter |
| `extra` (campos custom) | Extended RDF Predicates | Extended properties | `extras` key-values | Arrow Schema Meta Key-Values |

---

## 5. Referencias Bibliográficas (APA 7ma / IEEE)

1. **Anthropic.** (2024). *Context Engineering for Large Language Model Agents: Principles and System Patterns*. Anthropic Technical Reports.
2. **Doan, A., Halevy, A., & Ives, Z.** (2012). *Principles of Data Integration*. Morgan Kaufmann.
3. **Open Knowledge Foundation.** (2021). *Frictionless Data Specifications: Data Package and Table Schema*. OKFN Specifications. https://specs.frictionlessdata.io/
4. **W3C.** (2024). *Data Catalog Vocabulary (DCAT) - Version 3*. W3C Recommendation. World Wide Web Consortium. https://www.w3.org/TR/vocab-dcat-3/
5. **W3C.** (2013). *PROV-O: The PROV Ontology*. W3C Recommendation. World Wide Web Consortium. https://www.w3.org/TR/prov-o/
6. **W3C.** (2009). *SKOS Simple Knowledge Organization System Reference*. W3C Recommendation. World Wide Web Consortium. https://www.w3.org/TR/skos-reference/
7. **Cloudflare.** (2024). *Content Negotiation at the Edge: Dynamic Markdown Delivery for AI Crawlers and LLMs*. Cloudflare Architecture Blog.
