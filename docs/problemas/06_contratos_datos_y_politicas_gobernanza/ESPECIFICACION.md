# ESPECIFICACIÓN TÉCNICA Y PROBLEMÁTICA 06
## Contratos de Datos, Calidad Sintáctica y Políticas de Gobernanza Cero-Microdatos

**Estado:** Documento de Especificación Normativa (Estilo W3C Working Draft)  
**Módulo:** Contratos de Datos, Validación de Calidad y Políticas Computacionales OKF  
**Repositorio:** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)  
**Cuaderno de Investigación:** [DATAMESH](https://notebook.google.com/notebook/52ebcfe9-f328-4885-834f-f1f2f4d7bc4b) (`52ebcfe9-f328-4885-834f-f1f2f4d7bc4b`)

---

## 1. Declaración de la Problemática

En las organizaciones tradicionales, la relación entre productores y consumidores de datos adolece de una desconexión crítica:
1. **Rupturas Silenciosas en Pipelines (*Silent Schema Drifts*):** Un cambio inadvertido en la base de datos origen (ej. renombrar una columna `fecha_hora` a `timestamp`, o cambiar un tipo de entero a texto) rompe los dashboards de inteligencia de negocios y los modelos de machine learning sin emitir alertas previas.
2. **Exposición Indiscriminada de Datos Sensibles:** La falta de políticas ejecutables en el punto de intercambio permite que datos personales o información confidencial se filtren hacia analistas que solo requerían agregaciones estadísticas generales.
3. **Falta de Responsabilidad Formal:** No existe un acuerdo vinculante que defina el nivel de servicio (SLA), frescura y tolerancias de valores nulos entre quien produce los datos y quien los consume.

### El Desafío Arquitectónico:
> *¿Cómo implementar Contratos de Datos (Data Contracts) formales como código declarativo y políticas de gobernanza computacional federada (Zero-Microdata), asegurando que ningún dato corrupto ingrese a la red y protegiendo la privacidad de los microdatos?*

---

## 2. Estado del Arte: Contratos de Datos y Gobernanza Computacional

```
                     ARQUITECTURA DE CONTRATOS Y POLÍTICAS

  [Productor de Datos]
           │
           │ 1. Emite Producto con Contrato Estricto
           v
  [Data Contract (Sanderson et al., 2025)]
  * Esquema Tipado (Frictionless Table Schema / JSON Schema)
  * Invariantes de Calidad (NULL ratios, rangos numéricos)
  * SLAs y Reglas de Frescura (stale_after)
           │
           │ 2. Evaluación de Gobernanza Computacional
           v
  [Motor de Políticas como Código (OPA / Rego)]
  * Política: zero_microdata
  * ¿Consulta pide filas individuales (SELECT *)? -> RECHAZADO (403)
  * ¿Consulta pide agregaciones (AVG, COUNT, GROUP BY)? -> PERMITIDO
           │
           │ 3. Verificación de Integridad Criptográfica
           v
  [Confianza Criptográfica (SHA-256 + Firma SSH)]
```

### 2.1. Contratos de Datos como Código (Sanderson, Freeman y Schmidt, 2025)
- El contrato de datos es un artefacto declarativo versionado en Git que define formalmente la interfaz de datos entre dominios.
- Los cambios de esquema se gestionan mediante versionado semántico; un cambio incompatible (*breaking change*) incrementa la versión mayor del contrato, alertando a los consumidores antes del despliegue.

### 2.2. Gobernanza Computacional Federada y Open Policy Agent (OPA)
- Como postula **Zhamak Dehghani (2022)**, la gobernanza en Data Mesh no es un comité burocrático, sino **código automatizado** embebido en la plataforma.
- Mediante políticas declarativas escritas en el lenguaje **Rego (OPA)**, los nodos evalúan automáticamente el cumplimiento de normativas de privacidad antes de despachar bloques de datos.

---

## 3. Especificación de Contratos y Políticas en OKF

### 3.1. Declaración en el Frontmatter del Concepto

```yaml
---
id: registros_salud_publica
title: "Estadísticas Hospitalarias y Diagnósticos Clínicos"
type: dataset
domain: salud_bolivia

# =============================================================================
# CONTRATO SINTÁCTICO DE DATOS (Data Contract)
# =============================================================================
contract_type: frictionless_table_schema
contract_path: "schemas/salud_schema.json"
schema_hash: "a1b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef0"

# Invariantes de calidad sintáctica obligatorios
quality_rules:
  max_null_percentage:
    fecha_ingreso: 0.0          # Clave primaria obligatoria
    diagnostico_cie10: 0.01     # Máximo 1% de nulos
  value_ranges:
    edad_paciente:
      min: 0
      max: 125

# =============================================================================
# POLÍTICA DE GOBERNANZA: ZERO-MICRODATA (Protección de Privacidad)
# =============================================================================
policy: zero_microdata          # [allow_all | zero_microdata | deny]
retention_policy: "5_years"
confidentiality_tier: "restricted_academic"
---
```

---

## 4. Reglas Normativas de Evaluación y Enforcement

### 4.1. Invariante OE3: Aplicación de Zero-Microdatos (Zero-Microdata Enforcement)
> **Invariante OE3 (Norma OKF):**  
> Si un concepto declara `policy: zero_microdata`, los motores consumidores y nodos pares DEBEN rechazar escaneos no autorizados de microdatos a nivel de fila (`SELECT *`). El motor solo permitirá consultas analíticas de agregación (`COUNT`, `AVG`, `SUM`, `MIN`, `MAX`, `GROUP BY`) donde el tamaño del grupo sea superior a un umbral de anonimato $k$-anonymity ($k \ge 5$).

### 4.2. Invariante PE5: Integridad Criptográfica de Linaje
> **Invariante PE5 (Norma OKF):**  
> Todo activo físico o script de transformación referenciado en `lineage.transformations[].hash` DEBE seguir el formato `sha256:<hex64>` y coincidir exactamente con el digest computado del archivo. La detección de una discrepancia de hash bloquea inmediatamente la carga del producto de datos.

---

## 5. Referencias Bibliográficas (APA 7ma / IEEE)

* **Sanderson, C., Freeman, M., & Schmidt, B. E. (2025).** *Data Contracts: Developing Production-Grade Pipelines at Scale*. O'Reilly Media.
* **Dehghani, Z. (2022).** *Data Mesh: Delivering Data-Driven Value at Scale*. O'Reilly Media.
* **Open Policy Agent (OPA). (2024).** *Policy-based control for cloud native environments using Rego*. Cloud Native Computing Foundation (CNCF).
* **Open Knowledge Foundation. (2024).** *Table Schema: A simple specification for tabular data*. Frictionless Standards.
