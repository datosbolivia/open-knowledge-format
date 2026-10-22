# ESPECIFICACIÓN TÉCNICA Y PROBLEMÁTICA 05
## Red P2P, Direccionamiento DNS Soberano y Creación de Subespacios de Información en Internet

**Estado:** Documento de Especificación Normativa (Estilo W3C Working Draft)  
**Módulo:** Protocolo de Red P2P, Métrica XOR y Resolución Descentralizada OKF  
**Repositorio:** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format)  
**Sección Normativa de Referencia:** [`SPEC.md (§15)`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md)  
**Cuaderno de Investigación:** [Sistemas Distribuidos P2P y Subespacios de Información](https://notebook.google.com/notebook/b952daca-ed07-477d-aa73-964002a1f485) (`b952daca-ed07-477d-aa73-964002a1f485`)  
**Cuaderno de Simulación:** [`notebooks/investigacion_p2p_subespacios.ipynb`](file:///home/andreschirinos/Proyectos/datamesh-sdk/notebooks/investigacion_p2p_subespacios.ipynb)

---

## 1. Declaración de la Problemática

En Internet, los nombres de dominio y la localización de recursos dependen de una jerarquía centralizada gobernada por ICANN y grandes proveedores cloud:
1. **Puntos Únicos de Fallo y Vulnerabilidad a Censura:** Si un servidor DNS institucional se desconecta o es revocado, todos los catálogos y datasets dependientes quedan inaccesibles, incluso si los datos residen en servidores locales intactos.
2. **Falta de Localidad en Búsquedas:** Las redes P2P tradicionales tratan a todos los nodos por igual; no proveen mecanismos nativos para que universidades o ministerios formen **subespacios de información acotados** (*Data Spaces*) donde las búsquedas resuelvan en milisegundos en la red local (LAN/MAN) sin saturar enlaces externos.
3. **Rigidez de las DHTs Clásicas:** Las DHTs estándar (Chord, Pastry) solo soportan búsquedas por hash exacto y carecen de capacidad de búsqueda semántica o de una interfaz intuitiva estilo DNS para ingenieros y científicos.

### El Desafío Arquitectónico:
> *¿Cómo diseñar un protocolo de red P2P que asegure una red coherente y estable con enrutamiento $O(\log N)$, que soporte un "DNS Soberano de Datos" con la sintaxis `dataset.recurso@catalogo` y que permita la creación de subespacios de información lógicamente agrupados sobre la infraestructura abierta de Internet?*

---

## 2. Modelos Matemáticos y Arquitectura de Sistemas Distribuidos

```
                         ARQUITECTURA DE RESOLUCIÓN P2P

   Consulta: okf-dig calidad_aire.pm25_horario@lapaz_alcaldia
                              │
                              │ 1. Mapeo XOR a Clave de 256 bits
                              v
   [Tabla de Hash Distribuida - Kademlia DHT (Maymounkov & Mazières)]
   * Métrica de distancia unidireccional: d(x, y) = x ⊕ y
   * Tablas k-buckets organizadas por prefijo común
   * Enrutamiento iterativo paralelo (α = 3) en O(log N) saltos
                              │
                              │ 2. Localización de Pares Candidatos
                              v
   [Separación de Planos de Control y Datos (Giongo & Fiorese, 2026)]
   * Plano de Control (UDP): Sondeos de liveness y k-buckets (< 1 ms)
   * Plano de Datos (QUIC): Transferencia multiplexada 0-RTT cifrada
                              │
                              │ 3. Retorno de Registros Soberanos Firmados
                              v
   [Registros DNS Soberanos: SOA, CAT, RES, SCH, LIN, SRV]
```

### 2.1. Métrica XOR y Convergencia Unidireccional
- **Petar Maymounkov y David Mazières (MIT, 2002):** La métrica $d(x, y) = x \oplus y$ garantiza que las búsquedas dirigidas a un mismo catálogo converjan estrictamente por el mismo camino independientemente del nodo originador, permitiendo un almacenamiento en caché coherente y determinista.
- **Roos, Salah y Strufe (TU Darmstadt, 2013):** Modelan analíticamente el recuento de saltos mediante cadenas de Markov $\alpha$-dimensionales, demostrando que retornar $\beta > 1$ contactos por salto minimiza los retardos ante el abandono de nodos (*churn*).

### 2.2. Seguridad Sybil y Disjoint Paths (S/Kademlia)
- **Baumgart y Mies (2007):** Acertijos criptográficos estáticos ($c_1$) y dinámicos ($c_2$) con complejidad $O(2^{c_1} + 2^{c_2})$ para la generación de `NodeID`, junto con búsquedas paralelas por $d$ caminos disjuntos (99% de éxito de búsqueda bajo 20% de nodos maliciosos).

### 2.3. Búsqueda y Recuperación de Información (P2P-IR)
- **Aberer, Klemm, Rajman y Wu (EPFL, 2004):** Arquitectura de 4 capas para Information Retrieval en P2P, combinando Key-Based Routing (KBR) en Capa 2 con indexación semántica local Approximate Nearest Neighbor (ANN con HNSW) en Capa 3/4.

---

## 3. Subespacios de Información y Particionamiento Topológico

Para permitir que instituciones (como la Universidad de Aquino Bolivia o la UMSA) formen subespacios de información soberanos dentro de Internet:

### 3.1. Soft Partitioning (Protocol Labs Research, Monteiro et al., 2022)
- Se antepone un prefijo de espacio institucional (ej. `bo:edu:udabol:`) al espacio de claves de 256 bits.
- La métrica XOR agrupa a los nodos con el mismo prefijo más cerca lógicamente, llenando sus $k$-buckets preferentemente con pares de su propia universidad.
- **Resultado:** Las consultas internas resuelven en 1 o 2 saltos en la red universitaria sin consumir ancho de banda internacional, manteniendo al mismo tiempo la capacidad de saltar a la red global mediante los contactos de prefijo distante.

### 3.2. Espacios de Datos Federados (Data Spaces Interoperability Compass - DSSC)
- Se implementa el modelo de **Federación-de-Federaciones (FoF)**:
- Los subespacios comparten vocabularios semánticos y políticas de gobernanza (`zero_microdata`), preservando la soberanía de los datos en su nodo de origen.

---

## 4. El Modelo de "DNS Soberano" y la Herramienta `okf-dig`

### 4.1. Sintaxis Universal de Consulta
```text
<dataset>.<recurso>@<catalogo>
```

### 4.2. Tipos de Registros Soberanos

| Registro | Nombre | Función |
| :--- | :--- | :--- |
| **`SOA`** | Start of Authority | Identifica al nodo emisor soberano, su clave Ed25519 y su nivel de confianza (*Trust Tier*). |
| **`CAT`** | Catalog Manifest | CID del archivo `metadata.yaml` del bundle. |
| **`RES`** | Resource Location | Multiaddresses (`multiaddr`) de descarga (QUIC, TCP, HTTPS, S3) y digest SHA-256. |
| **`SCH`** | Syntactic Contract | Contrato de esquema (Frictionless / JSON Schema) y hash del validador. |
| **`LIN`** | Lineage Record | Actividad W3C PROV-O y script transformador (`prov:wasGeneratedBy`). |
| **`SRV`** | Dynamic Service | Endpoints de streaming analítico (Arrow Flight SQL / DuckDB). |

---

## 5. Referencias Bibliográficas (APA 7ma / IEEE)

* **Maymounkov, P., & Mazières, D. (2002).** Kademlia: A peer-to-peer information system based on the XOR metric. In *IPTPS '01* (pp. 53-65). Springer.
* **Aberer, K., Klemm, F., Rajman, M., & Wu, J. (2004).** *An Architecture for Peer-to-Peer Information Retrieval*. EPFL.
* **Roos, S., Salah, H., & Strufe, T. (2013).** Comprehending Kademlia Routing: A Theoretical Framework for the Hop Count Distribution. In *IEEE LCN 2013* (pp. 577-584). IEEE.
* **Monteiro, J., Costa, P. Á., Leitão, J., de la Rocha, A., & Psaras, Y. (2022).** *Enriching Kademlia by Partitioning*. Protocol Labs Research.
* **Giongo, C., & Fiorese, A. (2026).** A Decentralized Architecture for Scientific Document Storage and Retrieval Using QUIC, Erasure Coding, and DHT-Based Overlays. In *SEMISH 2026* (pp. 1-10). UDESC.
* **Data Spaces Support Centre (DSSC). (2025).** *DSIC Data Spaces Interoperability Compass: Living Document*. CoE-DSC.
