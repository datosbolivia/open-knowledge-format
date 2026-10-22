# ESPECIFICACIÓN TÉCNICA Y PROBLEMÁTICA 02
## Sindicación, Co-referencia y Múltiples Catálogos Referenciando el Mismo Dataset (Many-to-Many)

**Estado:** Documento de Especificación Normativa (Estilo W3C Working Draft)  
**Módulo:** Sindicación de Datos y Desduplicación Semántica OKF  
**Repositorio:** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)  
**Cuaderno de Investigación:** [Federación de Catálogos, Recursos Anidados y Topología OKF](https://notebook.google.com/notebook/4c4d4a2b-c49c-465c-bf60-8bf2b381b3f0) (`4c4d4a2b-c49c-465c-bf60-8bf2b381b3f0`)

---

## 1. Declaración de la Problemática

En un ecosistema abierto de datos, **un dataset nunca debe quedar atrapado en un único catálogo**:
1. **Sindicación Multi-Institucional:** Un conjunto de datos de *Calidad del Aire en La Paz* puede ser generado y publicado oficialmente por la Alcaldía de La Paz, pero simultáneamente es sindicado por el catálogo del Ministerio de Medio Ambiente, el catálogo de la UDABOL y el portal nacional de ciencia y tecnología.
2. **El Problema de la Duplicación vs. Identidad:** Si cada catálogo crea su propia copia del dataset con un identificador local arbitrario, se produce fragmentación: los consumidores no saben si el dataset en la UDABOL es el mismo que el de la Alcaldía, si tiene las mismas filas o si está desactualizado.
3. **Pérdida de Autoría y Metadatos Administrativos:** Es necesario distinguir entre los metadatos propios del dataset (fecha de creación original, autor, licencia) y los metadatos de inclusión en un catálogo (fecha en la que el catálogo agregó el registro, prioridad o notas locales).

### El Desafío Arquitectónico:
> *¿Cómo permitir que múltiples catálogos referencien y sindiquen exactamente el mismo dataset o recurso, preservando la identidad criptográfica única del productor original, desduplicando el contenido en red y separando los metadatos conceptuales del registro administrativo local?*

---

## 2. Estado del Arte y Solución en W3C DCAT v3

La especificación **W3C DCAT v3** resuelve la relación muchos-a-muchos mediante un desacoplamiento de clases:

```
                  DESACOPLAMIENTO SEMÁNTICO EN W3C DCAT v3

      dcat:Catalog                                    dcat:Catalog
   [Catálogo UDABOL]                             [Catálogo Alcaldía LPZ]
          │                                                │
          │ dcat:record                                    │ dcat:record
          v                                                v
  dcat:CatalogRecord                              dcat:CatalogRecord
  (Registro en UDABOL)                            (Registro en Alcaldía)
  * dcterms:issued: 2026-10-21                    * dcterms:issued: 2026-05-10
          │                                                │
          │ foaf:primaryTopic                              │ foaf:primaryTopic
          └───────────────────────┬────────────────────────┘
                                  │
                                  v
                            dcat:Dataset
                   [Calidad del Aire PM2.5 La Paz]
                   * IRI Canónica Global
                   * Hash Inmutable: sha256:e3b0c44...
                   * Autor Original: Red MoniCA LPZ
```

### 2.1. El Patrón `dcat:CatalogRecord`
- **`dcat:Dataset`**: Entidad conceptual soberana. Su fecha de emisión (`dcterms:issued`) y autor no cambian cuando un tercero lo cosecha.
- **`dcat:CatalogRecord`**: Representa la inscripción del dataset *dentro de un catálogo específico*. Contiene la fecha de cosecha o sindicación local sin tocar el dataset original.
- **`foaf:primaryTopic`**: Predicado que une la entrada de catálogo con el dataset conceptual.

### 2.2. Identidad y Co-referencia con `owl:sameAs` y `adms:identifier`
Cuando dos catálogos utilizan identificadores locales distintos para el mismo recurso físico, DCAT v3 estipula el uso de:
- **`owl:sameAs`**: Declaración ontológica de equivalencia estricta.
- **`adms:identifier`**: Soporte para identificadores globales persistentes (DOIs, handles, hashes).

---

## 3. Especificación en el Formato de Conocimiento Abierto (OKF)

En OKF, un dataset o concepto posee un **identificador inmutable por contenido (CID)** y una firma criptográfica. Un catálogo puede sindicar un dataset remoto simplemente declarando un registro de sindicación en su bundle:

### 3.1. Declaración de Sindicación en el Frontmatter del Concepto

```yaml
---
id: calidad_aire_lpz_syndicated
title: "Monitoreo de Calidad del Aire - La Paz (Sindicado de Alcaldía)"
type: dataset
domain: udabol_medio_ambiente

# =============================================================================
# CO-REFERENCIA Y SINDICACIÓN MANY-TO-MANY
# =============================================================================
syndicated_from:
  catalog: "lapaz_alcaldia"
  dataset_id: "calidad_aire_pm25"
  resource_id: "pm25_horario"
  canonical_triad: "lapaz_alcaldia:calidad_aire_pm25:pm25_horario"
  canonical_cid: "bafybeie3b0c44298fc1c14..."
  original_signed_by: "ssh-ed25519 AAAAC3NzaC... alcaldia@lapaz.bo"

# Metadatos del registro local en el catálogo UDABOL (CatalogRecord)
catalog_record:
  indexed_at: "2026-10-21T16:30:00Z"
  harvest_strategy: "remote_pointer"    # [remote_pointer | local_cached]
  local_notes: "Utilizado para tesis de grado en Ingeniería de Sistemas"

# Equivalencia ontológica
same_as:
  - "http://datos.lapaz.bo/dataset/calidad-aire-pm25"
  - "urn:uuid:f47ac10b-58cc-4372-a567-0e02b2c3d479"

# Huella digital SHA-256 obligatoria (PE5): DEBE coincidir con el original
data_hash_sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
---
```

---

## 4. Resolución Canónica y Desduplicación en Redes P2P

### 4.1. Desduplicación por Contenido (Zero-Redundancy)
Cuando un motor o cliente (como `datamesh-sdk`) consulta el recurso a través de cualquiera de los catálogos:
- `okf-dig calidad_aire_pm25.pm25_horario@lapaz_alcaldia`
- `okf-dig calidad_aire_lpz_syndicated.pm25_horario@udabol_medio_ambiente`

Ambas consultas resuelven al **mismo hash SHA-256** (`e3b0c44...`). El motor de almacenamiento local en caché (`~/.datamesh/cache/blobs/`) almacena el archivo físico una sola vez, reutilizándolo para ambos catálogos y evitando el consumo redundante de ancho de banda y disco.

### 4.2. Validación de Invariante de Sindicación (Norma PE-SYN-01)
> **Invariante PE-SYN-01:**  
> Un catálogo sindicador NO PUEDE modificar el `data_hash_sha256` ni la firma `original_signed_by` del recurso sindicado. Si el catálogo sindicador transforma, limpia o añade columnas a los datos, DEBE emitir un nuevo dataset propio con su propio linaje W3C PROV-O (`prov:wasDerivedFrom`), asumiendo la autoría criptográfica del nuevo producto derivado.

---

## 5. Referencias Bibliográficas (APA 7ma / IEEE)

* **World Wide Web Consortium (W3C). (2024).** *Data Catalog Vocabulary (DCAT) - Version 3*. W3C Recommendation.
* **World Wide Web Consortium (W3C). (2017).** *Data on the Web Best Practices (DWBP)*. W3C Recommendation. https://www.w3.org/TR/dwbp/
* **Berners-Lee, T. (2006).** *Linked Data - Design Issues*. W3C. https://www.w3.org/DesignIssues/LinkedData.html
* **Open Knowledge Foundation. (2024).** *Frictionless Standards: Cross-referencing Data Resources*.
