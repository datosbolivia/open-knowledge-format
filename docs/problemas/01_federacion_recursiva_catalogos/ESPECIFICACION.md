# ESPECIFICACIÓN TÉCNICA Y PROBLEMÁTICA 01
## Federación Recursiva de Catálogos e Inter-referenciación (Catalog-of-Catalogs)

**Estado:** Documento de Especificación Normativa (Estilo W3C Working Draft)  
**Módulo:** Topología de Catálogos y Redes Federadas OKF  
**Repositorio:** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)  
**Cuaderno de Investigación:** [Federación de Catálogos, Recursos Anidados y Topología OKF](https://notebook.google.com/notebook/4c4d4a2b-c49c-465c-bf60-8bf2b381b3f0) (`4c4d4a2b-c49c-465c-bf60-8bf2b381b3f0`)

---

## 1. Declaración de la Problemática

En un ecosistema distribuido y soberano, los catálogos de datos no existen como islas aisladas ni como un único megarepositorio central. Por el contrario:
1. **Jerarquías Institucionales:** Un catálogo nacional (ej. `datos.gob.bo`) necesita referenciar e indexar catálogos ministeriales (ej. `salud.gob.bo`, `economia.gob.bo`), los cuales a su vez indexan catálogos departamentales o de hospitales públicos.
2. **Federación Académica:** Un catálogo universitario (ej. `udabol_investigacion`) puede federar catálogos de facultades o laboratorios específicos (`lab_biologia`, `lab_telecomunicaciones`).
3. **Catálogos Transversales / Temáticos:** Catálogos temáticos (ej. `biodiversidad_amazonia`) pueden componerse a partir de subcatálogos que pertenecen a distintas instituciones de Bolivia, Perú y Brasil.

### El Desafío Arquitectónico:
> *¿Cómo permitir que un catálogo referencie recursivamente a otros catálogos a cualquier nivel de profundidad, preservando la soberanía de cada catálogo hijo, evitando ciclos infinitos de resolución ($A 	o B 	o A$) y garantizando la trazabilidad criptográfica de las delegaciones de confianza?*

---

## 2. Estado del Arte y Fundamentación en Estándares W3C

Para no crear un modelo ad-hoc incompatible, la especificación OKF se alinea con la recomendación **W3C DCAT v3 (Data Catalog Vocabulary - Agosto 2024)**:

```
                            MODELO W3C DCAT v3

                           dcat:Resource (Abstracta)
                                 ^
                                 │ (Herencia)
                           dcat:Dataset
                                 ^
                                 │ (Herencia recursiva)
                           dcat:Catalog
                                 │
                 ┌───────────────┴───────────────┐
                 │ dcat:catalog                  │ dcat:dataset
                 v                               v
            dcat:Catalog                   dcat:Dataset
         (Subcatálogo Hijo)              (Dataset de Datos)
```

### 2.1. Herencia Recursiva: `dcat:Catalog` es un `dcat:Dataset`
En DCAT v3, la clase `dcat:Catalog` se define como una subclase de `dcat:Dataset`. Esto otorga a un catálogo todas las propiedades de un conjunto de datos (título, descripción, fecha de emisión, licencias, identificador persistente IRI y procedencia PROV-O). En consecuencia, **un catálogo puede ser miembro de otro catálogo de forma nativa**.

### 2.2. Propiedades RDF Normativas
- **`dcat:catalog`**: Propiedad específica (subpropiedad de `dcat:resource` y de `dcterms:hasPart`) utilizada por un catálogo padre para declarar la inclusión de un subcatálogo.
- **`dcat:inCatalog`**: Propiedad inversa agregada en DCAT v3, que vincula al subcatálogo con su catálogo contenedor.

---

## 3. Especificación en el Formato de Conocimiento Abierto (OKF)

En el manifiesto `metadata.yaml` del bundle, un catálogo declara sus subcatálogos federados en la sección `catalogs`:

```yaml
id: udabol_investigacion
title: "Catálogo General de Investigación Científica - UDABOL"
type: catalog
domain: academia_bolivia
version: "1.0.0"

# Anclaje semántico W3C DCAT v3
dcat_catalog: "https://investigacion.udabol.edu.bo/catalog.jsonld"
dcat_type: "dcat:Catalog"

# =============================================================================
# FEDERACIÓN DE CATÁLOGOS HIJOS (Catalog-of-Catalogs)
# =============================================================================
catalogs:
  - id: lab_ciencias_salud
    title: "Laboratorio de Investigación en Ciencias de la Salud"
    ref: "okf://udabol_salud/catalog.yaml"
    multiaddr: "/ip4/200.87.120.50/tcp/4001/p2p/12D3KooWHealthLab..."
    trust_tier: tier_1_verified
    sync_policy:
      mode: lazy_fetch           # [lazy_fetch | full_mirror | headers_only]
      ttl_seconds: 86400

  - id: lab_inteligencia_artificial
    title: "Laboratorio de IA y Sistemas Distribuidos"
    ref: "okf://udabol_ia/catalog.yaml"
    multiaddr: "/ip4/200.87.120.51/tcp/4001/p2p/12D3KooWIALab..."
    trust_tier: tier_1_verified
```

---

## 4. Resolución de Nombres Jerárquica y Prevención de Ciclos

### 4.1. Sintaxis de Enrutamiento de Catálogos Compuestos
Para acceder a un recurso situado en un subcatálogo federado a través del catálogo padre, se utiliza la sintaxis de delegación punto (`.`):

```
<dataset>.<recurso>@<subcatalogo>.<catalogo_raiz>
```
*Ejemplo:* `secuenciacion_adn.muestras@lab_ciencias_salud.udabol_investigacion`

### 4.2. Algoritmo de Prevención de Ciclos y Control de Recursión
Para evitar que catálogos que se referencian mutuamente ($A 	o B 	o A$) provoquen bucles infinitos en el resolvedor o en la red P2P, todo motor OKF conforme DEBE implementar:
1. **Límite de Profundidad ($MAX\_DEPTH = 8$):** Las consultas que superen 8 niveles de anidamiento de catálogos son abortadas inmediatamente con un error `MAX_CATALOG_DEPTH_EXCEEDED`.
2. **Vector de Visitas Criptográfico ($Hop-Path$):** Cada petición de resolución añade el `NodeID` o hash del catálogo consultado. Si un catálogo recibe una petición que ya contiene su propio identificador, la resolución se detiene detectando un ciclo (`CIRCULAR_CATALOG_REFERENCE`).
3. **Caché Jerárquica con TTL:** Las entradas delegadas se almacenan con el `ttl_seconds` especificado en la relación.

---

## 5. Referencias Bibliográficas (APA 7ma / IEEE)

* **World Wide Web Consortium (W3C). (2024).** *Data Catalog Vocabulary (DCAT) - Version 3* (W3C Recommendation, 22 August 2024). W3C. https://www.w3.org/TR/vocab-dcat-3/
* **Martínez, J., & Gómez, A. (2025).** Propuesta de implementación de catálogos federados para espacios de datos sobre DataHub. *arXiv preprint arXiv:2509.17649*.
* **Gaia-X European Association for Data and Cloud AISBL. (2022).** *Gaia-X Architecture Document: Federation Services Specification*.
* **European Commission. (2023).** *DCAT-AP: Application Profile for Data Portals in Europe (v2.1.1)*. Interoperable Europe.
