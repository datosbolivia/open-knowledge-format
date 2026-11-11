---
type: dcat:DataService
id: ckan-reader-service
title: Servicio Lector y Cosechador de Instancias CKAN
description: Conector ejecutable y servicio de consulta para descubrir, buscar y cosechar datasets desde portales CKAN hacia paquetes OKF v0.2.
conformsTo: https://openknowledge.org/okf/v0.2
publisher:
  name: Open Knowledge Format & DataMesh Core
dimensions:
  - ckan_url
  - dataset_id
skos:
  - https://www.wikidata.org/wiki/Q1197685
contracts:
  - type: owl-cli
    path: ./owl-cli.yaml
  - type: openapi
    path: ./openapi.yaml
---

# Servicio Lector y Cosechador de Instancias CKAN

Este servicio implementa el patrón de **Conector Activo y Servicio en Catálogo** de OKF v0.2:
En lugar de depender de configuraciones externas o código ad-hoc fuera del catálogo, el catálogo expone formalmente tanto la **interfaz remota (OpenAPI)** como el **script ejecutable local (OWL-CLI)** para interactuar con cualquier portal CKAN público o privado.

## 1. Ejecución Local Monitoreada (`okf run`)

El conector puede ser invocado directamente por analistas o agentes autónomos utilizando el CLI `okf` con seguimiento de procedencia y linaje criptográfico (SHA-256):

```bash
# Búsqueda y cosecha hacia un directorio de bundles
okf run connectors/ckan/read_ckan.py --target "ckan:datos_gob_bo" -- \
  --url https://datos.gob.bo \
  --query "presupuesto" \
  --limit 5 \
  --out ./bundles/presupuesto_bo
```

O ejecutable directamente con Python:

```bash
python3 connectors/ckan/read_ckan.py \
  --url https://datos.gob.bo \
  --package-id "ejecucion-presupuestaria-2024" \
  --out ./bundles/presupuesto
```

## 2. Parámetros de Entrada (Inputs Semánticos)

| Parámetro | Flag CLI | Concepto Semántico | Requerido | Descripción |
| :--- | :--- | :--- | :--- | :--- |
| `url` | `--url` / `-u` | `wd:Q1197685` (CKAN) | **Sí** | URL base de la instancia CKAN (ej. `https://datos.gob.bo`) |
| `query` | `--query` / `-q` | `okf:search_query` | No | Término o filtro de búsqueda textual |
| `package_id` | `--package-id` / `-p` | `okf:dataset_id` | No | Identificador o slug del paquete en CKAN |
| `limit` | `--limit` / `-l` | `okf:result_limit` | No | Máximo de paquetes a cosechar (default: `10`) |
| `token` | `--token` / `-t` | `okf:auth_token` | No | Token de autorización para portales protegidos |
| `out` | `--out` / `-o` | `okf:output_directory` | No | Directorio destino para exportar los bundles OKF |

## 3. Salida Producida (Outputs Semánticos)

- **Manifiesto de Datos**: `datapackage.yaml` con especificación Frictionless y recursos tabulares detectados.
- **Documento de Concepto**: `index.md` con metadatos de procedencia, cobertura espacial/temporal y enlaces de descarga.
- **Linaje Epistémico**: Registro de origen enlazado al portal CKAN y sellado temporalmente.
