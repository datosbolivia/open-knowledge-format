# ESPECIFICACIÓN TÉCNICA Y PROBLEMÁTICA 04
## Recursos Compuestos y Archivos Anidados (Archives ZIP, TAR, Multipart y Fragment Identifiers)

**Estado:** Documento de Especificación Normativa (Estilo W3C Working Draft)  
**Módulo:** Gestión de Archivos Compuestos, Compresión y Lectura Selectiva OKF  
**Repositorio:** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)  
**Cuaderno de Investigación:** [Federación de Catálogos, Recursos Anidados y Topología OKF](https://notebook.google.com/notebook/4c4d4a2b-c49c-465c-bf60-8bf2b381b3f0) (`4c4d4a2b-c49c-465c-bf60-8bf2b381b3f0`)

---

## 1. Declaración de la Problemática

En la práctica de la ciencia de datos y la administración pública (como el INE, SENAMHI o investigaciones geológicas de la UMSA/UDABOL), una gran proporción de los datos se distribuye como **archivos contenedores o empaquetados** (archivos `.zip`, `.tar.gz`, carpetas particionadas o bases de datos GeoPackage/HDF5):
1. **La Ineficiencia de la Descarga y Descompresión Total:** Si un investigador necesita una única tabla de 5 MB dentro de un archivo ZIP de 10 GB, las herramientas tradicionales exigen descargar los 10 GB enteros y descomprimirlos en disco, saturando el ancho de banda y la memoria.
2. **Ambigüedad en el Direccionamiento:** Un archivo ZIP no es una tabla; es un contenedor que puede incluir 15 archivos CSV, 3 GeoJSONs y un manual PDF. No existía una sintaxis canónica uniforme para referenciar directamente a un archivo interno sin perder la relación con el contenedor original.
3. **Pérdida de Verificación de Integridad:** Verificar solo el hash del ZIP garantiza que el archivo comprimido no se alteró, pero no permite validar individualmente si uno de los archivos internos fue corrompido o manipulado.

### El Desafío Arquitectónico:
> *¿Cómo especificar y direccionar recursos compuestos (ZIPs, TARs, multiparts), permitiendo la lectura selectiva parcial de subarchivos sin descompresión masiva en disco, con verificación de integridad granular (Merkle Tree / SPDX Checksums) y esquemas sintácticos independientes para cada subarchivo interno?*

---

## 2. Estado del Arte y Fundamentación en Estándares

La especificación unifica las directrices de tres estándares internacionales:

```
                    ARQUITECTURA DE RECURSOS COMPUESTOS

      dcat:Distribution (Recurso Contenedor / ZIP)
      * dcat:packageFormat: application/zip
      * dcat:compressFormat: application/gzip
      * spdx:checksum: sha256:7f8a9b... (Hash del ZIP completo)
                           │
      ┌────────────────────┴────────────────────┐
      │ Fragment Identifier (#)                 │ Fragment Identifier (#)
      v                                         v
  Sub-recurso Interno A                     Sub-recurso Interno B
  [estaciones_senamhi.csv]                 [mediciones_historicas.parquet]
  * Hash individual SHA-256                * Hash individual SHA-256
  * Table Schema A                         * Table Schema B
  * Offset HTTP Range: Bytes 1024-5048     * Offset HTTP Range: Bytes 5049-98000
```

### 2.1. W3C DCAT v3: `dcat:packageFormat` y `spdx:checksum`
DCAT v3 introduce explícitamente:
- **`dcat:packageFormat`**: Declara el formato del contenedor (ej. `application/zip`, `application/x-tar` o Frictionless Data Package).
- **`dcat:compressFormat`**: Declara el algoritmo de compresión aplicado (ej. `gzip`, `bzip2`, `zstd`).
- **`spdx:checksum`**: Utiliza la ontología SPDX para asociar digests criptográficos tanto al archivo contenedor como a sus partes constituyentes.

### 2.2. Frictionless Data: Multipart Resources y Data Packages
La especificación Frictionless Data Package v2 permite que un solo descriptor `datapackage.json` declare múltiples `resources`, cada uno con su propio `path` relativo dentro del paquete y su propio `Table Schema`.

### 2.3. Estándar de Fragment Identifiers (IETF RFC 3986)
El estándar de URIs de la IETF define que el símbolo numeral (`#`) delimita un identificador de fragmento dentro de un recurso principal sin requerir un nuevo protocolo de red.

---

## 3. Especificación en el Formato de Conocimiento Abierto (OKF)

### 3.1. Sintaxis de Direccionamiento de Subarchivos Anidados
Para acceder a un archivo específico dentro de un contenedor comprimido:

```
<dataset>.<recurso_contenedor>#<ruta_interna>@<catalogo>
```

#### Ejemplos Canónicos:
- **Acceso a un CSV dentro de un ZIP:**
  `censo_2024.cartografia_nacional#zonas_censales/la_paz.csv@ine_bolivia`
- **Consulta SQL con DuckDB en el SDK:**
  ```sql
  SELECT * FROM "ine_bolivia:censo_2024:cartografia_nacional#zonas_censales/la_paz.csv" WHERE poblacion > 50000;
  ```

---

## 4. Declaración en el Manifiesto del Concepto

En el frontmatter del concepto, el recurso contenedor declara su estructura interna en el arreglo `contents`:

```yaml
---
id: cartografia_censo_2024
title: "Cartografía Digital y Manzanas Censales 2024"
type: dataset
domain: demografia_bolivia

resources:
  - id: cartografia_nacional_zip
    title: "Archivo Consolidado Nacional de Cartografía"
    type: static_distribution
    format: "zip"
    package_format: "application/zip"
    path: "data/cartografia_2024.zip"
    data_hash_sha256: "7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a"

    # =========================================================================
    # CATÁLOGO DE SUB-RECURSOS ANIDADOS DENTRO DEL CONTENEDOR
    # =========================================================================
    contents:
      - subpath: "zonas_censales/la_paz.csv"
        media_type: "text/csv"
        hash_sha256: "a1b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef0"
        contract:
          type: "frictionless_table_schema"
          path: "schemas/la_paz_manzanas.schema.json"

      - subpath: "zonas_censales/santa_cruz.parquet"
        media_type: "application/vnd.apache.parquet"
        hash_sha256: "b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3"
        contract:
          type: "arrow_schema"
---
```

---

## 5. Mecanismo de Lectura Parcial Cero-Copia (Zero-Copy Range Fetch)

Los motores conformes (como `datamesh-sdk`) DEBEN implementar la lectura selectiva sin extracción total:

1. **Lectura del Directorio Central del ZIP (*End of Central Directory*):**
   - El motor emite una petición HTTP con cabecera `Range: bytes=-65536` para leer los últimos 64 KB del archivo remoto.
   - Analiza la tabla de contenido interna del ZIP sin descargar el resto de los datos.
2. **Extracción Directa por Offset:**
   - Una vez localizado el offset exacto del subarchivo solicitado (ej. `zonas_censales/la_paz.csv`), emite una petición de rango acotada (`Range: bytes=1048576-2097152`).
   - El flujo comprimido se descomprime en memoria al vuelo (streaming) directamente hacia el buffer de DuckDB o Apache Arrow, logrando **cero copias en disco y latencia de red milimétrica**.

---

## 6. Referencias Bibliográficas (APA 7ma / IEEE)

* **Berners-Lee, T., Fielding, R., & Masinter, L. (2005).** *Uniform Resource Identifier (URI): Generic Syntax* (RFC 3986). Internet Engineering Task Force (IETF).
* **World Wide Web Consortium (W3C). (2024).** *Data Catalog Vocabulary (DCAT) - Version 3*. W3C Recommendation.
* **PKWARE Inc. (2020).** *ZIP File Format Specification (Version 6.3.9)*. PKWARE Application Note.
* **Open Knowledge Foundation. (2024).** *Frictionless Data Package: Multi-resource Specifications*.
