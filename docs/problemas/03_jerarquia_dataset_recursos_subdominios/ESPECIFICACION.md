# ESPECIFICACIÓN TÉCNICA Y PROBLEMÁTICA 03
## Jerarquía Semántica: El Dataset como Subdominio y Recursos Dependientes (Estáticos y Dinámicos)

**Estado:** Documento de Especificación Normativa (Estilo W3C Working Draft)  
**Módulo:** Modelo de Dominio, Ciclo de Vida y Servicios Dinámicos OKF  
**Repositorio:** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)  
**Cuaderno de Investigación:** [Federación de Catálogos, Recursos Anidados y Topología OKF](https://notebook.google.com/notebook/4c4d4a2b-c49c-465c-bf60-8bf2b381b3f0) (`4c4d4a2b-c49c-465c-bf60-8bf2b381b3f0`)

---

## 1. Declaración de la Problemática

En sistemas de archivos o APIs mal diseñadas, es común encontrar "archivos de datos sueltos" (ej. `datos_final_v2.csv` o un endpoint `/get-data`) que carecen de pertenencia a una entidad lógica de negocio:
1. **La Confusión entre Contenido y Distribución:** Un usuario no busca "un archivo Parquet", busca "las mediciones de calidad del aire". El archivo Parquet es solo una manifestación física (*Distribution*) de una entidad conceptual abstracta (*Dataset*).
2. **Dependencia Ontológica Estricta:** Un recurso físico o dinámico NO PUEDE existir en el vacío; depende intrínsecamente del dataset al que pertenece. Si el dataset es deprecado o cambia de licencia, todos sus recursos subordinados heredan dicho cambio.
3. **Coexistencia de Recursos Estáticos y Dinámicos:** Dentro del mismo subdominio de datos coexisten archivos históricos inmutables (estáticos) y servicios de streaming analítico en tiempo real o vistas SQL calculadas al vuelo (dinámicos).

### El Desafío Arquitectónico:
> *¿Cómo formalizar la jerarquía semántica donde el Dataset actúa como un subdominio/raíz de agregación (Aggregate Root en DDD) que gobierna el ciclo de vida, licencias y contexto, mientras que los Recursos actúan como distribuciones físicas dependientes (estáticas o dinámicas)?*

---

## 2. Estado del Arte: Domain-Driven Design (DDD) y W3C DCAT v3

La especificación articula el diseño guiado por el dominio (**Eric Evans, 2003**) con el estándar **W3C DCAT v3 (2024)**:

```
              JERARQUÍA SEMÁNTICA DEL DOMINIO OKF

                     dcat:Catalog (Dominio Raíz)
                                │
                                │ dcat:dataset
                                v
                     dcat:Dataset (Subdominio)
                     [Raíz de Agregación / DDD Aggregate]
                     * Licencia Global (CC-BY-4.0)
                     * Anclaje Semántico (SKOS / Wikidata)
                     * Dueño del Dominio (Domain Ownership)
                                │
         ┌──────────────────────┴──────────────────────┐
         │ dcat:distribution                           │ dcat:distribution / service
         v                                             v
  dcat:Distribution (Estática)                 dcat:DataService (Dinámica)
  * Archivo Parquet inmutable                  * Endpoint Flight SQL / DuckDB
  * Hash SHA-256 verificado                    * Consulta parametrizada
  * Table Schema sintáctico                    * Linaje dinámico PROV-O
```

### 2.1. El Dataset como Raíz de Agregación (DDD Aggregate Root)
- En DDD, un Aggregate Root garantiza la consistencia e invariantes de todas las entidades subordinadas.
- El `dcat:Dataset` define las reglas de acceso, el contexto temático (`skos:Concept`), los autores responsables y las políticas de retención (`retention_policy`).
- Un recurso individual (`dcat:Distribution` o `dcat:DataService`) no puede alterar la licencia ni violar la política `zero_microdata` establecida por su dataset padre.

### 2.2. Agrupaciones Temporales mediante `dcat:DatasetSeries`
Cuando un subdominio produce particiones periódicas (ej. el Censo o el Índice de Precios al Consumidor mensual), DCAT v3 introduce la clase **`dcat:DatasetSeries`**:
- Modela la serie completa (`dcat:inSeries`).
- Establece la navegación secuencial cronológica mediante `dcat:first`, `dcat:prev`, `dcat:next` y `dcat:last`.

---

## 3. Especificación de Recursos Estáticos vs. Dinámicos en OKF

En el documento de concepto del bundle, la jerarquía se formaliza declarando el tipo de oferta en el bloque `resources`:

```yaml
---
id: calidad_aire_lpz
title: "Sistema de Monitoreo de Calidad del Aire de La Paz"
type: dataset                          # El concepto actúa como Dataset (Subdominio)
domain: medio_ambiente_bolivia
license: "CC-BY-4.0"
skos_concept: "http://aims.fao.org/aos/agrovoc/c_330843"

# =============================================================================
# RECURSOS DEPENDIENTES (Estáticos y Dinámicos)
# =============================================================================
resources:
  # ---------------------------------------------------------------------------
  # 1. RECURSO ESTÁTICO: Archivo columnar inmutable para analítica histórica
  # ---------------------------------------------------------------------------
  - id: historico_2025_parquet
    title: "Consolidado Anual Histórico 2025"
    type: static_distribution          # [static_distribution | dynamic_service]
    format: "parquet"
    media_type: "application/vnd.apache.parquet"
    path: "data/calidad_aire_2025.parquet"
    data_hash_sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    contract:
      type: "frictionless_table_schema"
      path: "schemas/pm25.schema.json"

  # ---------------------------------------------------------------------------
  # 2. RECURSO DINÁMICO: Servicio analítico en vivo o cómputo atestado
  # ---------------------------------------------------------------------------
  - id: mediciones_tiempo_real
    title: "Mediciones de Últimas 24 Horas en Streaming"
    type: dynamic_service              # [static_distribution | dynamic_service]
    protocol: "flight-sql"             # [flight-sql | duckdb-query | rest-api]
    endpoint: "arrow-flight://sensores.lapaz.bo:8443/stream"
    query_template: "SELECT * FROM estaciones WHERE timestamp >= NOW() - INTERVAL '24 hours'"
    stale_after_seconds: 300           # Frescura de 5 minutos
    prov_activity: "urn:uuid:stream-pipeline-realtime"
---
```

---

## 4. Reglas Normativas de Dependencia y Direccionamiento

### 4.1. Imposibilidad de Direccionamiento Huérfano (Regla OE-HIER-01)
> **Regla OE-HIER-01:**  
> Ningún motor conforme PUEDE resolver un recurso sin especificar su dataset padre.  
> La sintaxis válida es estrictamente jerárquica:  
> `<dataset>.<recurso>@<catalogo>`  
> El intento de acceder a `.recurso@catalogo` DEBE fallar con un error `ORPHAN_RESOURCE_ACCESS_DENIED`.

### 4.2. Cascada de Herencia de Políticas
1. Si el Dataset declara `policy: zero_microdata`, **ningún recurso (ni estático ni dinámico)** puede exponer lecturas a nivel de fila cruda sin previa agregación autorizada.
2. Si el Dataset pasa a estado `status: deprecated`, todos sus recursos asociados entran en obsolescencia automática.

---

## 5. Referencias Bibliográficas (APA 7ma / IEEE)

* **Evans, E. (2003).** *Domain-Driven Design: Tackling Complexity in the Heart of Software*. Addison-Wesley.
* **World Wide Web Consortium (W3C). (2024).** *Data Catalog Vocabulary (DCAT) - Version 3*. W3C Recommendation.
* **Apache Software Foundation. (2024).** *Apache Arrow Flight SQL: A high-performance database client protocol*. Apache Arrow Documentation.
* **Dehghani, Z. (2022).** *Data Mesh: Delivering Data-Driven Value at Scale*. O'Reilly Media.
