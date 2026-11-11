# Conectores y Servicios de Integración OKF

Este directorio contiene los conectores y servicios ejecutables para la ingesta, cosecha y sincronización bidireccional entre el estándar **Open Knowledge Format (OKF v0.2)** y plataformas externas de catalogación y almacenamiento de datos.

## Conectores Disponibles

| Conector / Servicio | Tipo | Protocolos / Contratos | Descripción |
| :--- | :--- | :--- | :--- |
| [`ckan`](file:///home/andreschirinos/Proyectos/open-knowledge-format/connectors/ckan/index.md) | `dcat:DataService` | `openapi` (v3.1), `owl-cli` (`read_ckan.py`) | Lector y cosechador de instancias CKAN. Convierte paquetes de Action API v3 en bundles OKF (`index.md` + `datapackage.yaml`). |
| [`gcp-knowledge-catalog`](file:///home/andreschirinos/Proyectos/open-knowledge-format/connectors/gcp-knowledge-catalog.md) | Guía & Tooling | Google Cloud Dataplex / `kcmd` | Publicación e ingesta bidireccional entre bundles OKF y Google Cloud Knowledge Catalog EntryGroups. |

---

## 1. Conector CKAN (`connectors/ckan/`)

El conector CKAN está modelado como un nodo ciudadano de primera clase dentro del grafo de conocimiento OKF bajo la especificación **SPEC.md §14.3**:

- **Declaración Semántica:** [`connectors/ckan/index.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/connectors/ckan/index.md)
- **Contrato CLI Local:** [`connectors/ckan/owl-cli.yaml`](file:///home/andreschirinos/Proyectos/open-knowledge-format/connectors/ckan/owl-cli.yaml)
- **Contrato OpenAPI Remoto:** [`connectors/ckan/openapi.yaml`](file:///home/andreschirinos/Proyectos/open-knowledge-format/connectors/ckan/openapi.yaml)
- **Script Ejecutor:** [`connectors/ckan/read_ckan.py`](file:///home/andreschirinos/Proyectos/open-knowledge-format/connectors/ckan/read_ckan.py)

### Ejecución Directa con `okf run`

Para rastrear la ejecución con linaje criptográfico (SHA-256) en `log.md`:

```bash
# Cosechar datasets de un portal CKAN público
./bin/okf run python3 connectors/ckan/read_ckan.py --url https://datos.gob.bo --limit 5 --out ./bundles/cosecha_ckan/

# Cosechar un dataset específico por slug o ID
./bin/okf run python3 connectors/ckan/read_ckan.py --url https://datos.gob.bo --package-id presupuesto-general-2024 --out ./bundles/
```

### Inspección Sintetizada

El motor de metadatos unificado de OKF sintetiza los contratos OpenAPI y CLI en una vista unificada para agentes y humanos:

```bash
./bin/okf inspect connectors/ckan/index.md
```
