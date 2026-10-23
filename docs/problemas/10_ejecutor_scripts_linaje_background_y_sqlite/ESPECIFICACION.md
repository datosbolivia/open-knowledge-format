# ESPECIFICACIÓN TÉCNICA 10
# Ejecución Monitoreada de Scripts/Funciones, Cálculo de Linaje en Segundo Plano y Memoria Operacional SQLite para Publicación Asistida por Agentes

**Estado:** Aprobado / Normativo  
**Fecha:** 2026-10-22  
**Estándares de Referencia:** W3C PROV-O, OpenLineage Standard, SQLite Application File-Format (RFC/Hipp), Frictionless Data Package, Model Context Protocol (MCP).

---

## 1. Planteamiento Formal del Problema

En la ingeniería de datos contemporánea, la captura de procedencia y linaje epistémico adolece de una brecha insalvable entre la teoría y la práctica:
1. **La Paradoja de la Carga Manual:** Los estándares de metadatos (como W3C PROV-O o Frictionless) asumen que los autores o ingenieros transcribirán manualmente en archivos YAML/JSON las fuentes aguas arriba (*upstream*), los scripts intermedios y las huellas criptográficas. En la práctica, esto genera *desfase epistemológico* (provenance drift): los scripts evolucionan, pero los manifiestos quedan obsoletos.
2. **Ceguera Operacional en Tiempo de Ejecución:** Cuando un script en Python, R, DuckDB SQL o Shell transforma datos brutos en un producto final, el entorno de ejecución no captura intrínsecamente los artefactos leídos, los tiempos de procesamiento ni los parámetros aplicados, perdiéndose la trazabilidad científica.
3. **Falta de Memoria Local Persistente Ligera:** Sin un almacenamiento de estado local embebido y de cero configuración, la reconstrucción de grafos dirigidos de derivación (DAG) requiere escanear y re-parsear archivos de texto masivos en cada interacción, saturando la latencia del sistema.
4. **La Desconexión del Agente Curador:** Los agentes de IA son eficaces para razonar, auditar y publicar información, pero carecen de una memoria operativa estructurada donde inspeccionar qué hizo un script antes de autorizar o sugerir la publicación formal hacia catálogos federados.

---

## 2. Solución Normativa: Motor de Linaje en Background y Memoria SQLite

Se establece un subsistema de ejecución dual y persistencia relacional embebida:

```
+---------------------------------------------------------------------------------+
|                                CAPA DE EJECUCIÓN                                |
|                                                                                 |
|   MODALIDAD 1: Terminal / CLI              MODALIDAD 2: SDK / Funciones        |
|   $ okf run script.py --target ds:res      @dm.track_lineage(target="ds:res")   |
|   $ datamesh run pipeline.py               with dm.track(target="..."):         |
+---------------------------------------------------------------------------------+
                                       |
                                       v
+---------------------------------------------------------------------------------+
|                       INTERCEPTOR DE LINAJE EN BACKGROUND                       |
|  * Captura de inputs (archivos leídos, tríadas consultadas dm.sql)              |
|  * Hashing criptográfico SHA-256 de entradas, código fuente y salidas           |
|  * Métricas operacionales: duración, exit_code, timestamp UTC                   |
+---------------------------------------------------------------------------------+
                                       |
                                       v
+---------------------------------------------------------------------------------+
|                       MEMORIA OPERACIONAL LOCAL (SQLite)                        |
|  Archivo: .datamesh/lineage.db  |  .okf/provenance.db                           |
|  Tablas: executions, lineage_sources, lineage_transformations, artifacts       |
+---------------------------------------------------------------------------------+
                                       |
                   +-------------------+-------------------+
                   |                                       |
                   v                                       v
+--------------------------------------+  +---------------------------------------+
|    PUBLICACIÓN AUTOMATIZADA          |  |      CURACIÓN POR AGENTE (MCP)        |
|  * Inyección en datapackage.yaml     |  |  * datamesh_get_lineage               |
|  * Creación de nodo concepts/<id>.md |  |  * datamesh_propose_datapackage       |
|  * Publicación con dm.publish()      |  |  * Auditoría y aprobación contextual  |
+--------------------------------------+  +---------------------------------------+
```

---

## 3. Esquema Relacional de Memoria SQLite

Para garantizar cero dependencias de infraestructura externa (KISS/YAGNI), la memoria operacional se persiste en SQLite estándar:

```sql
-- Ejecuciones monitoreadas
CREATE TABLE IF NOT EXISTS executions (
    id TEXT PRIMARY KEY,
    target_triad TEXT NOT NULL,
    script_path TEXT,
    entrypoint TEXT,
    command TEXT,
    status TEXT NOT NULL, -- 'SUCCESS', 'FAILED', 'RUNNING'
    exit_code INTEGER,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    duration_ms INTEGER,
    user_identity TEXT
);

-- Fuentes de entrada (Upstream / Sources)
CREATE TABLE IF NOT EXISTS lineage_sources (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL,
    source_name TEXT NOT NULL,
    source_uri TEXT NOT NULL,
    source_hash TEXT,
    format TEXT,
    FOREIGN KEY(execution_id) REFERENCES executions(id) ON DELETE CASCADE
);

-- Pasos de transformación (Activities)
CREATE TABLE IF NOT EXISTS lineage_transformations (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL,
    step_name TEXT NOT NULL,
    tool TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    query_text TEXT,
    parameters_json TEXT,
    FOREIGN KEY(execution_id) REFERENCES executions(id) ON DELETE CASCADE
);

-- Artefactos producidos (Outputs / Derivations)
CREATE TABLE IF NOT EXISTS artifacts (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL,
    path TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    bytes_size INTEGER NOT NULL,
    format TEXT NOT NULL,
    FOREIGN KEY(execution_id) REFERENCES executions(id) ON DELETE CASCADE
);
```

---

## 4. Reglas Normativas de Ejecución y Publicación

### RULE-13 (Cálculo Automático de Huellas de Linaje)
Todo proceso ejecutado mediante `okf run` o decorado con `@track_lineage` DEBE computar el hash SHA-256 del script o bloque de código ejecutor y de los artefactos resultantes. Si el código fuente cambia, la transformación DEBE emitir un nuevo `code_hash`.

### RULE-14 (Persistencia Cero-Configuración en SQLite)
El runtime local DEBE inicializar automáticamente la base de datos `.datamesh/lineage.db` (o `.okf/provenance.db`) en el espacio de trabajo si no existe. Las operaciones de lectura y escritura no deben requerir privilegios de administrador ni servicios externos de base de datos.

### RULE-15 (Cierre del Bucle de Publicación Asistida por Agentes)
Al completarse una ejecución con estado `SUCCESS`, el sistema DEBE permitir:
1. **Generación Inmediata de Manifiestos:** Exportar el registro de linaje hacia la sección `lineage:` de `datapackage.yaml` conforme a W3C PROV-O (`sources`, `transformations`, `was_derived_from`).
2. **Exposición Agéntica Semántica:** Exponer el estado en herramientas MCP (`datamesh_get_lineage`) para que un agente autónomo o asistente humano pueda verificar la calidad antes de invocar `dm.publish()`.

---

## 5. Referencias Bibliográficas (APA 7ma / IEEE)

1. **W3C Provenance Working Group.** (2013). *PROV-O: The PROV Ontology*. W3C Recommendation. https://www.w3.org/TR/prov-o/
2. **OpenLineage Project.** (2024). *OpenLineage Specification: An Open Standard for Data Lineage Collection*. https://openlineage.io/
3. **Hipp, D. R., Kennedy, D., & Mistachkin, J.** (2020). *SQLite: An Embeddable Database as an Application File Format*. IEEE Software.
4. **Sanderson, C., Freeman, A., & Schmidt, T.** (2025). *Operational Data Contracts and Automated Lineage Synthesis*. ACM Computing Surveys.
5. **Anthropic.** (2024). *Model Context Protocol (MCP): Architecture and Tool Interoperability*.
