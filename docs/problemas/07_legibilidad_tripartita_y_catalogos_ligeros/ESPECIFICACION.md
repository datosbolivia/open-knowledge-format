# ESPECIFICACIÓN TÉCNICA Y PROBLEMÁTICA 07
## Legibilidad Tripartita (Human / Machine / AI-Readable), Traducción Inter-Estándares y Catálogos Ultraligeros (Estilo llms.txt)

**Estado:** Documento de Especificación Normativa (Estilo W3C Working Draft)  
**Módulo:** Usabilidad Universal, Traducción Semántica y Perfil de Catálogo Ultraligero OKF  
**Repositorio:** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)  
**Referencia de Producción:** [`https://datosbolivia.github.io/llms.txt`](https://datosbolivia.github.io/llms.txt)  
**Cuaderno de Investigación:** [DATAMESH](https://notebook.google.com/notebook/52ebcfe9-f328-4885-834f-f1f2f4d7bc4b) (`52ebcfe9-f328-4885-834f-f1f2f4d7bc4b`)

---

## 1. Declaración de la Problemática

Históricamente, los sistemas de gestión y catalogación de datos han padecido un sesgo excluyente:
1. **La Pesadez Operativa de los Portales Tradicionales (CKAN, Socrata, Collibra):** Publicar un catálogo de datos requería instalar servidores dedicados, bases de datos PostgreSQL/PostGIS complejas, motores de búsqueda Solr y paneles de administración pesados. Para una alcaldía pequeña en Bolivia, un laboratorio universitario de la UDABOL o un investigador individual, el costo de infraestructura y mantenimiento técnico resulta inalcanzable.
2. **La Ruptura de Lecturabilidad (El Dilema Humano vs. Máquina):**
   - Formatos como RDF/XML o dumps SPARQL son procesables por máquinas pero ilegibles para humanos comunes.
   - Páginas web HTML o documentos PDF son amigables para humanos pero opacos y frágiles para analizadores sintácticos automáticos.
3. **La Ceguera y Alucinación de la Inteligencia Artificial (AI/LLM-Hostile):** Los modelos de lenguaje y agentes autónomos sufren alucinaciones cuando se les presentan APIs complejas o datos sin contexto conceptual explícito. Los agentes necesitan formatos concisos, estructurados y de bajo consumo de tokens para realizar búsquedas fundamentadas (*grounded search*) sin desbordar su ventana de contexto.

### El Desafío Arquitectónico:
> *¿Cómo diseñar un formato que actúe como un traductor universal entre estándares existentes, satisfaciendo simultáneamente el triple principio de legibilidad (Human / Machine / AI-Readable), y que permita a cualquier usuario u organismo crear un catálogo de datos soberano en menos de 2 minutos utilizando únicamente un archivo de texto plano (estilo `llms.txt` / `index.md`) desplegable en GitHub Pages o cualquier servidor estático?*

---

## 2. El Principio de Legibilidad Tripartita (Tri-Readable Architecture)

El Formato de Conocimiento Abierto (OKF) adopta la filosofía de diseño originada en el patrón Open Knowledge de Google, perfeccionándola con anclajes semánticos de dominio y referencias cruzadas:

```
                     EL PARADIGMA TRI-READABLE (OKF)

  +-------------------------------------------------------------------------+
  |                             DOCUMENTO OKF                               |
  +-------------------------------------------------------------------------+
  |  YAML FRONTMATTER (Delimitado por '---')                                |
  |  * Tipado estricto, esquemas JSON, hashes SHA-256, URIs W3C              |
  |  * DESTINATARIO: MÁQUINAS Y PARSERS TIPADOS                             |
  +-------------------------------------------------------------------------+
  |  CUERPO EN MARKDOWN (Texto Plano Estructurado)                          |
  |  * Narrativa de negocio, tablas, advertencias de uso, fórmulas          |
  |  * DESTINATARIO: HUMANOS (Lectura en cualquier editor o terminal)       |
  +-------------------------------------------------------------------------+
  |  VÍNCULOS RELATIVOS, REFS/ Y METADATOS DE ANCLAJE                       |
  |  * Tríadas [catalog:dataset:resource], links directos, grounding estricto|
  |  * DESTINATARIO: AGENTES DE IA / LLMS (Context Ingestion Zero-Alucinac.)|
  +-------------------------------------------------------------------------+
```

### 2.1. Virtudes por Actor del Sistema
1. **Para Humanos (Human-Readable):** Cero dependencia de herramientas propietarias. Cualquier persona puede abrir un archivo `.md` en Notepad, VS Code, Vim o visualizarlo en GitHub sin compilar nada.
2. **Para Máquinas (Machine-Readable):** El frontmatter es directamente convertible a JSON/YAML y validable contra `okf-concept.schema.json`. Los analizadores sintácticos extraen tipos de datos, hashes criptográficos y contratos sin ambigüedad.
3. **Para Agentes de IA y LLMs (AI-Readable):** Un agente puede cargar el documento directamente en su *context window*. Las claves semánticas (`skos_concept`, `wikidata_id`, `contract_type`) proveen a la IA la base de conocimiento (*knowledge grounding*) necesaria para generar consultas SQL correctas sin alucinar esquemas.

---

## 3. El Formato como Traductor Universal entre Estándares

El contenedor OKF no compite con los estándares preexistentes; **los indexa y traduce entre sí**:

```
                       OKF COMO MOTOR DE TRADUCCIÓN

    W3C DCAT v3             Frictionless Table Schema          W3C PROV-O
  (Catálogos / LOD)           (Tipado Tabular / CSV)        (Linaje y Auditoría)
          │                             │                            │
          └─────────────────────────────┼────────────────────────────┘
                                        │
                                        v
                    +---------------------------------------+
                    |    OPEN KNOWLEDGE FORMAT (OKF)        |
                    |    [El Contenedor Traductor Común]    |
                    +---------------------------------------+
                                        │
          ┌─────────────────────────────┼────────────────────────────┐
          v                             v                            v
    Apache Arrow / Parquet        DuckDB SQL Views             OpenAPI REST
  (Cómputo en Memoria Zero-Copy) (Consultas Analíticas Edge) (Servicios Web HTTP)
```

- **De DCAT a Frictionless:** Traduce una `dcat:Distribution` física hacia un recurso validado con columnas tipadas en `Table Schema`.
- **De Frictionless a Apache Arrow:** Proyecta los tipos de datos de Frictionless (`string`, `number`, `datetime`) a tipos nativos en memoria de Apache Arrow para procesamiento de cero-copia.
- **De CWL a PROV-O:** Convierte la declaración de ejecución de una herramienta en un grafo de procedencia `prov:Activity` auditable.

---

## 4. Creación Ultraligera de Catálogos: El Patrón `llms.txt` / `index.md`

Inspirado en la iniciativa global `llms.txt` y demostrado en producción en [`https://datosbolivia.github.io/llms.txt`](https://datosbolivia.github.io/llms.txt), OKF formaliza el **Perfil de Catálogo Ultraligero**:

### 4.1. Filosofía: El Catálogo de Archivo Único
Una institución, carrera universitaria o laboratorio no necesita bases de datos para publicar su catálogo soberano. **Un único archivo de texto plano es suficiente**:

```markdown
# Catálogo de Investigación Científica — UDABOL
> Catálogo Soberano de Productos de Datos de la Universidad de Aquino Bolivia

Este catálogo expone los productos de datos abiertos conforme al estándar OKF y llms.txt, permitiendo a investigadores, máquinas y agentes de IA descubrir y consultar datos con rigor formal.

## Información Institucional
- Autoridad Soberana: Universidad de Aquino Bolivia (UDABOL)
- Nodo P2P PeerID: 12D3KooWUDABOLResearchNode...
- Clave Pública SSH: ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG5...

## Catálogo de Datasets (Open Knowledge Format v0.2)
- [Monitoreo de Calidad del Aire La Paz (PM2.5)](/nodes/calidad_aire/index.md): Mediciones horarias del material particulado PM2.5 en estaciones urbanas. (Dominio: Medio Ambiente. Recursos: pm25_horario, historico_2025_parquet. Contrato: Table Schema. Hash: sha256:e3b0c44...)
- [Indicadores Socioeconómicos Regionales](/nodes/economia/index.md): Series de tiempo de canasta básica e IPC departamental. (Dominio: Economía. Recursos: ipc_mensual. Contrato: Table Schema. Hash: sha256:4b22777...)
```

### 4.2. Ventajas Radicales para Bolivia y la Academia
1. **Despliegue en 2 minutos a Costo Cero:** Se aloja en **GitHub Pages**, GitLab Pages o un servidor Nginx/Apache estático básico.
2. **Mantenimiento GitOps:** Modificar el catálogo es tan simple como hacer un `git commit` y `git push`. La historia de cambios queda preservada criptográficamente en el historial de Git.
3. **Indexación Inmediata por LLMs:** Los navegadores de agentes (como Claude, Gemini, ChatGPT) leen `llms.txt` de forma nativa en la raíz del dominio (`https://institucion.edu.bo/llms.txt`), descubriendo instantáneamente todos los datasets disponibles sin necesidad de scrapers frágiles.

---

## 5. Gramática del Perfil Ultraligero OKF-Lite

Un archivo `llms.txt` o `index.md` conforme con OKF-Lite DEBE cumplir con:

1. **Cabecera H1 con Nombre del Catálogo:** `# <Nombre del Catálogo>`
2. **Bloque Blockquote descriptivo:** `> <Propósito y alcance del catálogo>`
3. **Sección H2 de Documentación:** `## Información y Documentación del Proyecto`
4. **Sección H2 de Datasets Federados:** `## Catálogo de Datasets Federados`
5. **Entradas de Lista con Tríadas y Enlaces Relativos:**
   ```markdown
   - [<Título del Dataset>](<ruta_relativa_o_url>): <Descripción concisa>. (Dominio: <Nombre>. Recursos: <recurso_1>, <recurso_2>)
   ```

Los motores como `datamesh-sdk` pueden consumir directamente una URL de `llms.txt`:
```python
import datamesh as dm

# Carga automática del catálogo estático ultraligero
catalog = dm.load_catalog("https://datosbolivia.github.io/llms.txt")
df = catalog.read("air_quality:Compilación de datos de calidad del aire de Bolivia")
```

---

## 6. Referencias Bibliográficas (APA 7ma / IEEE)

* **Willison, S. (2024).** *llms.txt: A proposed standard for providing website context to LLMs*. Simon Willison’s Weblog. https://simonwillison.net/2024/Sep/8/llmstxt/
* **Google. (2024).** *Open Knowledge Format (OKF): A human, machine, and AI-readable knowledge standard*. Google Open Source.
* **Open Knowledge Foundation. (2024).** *Frictionless Data: Specifications for Data Packaging and Lightweight Catalogues*.
* **Berners-Lee, T. (1998).** *Cool URIs don't change*. W3C Style Guide. https://www.w3.org/Provider/Style/URI
* **Chirinos Lizondo, A. H. (2026).** *Portal Catálogo de Datos Abiertos de Bolivia (datosbolivia.github.io)*. Proyecto de Ciencia de Datos Soberana.
