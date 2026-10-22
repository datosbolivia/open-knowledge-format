# Guía de Ejecución en Google Colab: OKF Engine & DataMesh SDK

Esta guía permite inicializar y operar el motor normativo **`okf` CLI** y el **`DataMesh SDK`** en cualquier sesión interactiva de **Google Colab** en menos de 5 segundos.

---

## 1. Instalación Rápida de 1 Celda

Copia y pega este bloque en la primera celda de tu cuaderno en Google Colab:

```python
# ==============================================================================
# Celda 1: Instalación de okf CLI (Go Core nativo) y DataMesh SDK (Python/DuckDB)
# ==============================================================================
import os

# 1. Descargar binario estático okf para Linux x86_64
!curl -sSL "https://github.com/datosbolivia/open-knowledge-format/releases/download/v0.2.1-rc.1/okf-linux-amd64" -o /usr/local/bin/okf
!chmod +x /usr/local/bin/okf

# 2. Instalar DataMesh SDK desde el release de GitHub
!pip install --quiet "https://github.com/datosbolivia/datamesh/releases/download/v0.2.1-rc.1/datamesh_sdk-0.2.1-py3-none-any.whl" duckdb pyyaml

# 3. Comprobar instalación
!okf version
import datamesh as dm
print(f"✓ DataMesh SDK v{dm.__version__} instalado correctamente con motor {dm.get_engine().name}.")
```

---

## 2. Operaciones Agénticas con `okf` CLI

El motor nativo de Go permite inicializar, curar, auditar y graficar bundles de conocimiento directamente desde bash en Colab:

```python
# Crear un nuevo bundle de conocimiento OKF v0.2
!okf init /content/mi_bundle -title "Estudio Macro Colab" -desc "Análisis experimental en Google Colab"

# Crear nodos conceptuales
!okf add /content/mi_bundle/concepts/inflacion.md -type dataset -title "Inflación Mensual" -desc "Índice de precios al consumidor"
!okf add /content/mi_bundle/concepts/tipo_cambio.md -type dataset -title "Tipo de Cambio Paralelo" -desc "Cotizaciones P2P"

# Enlazar conceptualmente (Grafo de Conocimiento)
!okf link /content/mi_bundle/concepts/inflacion.md /content/mi_bundle/concepts/tipo_cambio.md -rel "Correlaciona Con"

# Re-indexar y Auditar
!okf index /content/mi_bundle
!okf audit /content/mi_bundle

# Exportar grafo de adyacencia en JSON para LLMs
!okf graph /content/mi_bundle
```

---

## 3. Consultas Analíticas con DataMesh SDK

Ejecución de SQL analítico federado DuckDB sobre tríadas o datos locales:

```python
import datamesh as dm

# Descubrimiento de catálogos soberanos
catalogo = dm.discover()
print("Catálogos disponibles:", catalogo.get("title"))

# Consultas analíticas SQL con DuckDB streaming
resultado = dm.sql("SELECT 1 AS status, 'DataMesh en Colab' AS message")
print(resultado["rows"])
```
