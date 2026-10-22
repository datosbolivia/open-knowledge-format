# Guía de Contribución y Desarrollo (CONTRIBUTING.md)

Gracias por tu interés en contribuir al estándar **Open Knowledge Format (OKF / ODKF v0.2)** y su motor de referencia en Go (`src/`).

---

## 1. Principios de Contribución

1. **Specification-Driven Design (SDD):** Toda modificación a la lógica del motor debe estar respaldada por la especificación [`SPEC.md`](SPEC.md) o [`specificacion_knowledge.md`](docs/SPECIFICATION_DRIVEN_DESIGN.md).
2. **Domain-Driven Design (DDD) & Hexagonal Architecture:**
   - La lógica de dominio reside en `src/domain/` y es agnóstica de infraestructura y librerías externas.
   - Los adaptadores I/O residen en `src/adapters/` y se comunican a través de los puertos en `src/ports/`.
   - La CLI pública reside en `src/cmd/okf/`.
3. **Cero Dependencias Ocultas:** Todo binario de distribución debe compilarse de manera reproducible sin dependencias del sistema operativo (`CGO_ENABLED=0`).

---

## 2. Configuración del Entorno de Desarrollo

### Requisitos
- **Go:** 1.22 o superior.
- **Git:** 2.30+ con soporte de firmado de tags.

### Comandos Clave de Desarrollo
```bash
# Ejecutar todas las pruebas unitarias y de dominio con detector de condiciones de carrera
cd src
go test -v -race ./...

# Compilar localmente la CLI okf
go build -o ../bin/okf ./cmd/okf

# Probar la CLI local
../bin/okf version
../bin/okf validate ../testdata/bundles/calidad_aire.md
```

---

## 3. Políticas de Versionado y Pre-lanzamientos

Seguimos **Semantic Versioning 2.0 (SemVer)**:
- **Mayor (`v1.0.0`):** Cambios incompatibles en la especificación, contratos o CLI.
- **Menor (`v0.2.0`):** Nuevas funcionalidades, resolutores o validadores retrocompatibles.
- **Parche (`v0.2.1`):** Correcciones de errores y optimizaciones retrocompatibles.
- **Versiones de prueba (`-alpha.N`, `-beta.N`, `-rc.N`):** Pre-releases automáticos en GitHub Releases al subir tags con guion.

Consulta la guía detallada en [`docs/VERSIONADO_Y_LANZAMIENTOS.md`](docs/VERSIONADO_Y_LANZAMIENTOS.md).
