# Guía de Versionado, Testing y Ciclo de Vida de Lanzamientos

Este documento establece las políticas formales de Semantic Versioning (SemVer 2.0), diseño orientado a especificaciones (Specification-Driven Design), pruebas automatizadas y ciclo de pre-lanzamientos para el **Open Knowledge Format (OKF / ODKF v0.2)** y su motor en Go (`src/`).

---

## 1. Políticas de Versionado Semántico (SemVer 2.0)

El formato del número de versión sigue estrictamente el estándar:
`MAJOR.MINOR.PATCH[-PRERELEASE]` (ej. `v0.2.0`, `v0.2.0-rc.1`, `v1.0.0`)

### 1.1 Versión Principal (MAJOR - `X.0.0`)
Se incrementa cuando se introducen **cambios incompatibles con versiones anteriores (Breaking Changes)**:
- **Especificación OKF:** Modificación o eliminación de campos requeridos en el frontmatter YAML, alteración de la estructura del bloque normativo de conceptos o cambio en los algoritmos criptográficos obligatorios (ej. transición de SHA-256 a SHA-3 sin retrocompatibilidad).
- **Go Engine (`src/`):** Alteración de interfaces públicas de dominio (`Catalog`, `DataProduct`, `ManifestParser`, `CryptoEngine`), cambios de firma en casos de uso o eliminación de comandos de la CLI `okf`.
- **Procedimiento:** Requiere un ADR formal, discusión comunitaria previa, periodo de advertencia y guía de migración.

### 1.2 Versión Secundaria (MINOR - `x.Y.0`)
Se incrementa cuando se añade **nueva funcionalidad de forma retrocompatible**:
- **Especificación OKF:** Adición de campos opcionales en metadatos, nuevos tipos de conceptos o perfiles de aplicación semántica.
- **Go Engine (`src/`):** Soporte para nuevos resolutores de almacenamiento (ej. MEFP, S3, IPFS), nuevos formatos de exportación de conocimiento (SKOS, Frictionless, DCAT-AP) o nuevos subcomandos en la CLI `okf`.
- **Procedimiento:** Requiere pruebas de regresión, actualización del catálogo de casos de uso y verificación de compatibilidad retroactiva.

### 1.3 Versión de Parche (PATCH - `x.y.Z`)
Se incrementa cuando se realizan **correcciones de errores (Bug Fixes) o mejoras de rendimiento retrocompatibles**:
- Correcciones de errores de parsing o serialización de manifiestos.
- Parches de seguridad (ej. prevención de path traversal, validación estricta de punteros nulos).
- Optimizaciones de consumo de memoria y CPU.
- **Procedimiento:** Todo parche debe incluir una prueba automatizada que reproduzca el bug original y valide su resolución definitiva.

---

## 2. Ciclo de Pre-lanzamientos y Versiones de Prueba

Para garantizar la estabilidad antes de una entrega oficial a producción, se utilizan sufijos de pre-lanzamiento:

| Sufijo | Fase | Estado de la API / Especificación | Audiencia Objetivo |
| :--- | :--- | :--- | :--- |
| `-alpha.N` | Alfa | En desarrollo activo; APIs y esquemas sujetos a cambios rápidos. | Desarrolladores del Core y colaboradores internos. |
| `-beta.N` | Beta | **Feature Freeze**: Funcionalidades completas; pruebas de integración con datos y agentes reales. | Evaluadores comunitarios, investigadores y sistemas piloto. |
| `-rc.N` | Release Candidate | **Code Freeze**: Idéntico a producción; sólo se admiten parches a bugs bloqueantes (*showstoppers*). | Validación final y preparación de despliegue. |

### Reglas Inquebrantables de Pre-lanzamiento
1. **Inmutabilidad de Tags:** Jamás sobrescribir o forzar un tag (`git tag -f`). Si un `-rc.1` tiene un error, se genera `-rc.2`.
2. **Promoción a Estable:** Cuando un Release Candidate transcurre el período de verificación (mínimo 48 horas sin incidencias críticas), el **mismo commit** exacto se etiqueta con la versión final (`vX.Y.Z`).

---

## 3. Protocolo de Pruebas Automatizadas (Testing)

Antes de etiquetar cualquier versión (de prueba o estable), la suite debe aprobar al 100%:

```bash
# 1. Pruebas unitarias y detección de carreras en Go
cd src
go test -v -race ./...

# 2. Verificación de compilación del CLI okf
go build -v -o /dev/null ./cmd/okf

# 3. Validación de manifiestos y bundles de prueba
go run ./cmd/okf validate ../testdata/bundles/air_quality_bundle/
```

---

## 4. Publicación Automatizada en GitHub Releases

El flujo de trabajo [`.github/workflows/release.yml`](.github/workflows/release.yml) automatiza la entrega:

1. **Creación del tag:**
   ```bash
   git tag -a v0.2.0-rc.1 -m "OKF v0.2.0 Release Candidate 1"
   git push origin v0.2.0-rc.1
   ```
2. **Construcción multi-plataforma:** CI compila binarios estáticos (`CGO_ENABLED=0`) para `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` y `windows/amd64`.
3. **Integridad:** CI calcula hashes SHA-256 (`SHA256SUMS.txt`).
4. **Publicación:** Los tags con guion (`-`) se marcan automáticamente como **Pre-release** en GitHub, protegiendo a los usuarios de producción.
