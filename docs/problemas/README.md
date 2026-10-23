# ÍNDICE GENERAL DE PROBLEMÁTICAS Y ESPECIFICACIONES TÉCNICAS
## Desglose de Desafíos Arquitectónicos del Formato de Conocimiento Abierto (OKF/ODKF v0.2)
### Especificaciones Normativas al Estilo de Grupos de Trabajo W3C / IETF RFC

**Título del Proyecto de Grado:**  
# "Diseño e Implementación del Formato de Conocimiento Abierto Orientado a Datos"
**Carrera:** Ingeniería de Sistemas — Universidad de Aquino Bolivia (UDABOL)  
**Autor:** Andres Humberto Chirinos Lizondo

---

## 1. Propósito de esta Sección

Siguiendo la metodología rigurosa de los consorcios internacionales de estandarización (**W3C Recommendations, IETF RFCs y Open Knowledge Foundation**), cada desafío fundamental que este formato aborda ha sido aislado en su **propia carpeta de especificación**.

Cada documento técnico contiene:
1. *Planteamiento formal de la problemática y casos de uso en el mundo real.*
2. *Revisión del estado del arte y estándares preexistentes (W3C DCAT v3, Frictionless, Kademlia, etc.).*
3. *Diseño de la solución normativa, sintaxis y reglas de validación en OKF.*
4. *Justificación arquitectónica y compensaciones (trade-offs).*
5. *Referencias bibliográficas académicas completas (normas APA 7ma / IEEE).*

---

## 2. Catálogo de Problemáticas Especificadas

| Carpeta / Problemática | Título Técnico | Estándares y Autores de Referencia |
| :--- | :--- | :--- |
| **[`01_federacion_recursiva_catalogos/`](01_federacion_recursiva_catalogos/ESPECIFICACION.md)** | **Federación Recursiva de Catálogos e Inter-referenciación (Catalog-of-Catalogs)** | W3C DCAT v3 (`dcat:Catalog`, `dcat:catalog`, `dcat:inCatalog`), Gaia-X Federation Services, DCAT-AP. Prevención de bucles y límites de recursión ($MAX\_DEPTH=8$). |
| **[`02_sindicacion_y_coreferencia_datasets/`](02_sindicacion_y_coreferencia_datasets/ESPECIFICACION.md)** | **Sindicación, Co-referencia y Múltiples Catálogos Referenciando el Mismo Dataset (Many-to-Many)** | W3C DCAT v3 (`dcat:CatalogRecord`, `foaf:primaryTopic`), `owl:sameAs`, Content-Addressing (CID SHA-256 inmutable para desduplicación de almacenamiento en red). |
| **[`03_jerarquia_dataset_recursos_subdominios/`](03_jerarquia_dataset_recursos_subdominios/ESPECIFICACION.md)** | **Jerarquía Semántica: El Dataset como Subdominio y Recursos Dependientes (Estáticos y Dinámicos)** | Domain-Driven Design (Eric Evans - Bounded Contexts y Aggregate Roots), W3C DCAT (`dcat:Dataset` vs `dcat:Distribution`), Apache Arrow Flight SQL y DuckDB streaming. |
| **[`04_recursos_compuestos_archivos_anidados/`](04_recursos_compuestos_archivos_anidados/ESPECIFICACION.md)** | **Recursos Compuestos y Archivos Anidados (Archives ZIP, TAR, Multipart y Fragment Identifiers)** | W3C DCAT v3 (`dcat:packageFormat`, `dcat:compressFormat`, `spdx:checksum`), Frictionless Multipart Data Package, RFC 3986 URI Fragments (`#subarchivo`), lectura selectiva *zero-copy* por HTTP Range Requests. |
| **[`05_red_p2p_y_dns_soberano/`](05_red_p2p_y_dns_soberano/ESPECIFICACION.md)** | **Red P2P, Direccionamiento DNS Soberano y Subespacios de Información en Internet** | Kademlia XOR metric (Maymounkov & Mazières), Roos et al. (cadenas de Markov para saltos), S/Kademlia (crypto-puzzles), EPFL 4-Layer P2P-IR (Aberer et al.), Soft/Hard Partitioning (Protocol Labs) y Data Spaces Compass (DSSC). |
| **[`06_contratos_datos_y_politicas_gobernanza/`](06_contratos_datos_y_politicas_gobernanza/ESPECIFICACION.md)** | **Contratos de Datos, Calidad Sintáctica y Políticas de Gobernanza Cero-Microdatos** | Sanderson, Freeman & Schmidt (Data Contracts, 2025), Open Policy Agent (OPA/Rego), Frictionless Table Schema, políticas `zero_microdata` e invariantes PE1-PE5 / OE1-OE5. |
| **[`07_legibilidad_tripartita_y_catalogos_ligeros/`](07_legibilidad_tripartita_y_catalogos_ligeros/ESPECIFICACION.md)** | **Legibilidad Tripartita (Human/Machine/AI-Readable), Traducción Inter-Estándares y Catálogos Ultraligeros (Estilo llms.txt)** | Google OKF Pattern, Simon Willison (`llms.txt`), W3C DCAT v3 / Frictionless translation, despliegue estático en 2 min en GitHub Pages (`https://datosbolivia.github.io/llms.txt`). |
| **[`08_cuerpo_markdown_traduccion_lossless_y_w3c_baseline/`](08_cuerpo_markdown_traduccion_lossless_y_w3c_baseline/ESPECIFICACION.md)** | **Cuerpo Markdown Cognitivo Humano/IA, Matriz de Traducción Lossless y Núcleo Semántico W3C** | W3C DCAT v3, PROV-O, SKOS, Dublin Core, Frictionless Data Package, Context Engineering para LLMs. |
| **[`09_grafo_conocimiento_markdown_llm_wiki_y_mcp_semantico/`](09_grafo_conocimiento_markdown_llm_wiki_y_mcp_semantico/ESPECIFICACION.md)** | **Grafo de Conocimiento Markdown, Paradigma LLM Wiki y Protocolo MCP Semántico de Doble Modo** | Andrej Karpathy (LLM OS & Markdown Wiki), appliedAI (Memorial Architectures), Bedford (Obsidian Vaults), Anthropic MCP. |
| **[`10_ejecutor_scripts_linaje_background_y_sqlite/`](10_ejecutor_scripts_linaje_background_y_sqlite/ESPECIFICACION.md)** | **Ejecución Monitoreada de Scripts/Funciones, Cálculo de Linaje en Background y Memoria SQLite** | W3C PROV-O, OpenLineage, SQLite Application Format (Hipp et al.), Data Contracts, Publicación Asistida por Agentes MCP. |

---

## 3. Articulación con la Tesis de Grado (UDABOL)

Este desglose de problemáticas nutre directamente la estructura de la tesis:
- **Capítulo II (Marco Teórico y Estado del Arte):** Las secciones 2 de cada especificación compilan la literatura científica y estándares mundiales.
- **Capítulo IV (Ingeniería del Proyecto):**
  - *Subcapítulo IV.1:* Las especificaciones normativas de las carpetas 01, 02, 03, 04 y 06 definen el estándar `open-knowledge-format`.
  - *Subcapítulo IV.2:* La especificación de red P2P de la carpeta 05 define el protocolo de transporte y descubrimiento.
  - *Subcapítulo IV.3:* La implementación técnica en `datamesh-sdk` materializa los algoritmos de resolución, desduplicación, lectura parcial de ZIPs y verificación criptográfica.
- **Capítulo V (Evaluación y Resultados):** Las reglas de validación e invariantes normativos sirven como criterios cuantitativos de evaluación empírica.

---

## 3. Mapeo con el Marco de Interoperabilidad en Cuatro Niveles (EIF & FAIR)

Cada una de las 7 problemáticas abordadas se inscribe en las cuatro capas canónicas de interoperabilidad:

| Nivel de Interoperabilidad (EIF) | Problemáticas Asociadas | Estándares Clave Involucrados |
| :--- | :--- | :--- |
| **1. Técnica** | [Prob. 04: Recursos Compuestos y ZIPs](04_recursos_compuestos_archivos_anidados/ESPECIFICACION.md)<br>[Prob. 05: Red P2P y DNS Soberano](05_red_p2p_y_dns_soberano/ESPECIFICACION.md) | HTTP Range Requests (RFC 7233), Kademlia DHT, Apache Arrow Flight, ZIP Central Directory. |
| **2. Sintáctica** | [Prob. 06: Contratos de Datos](06_contratos_datos_y_politicas_gobernanza/ESPECIFICACION.md)<br>[Prob. 07: Legibilidad Tripartita y llms.txt](07_legibilidad_tripartita_y_catalogos_ligeros/ESPECIFICACION.md)<br>[Prob. 08: Traducción Lossless y Extras](08_cuerpo_markdown_traduccion_lossless_y_w3c_baseline/ESPECIFICACION.md) | Frictionless Data Package, OpenAPI 3.1, Apache Parquet, YAML Frontmatter, llms.txt, Schema Preservation. |
| **3. Semántica** | [Prob. 01: Federación Recursiva de Catálogos](01_federacion_recursiva_catalogos/ESPECIFICACION.md)<br>[Prob. 02: Sindicación y Co-referencia](02_sindicacion_y_coreferencia_datasets/ESPECIFICACION.md)<br>[Prob. 03: Jerarquía Dataset/Recursos](03_jerarquia_dataset_recursos_subdominios/ESPECIFICACION.md)<br>[Prob. 08: Núcleo Semántico W3C y Markdown Body](08_cuerpo_markdown_traduccion_lossless_y_w3c_baseline/ESPECIFICACION.md)<br>[Prob. 09: Markdown Knowledge Graph y LLM Wiki](09_grafo_conocimiento_markdown_llm_wiki_y_mcp_semantico/ESPECIFICACION.md) | W3C DCAT v3, W3C PROV-O, W3C SKOS, DCMI, Paradigma LLM OS / Personal LLM Wiki (Karpathy), MCP Semántico, Tríadas Canónicas. |
| **4. Organizativa / Jurídica** | [Prob. 06: Políticas Zero-Microdata y Gobernanza](06_contratos_datos_y_politicas_gobernanza/ESPECIFICACION.md)<br>[Prob. 10: Linaje en Background y SQLite](10_ejecutor_scripts_linaje_background_y_sqlite/ESPECIFICACION.md) | Licencias abiertas FAIR (CC/ODbL), Habeas Data, Firmas OpenSSH/Ed25519, W3C PROV-O, OpenLineage, Trazabilidad SLSA. |

Para más detalles, consultar [`docs/INTEROPERABILIDAD_EN_CUATRO_NIVELES.md`](../INTEROPERABILIDAD_EN_CUATRO_NIVELES.md).
