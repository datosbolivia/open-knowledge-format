# ESPECIFICACIÓN DETALLADA DEL CONTENEDOR Y CLAVES ESPECIALES
## La Arquitectura del Bundle OKF/ODKF v0.2: Tipos, Claves, Tríadas e Invariantes

**Repositorio Normativo:** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)  
**Documento Fuente de Verdad:** [`SPEC.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md)  
**Esquema de Validación:** [`schemas/okf-concept.schema.json`](file:///home/andreschirinos/Proyectos/open-knowledge-format/schemas/okf-concept.schema.json)

---

## 1. Anatomía Estructural del Contenedor (Bundle)

Un contenedor de Conocimiento Abierto (*Open Knowledge Bundle*) es un directorio estructurado que combina archivos de configuración declarativa, esquemas sintácticos y documentos de concepto legibles tanto por personas como por sistemas automáticos:

```
mi_dominio_bundle/
├── metadata.yaml               # [REQUERIDO] Identidad global del dominio y políticas
├── datapackage.json            # [OPCIONAL / CONTRATO] Contrato Frictionless Table Schema
├── context.jsonld              # [OPCIONAL / SEMÁNTICA] Mapeo semántico W3C DCAT/PROV/SKOS
├── index.md                    # [RECOMENDADO] Índice navegable de conceptos
├── log.md                      # [RECOMENDADO] Registro cronológico de cambios del bundle
├── concepts/                   # [REQUERIDO] Directorio de documentos de conceptos
│   ├── calidad_aire_pm25.md    # Documento de concepto con YAML frontmatter
│   └── transformacion_etl.md   # Documento de concepto de cómputo atestado
├── schemas/                    # Contratos sintácticos detallados (JSON Schema, SQL)
└── data/                       # Distribuciones físicas locales (Parquet, CSV opcionales)
```

### 1.1. Reglas de Nombres Reservados en el Bundle
- `metadata.yaml`: Manifiesto canónico del bundle.
- `datapackage.json`: Especificación de recursos tabulares.
- `context.jsonld`: Espacios de nombres semánticos W3C.
- `index.md`: Directorio humano y de agentes.
- `log.md`: Bitácora inmutable de versiones.

---

## 2. Documentos de Concepto y Claves Especiales en el Frontmatter

El documento de concepto (`concepts/*.md`) es la unidad atómica de transferencia de conocimiento. Consta de dos secciones indivisibles:
1. **YAML Frontmatter (delimitado por `---`):** Metadatos rigurosamente tipados para máquinas.
2. **Cuerpo en Markdown:** Contexto narrativo, explicaciones de negocio y lineamientos interpretativos para humanos y modelos de lenguaje.

### 2.1. Catálogo Completo de Familias de Claves Especiales

```yaml
---
# =============================================================================
# 1. IDENTIDAD Y TIPO DEL CONCEPTO
# =============================================================================
id: calidad_aire_la_paz_pm25
title: "Monitoreo Horario de Material Particulado PM2.5 en La Paz"
type: dataset                          # [concept | dataset | service | computation]
domain: medio_ambiente_bolivia
version: "1.0.0"

# =============================================================================
# 2. ANCLAJES SEMÁNTICOS (W3C Linked Data)
# =============================================================================
dcat_dataset: "http://datos.gob.bo/dataset/calidad-aire-lpz"
skos_concept: "http://aims.fao.org/aos/agrovoc/c_330843"     # PM2.5 en AGROVOC
wikidata_id: "Q2097941"                                      # Partículas en suspensión
same_as:
  - "https://es.wikipedia.org/wiki/Contaminaci%C3%B3n_del_aire"

# =============================================================================
# 3. CONTRATOS SINTÁCTICOS DE DATOS (Índice de Estándares)
# =============================================================================
contract_type: frictionless_table_schema   # [frictionless_table_schema | json_schema | arrow_schema | sql_ddl]
contract_path: "schemas/calidad_aire.schema.json"
schema_hash: "a1b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef0"

# =============================================================================
# 4. PROCEDENCIA, LINAJE Y TRANSFORMACIONES (W3C PROV-O y CWL)
# =============================================================================
sources:
  - title: "Red de Monitoreo de Calidad del Aire (Red MoniCA)"
    url: "https://lapaz.bo/medioambiente/calidad-aire"
    accessed_at: "2026-10-15T08:00:00Z"
prov_activity: "urn:uuid:f47ac10b-58cc-4372-a567-0e02b2c3d479"
prov_generated_by: "clean_monica_sensor_data.py"
transformation_hash: "9f8e7d6c5b4a3210fedcba9876543210fedcba9876543210fedcba9876543210"

# =============================================================================
# 5. CONFIANZA CRIPTOGRÁFICA Y GOBERNANZA
# =============================================================================
data_hash_sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
signed_by: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG5... andres@udabol.edu.bo"
trust_tier: tier_1_verified            # [tier_1_verified | tier_2_community | tier_3_raw]
generated: false                       # true si fue generado sintéticamente por IA
verified: true                         # true si pasó validación de contratos formal

# =============================================================================
# 6. CICLO DE VIDA Y VIGENCIA
# =============================================================================
status: active                         # [draft | active | deprecated]
created_at: "2026-10-15T10:00:00Z"
updated_at: "2026-10-21T12:00:00Z"
stale_after: "2026-11-15T00:00:00Z"    # Invalida cache tras esta fecha
license: "CC-BY-4.0"
---

# Descripción del Recurso
El presente dataset contiene mediciones horarias de partículas menores a 2.5 micrómetros...
