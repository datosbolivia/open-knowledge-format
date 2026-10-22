# ESPECIFICACIÓN TÉCNICA Y PROBLEMÁTICA 09
## El Cuerpo Markdown como Grafo de Conocimiento, el Paradigma "LLM Wiki" (Karpathy) y la Interfaz Semántica Agéntica (MCP / CLI)

**Área de Especialidad:** Grafos de Conocimiento Ligeros, Agentes Autónomos de Curación (LLM Wiki), Model Context Protocol (MCP) e Interfaces Semánticas  
**Fecha:** Octubre 2026  
**Autor:** Andres Humberto Chirinos Lizondo  

---

## 1. Declaración de la Problemática

Las bases de conocimiento tradicionales y los enfoques contemporáneos de RAG (Retrieval-Augmented Generation) enfrentan un triple dilema:

1. **La Asfixia por Verbosidad Ontológica:** Los grafos de conocimiento basados en RDF/OWL/SPARQL son computacionalmente costosos, consumen un volumen excesivo de tokens en la ventana de contexto de los LLMs y son completamente ininteligibles para los analistas humanos sin visores especializados.
2. **La Pasividad Estática de los Repositorios de Datos:** Los repositorios actuales son "cementerios de datos" donde los archivos se suben una vez y quedan huérfanos. No existe un ciclo de vida en el que un agente de IA pueda no solo consultar, sino **dar mantenimiento continuo, podar conceptos obsoletos y enriquecer activamente la red de relaciones**.
3. **La Asimetría de Interfaces:** Los humanos disponen de wikis visuales (Obsidian, Notion), mientras que los agentes de IA se ven relegados a consultar APIs planas sin capacidad de navegar la topología de hipervínculos ni herramientas para auditar la salud del grafo.

---

## 2. Estado del Arte y Fundamentación Teórica

### 2.1 El Paradigma "LLM OS" y "Personal LLM Wiki" de Andrej Karpathy
* **El LLM como Núcleo de Cómputo (CPU):** Karpathy (2023, 2024) conceptualiza a los modelos de lenguaje grande no como simples chatbots, sino como la unidad central de procesamiento de un nuevo sistema operativo. En este modelo:
  - La ventana de contexto es la memoria RAM (rápida, volátil y costosa).
  - Los sistemas de archivos locales basados en texto estructurado (Markdown vaults) actúan como el almacenamiento persistente en Disco.
* **La Wiki Mantenida por LLMs (*Self-Improving Knowledge Base*):** En lugar de realizar búsquedas vectoriales ciegas sobre miles de fragmentos no estructurados, el conocimiento se organiza en una **red Zettelkasten de archivos Markdown interconectados**. 
  - Los agentes operan como **curadores activos**: leen notas, sintetizan nuevos conceptos a partir de datos frescos, crean enlaces cruzados contextuales, identifican contradicciones y reescriben los resúmenes de índice (`index.md`).

### 2.2 El Cuerpo Markdown como Grafo de Conocimiento (Knowledge Graph)
* **Nodos y Aristas Nativos en Texto:**
  - **Nodo (Entity):** Cada archivo `.md` en el bundle representa un concepto o producto de datos formal.
  - **Aristas Dirigidas (Edges):** Los enlaces markdown relativos `[Texto del Enlace](otro_concepto.md)` definen relaciones semánticas donde el texto del ancla actúa como el predicado o rol de la conexión.
  - **Subgrafos y Contexto:** Los encabezados jerárquicos (`#`, `##`, `###`) delimitan facetas y dimensiones analíticas sin necesidad de esquemas OWL hiper-complejos.
* **Eficiencia de Tokens y Legibilidad Tripartita:** Markdown reduce el consumo de tokens en más de un 70% comparado con serializaciones JSON-LD o XML equivalentes, permitiendo a los LLMs procesar grafos completos dentro de su ventana de atención inmediata.

---

## 3. Especificación del Protocolo Semántico de Doble Modo (MCP / CLI)

Para habilitar la interacción simétrica entre el agente y la base de conocimiento OKF, se estandariza una interfaz de dos modos operativos:

```
+-----------------------------------------------------------------------------------+
|                        AGENTE DE INTELIGENCIA ARTIFICIAL                          |
+-----------------------------------------------------------------------------------+
                       |                                    |
          (Lectura / Consulta)                     (Curación / Mantenimiento)
                       v                                    v
+-------------------------------+        +------------------------------------------+
|      MODO 1: REMOTO / CONSUMO  |        |      MODO 2: PROPIETARIO / MANTENEDOR    |
| * okf inspect <triad_or_uri>  |        | * okf init <path>                        |
| * okf graph <bundle_dir>      |        | * okf add <path> --type --title          |
| * okf query <sql_expression>  |        | * okf link <source> <target>             |
| * okf grep <semantic_pattern> |        | * okf index <bundle_dir> (Re-indexación) |
| * datamesh_sql_query          |        | * okf audit <bundle_dir> (Ciclos / Rotos)|
| * datamesh_get_dataproduct    |        | * okf sign <bundle_dir> -k <key>         |
+-------------------------------+        +------------------------------------------+
                       |                                    |
                       +-----------------+------------------+
                                         |
                                         v
                         +-------------------------------+
                         |   OKF KNOWLEDGE BUNDLE (MD)   |
                         |   (Frontmatter + Body Graph)  |
                         +-------------------------------+
```

### 3.1 Modo 1: Consumidor Remoto (ReadOnly)
El agente consulta sin alterar el repositorio:
1. `okf inspect <path_or_url>`: Retorna el frontmatter saneado, anclas ontológicas SKOS y contratos sintácticos.
2. `okf graph <bundle_dir>`: Devuelve la lista de adyacencia del grafo (nodos, aristas y enlaces externos) en JSON compacto para razonamiento topológico del LLM.
3. `datamesh_sql_query`: Ejecuta consultas SQL federadas DuckDB sobre los recursos respaldados por contratos.

### 3.2 Modo 2: Propietario y Mantenedor Activo (ReadWrite & Curate)
El agente cura y evoluciona la base de conocimiento:
1. `okf init <path>`: Inicializa un bundle conforme con `index.md`, `log.md` y concepto inicial.
2. `okf add <path> -type <type> -title <title>`: Crea un nuevo nodo conceptual con plantilla pre-validada.
3. `okf link <source.md> <target.md> -rel <predicate>`: Inserta un hipervínculo en el cuerpo markdown de `source.md` apuntando a `target.md`.
4. `okf index <bundle_dir>`: Re-escanea recursivamente el bundle y regenera `index.md` con divulgación progresiva.
5. `okf audit <bundle_dir>`: Ejecuta la suite de verificación (detecta enlaces rotos, referencias circulares **RULE-12**, conceptos huérfanos sin enlaces y conceptos expirados según `freshness.expires`).
6. `okf sign <bundle_dir> -k <key>`: Aplica firma digital OpenSSH / Ed25519 sellando el nuevo estado de la wiki curada.

---

## 4. Cuaderno de Investigación Bibliográfica (NotebookLM)

Para profundizar en el estado del arte empírico y contrastar las arquitecturas de memoria agéntica, se ha creado el cuaderno de investigación en **Google NotebookLM**:

* **ID del Cuaderno:** `b7eb311b-0b64-4ba6-82b7-9dbb809e0f7c`
* **Título:** *LLM Wiki de Karpathy y Mantenimiento Agéntico de Bases de Conocimiento*
* **Fuentes Primarias Indexadas:**
  1. *A Comparative Study of Memorial Architectures for Long Term AI Agents* (appliedAI, 2026).
  2. *Build a Personal LLM Knowledge Base in 15 Minutes* (Minibase, 2026).
  3. *Building a Self-Improving Personal Knowledge Base Powered by LLM* (Wang, 2024).
  4. *Building an 'Agent Only' Obsidian Vault* (Bedford / The Thinkers Club, 2025).
  5. *How to Build a Personal LLM Knowledge Base (Karpathy's Method)* (YouTube / Karpathy Lectures).
  6. *LLM Knowledge Bases: A Practical Guide* (Ben Holmes, 2025).
  7. *LLM OS: Andrej Karpathy's Vision for AI Operating Systems* (Tech Analysis, 2024).

---

## 5. Referencias Bibliográficas (APA 7ma / IEEE)

1. **Karpathy, A.** (2023). *Intro to Large Language Models: LLM OS Architecture and External Memory Patterns*. Berkeley AI Systems Lectures.
2. **Karpathy, A.** (2024). *Building a Second Brain with LLMs: The Markdown Wiki Paradigm*. Technical Keynote.
3. **appliedAI.** (2026). *A Comparative Study of Memorial Architectures for Long Term AI Agents*. AppliedAI Initiative Whitepapers.
4. **Bedford, J.** (2025). *Building an Agent-Only Obsidian Vault: Autonomous Graph Maintenance*. The Thinkers Club.
5. **Anthropic.** (2024). *Model Context Protocol (MCP) Specification: Open Protocols for Seamless Agent Tool Integration*. https://modelcontextprotocol.io/
6. **Holmes, B.** (2025). *LLM Knowledge Bases: A Practical Guide to Structuring Text Repositories for AI Retrieval*.
