# ARQUITECTURA P2P, CONECTIVIDAD Y RESOLUCIÓN DISTRIBUIDA
## Principios de Redes Descentralizadas, Modelos Matemáticos y Subespacios de Información

**Título del Proyecto de Grado:**  
# "Diseño e Implementación del Formato de Conocimiento Abierto Orientado a Datos"

**Postulante:** Andres Humberto Chirinos Lizondo  
**Carrera:** Ingeniería de Sistemas — Universidad de Aquino Bolivia (UDABOL)  
**Sección Normativa de Referencia:** [`SPEC.md (§15)`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md)  
**Cuaderno NotebookLM de Respaldo:** [Sistemas Distribuidos P2P y Subespacios de Información](https://notebook.google.com/notebook/b952daca-ed07-477d-aa73-964002a1f485) (`b952daca-ed07-477d-aa73-964002a1f485`)  
**Cuaderno de Simulación Computacional:** [`notebooks/investigacion_p2p_subespacios.ipynb`](file:///home/andreschirinos/Proyectos/datamesh-sdk/notebooks/investigacion_p2p_subespacios.ipynb)

---

## 1. Fundamentos Matemáticos: Coherencia y Estabilidad en Redes P2P

Para que una red distribuida en Internet no colapse ante la desconexión constante de nodos (*churn*) ni dependa de un servidor maestro, el Formato de Conocimiento Abierto adopta los principios matemáticos de **Kademlia** y sus extensiones de seguridad y particionamiento:

### 1.1. Métrica de Distancia XOR y Convergencia Unidireccional
* **Autores Fundacionales:** Petar Maymounkov y David Mazières (MIT PDOS, 2002).
* **Definición Matemática:** La distancia lógica entre dos identificadores de $b$ bits ($x$ e $y$) se define mediante el operador binario XOR:
  $$d(x, y) = x \oplus y$$
* **Propiedades Críticas para OKF:**
  1. *Identidad:* $d(x, x) = 0$ y $d(x, y) > 0$ si $x \neq y$.
  2. *Simetría:* $d(x, y) = d(y, x)$. Permite que un nodo aprenda de los mensajes de búsqueda que recibe para actualizar sus propias tablas de rutas (*k-buckets*).
  3. *Unidireccionalidad:* Para cualquier clave fija $x$ y cualquier distancia $d$, **existe un único $y$** tal que $d(x, y) = d$. Esta propiedad asegura que las búsquedas dirigidas a un mismo catálogo converjan estrictamente por el mismo camino independientemente del nodo originador, permitiendo un almacenamiento en caché altamente coherente.
* **Complejidad de Enrutamiento:** Las tablas se estructuran en $k$-buckets organizados por la longitud de prefijo común, garantizando una complejidad de búsqueda acotada a $O(\log N)$ saltos, donde $N$ es el número total de nodos en la red.

### 1.2. Modelado Formal del Recuento de Saltos (Hop Count Distribution)
* **Autores:** Stefanie Roos, Hani Salah y Thorsten Strufe (TU Darmstadt, 2013).
* **Aporte a la Arquitectura:** Modelan el proceso de enrutamiento como una **cadena de Markov en un espacio de estados $\alpha$-dimensional** (donde $\alpha$ es el grado de concurrencia de búsqueda).
* Demuestran formalmente que retornar $\beta > 1$ contactos por cada salto reduce drásticamente el retardo y garantiza una convergencia de rutas estable en entornos dinámicos con fallas de red.

### 1.3. Resiliencia y Mitigación de Ataques Sybil (S/Kademlia)
* **Autores:** Ingmar Baumgart y Sebastian Mies (Universität Karlsruhe, 2007).
* **Mecanismos de Protección:**
  - *Crypto Puzzles:* Para evitar la generación masiva de identidades falsas (ataque Sybil), la generación del `NodeID` requiere resolver acertijos criptográficos estáticos ($c_1$) y dinámicos ($c_2$) con complejidad $O(2^{c_1} + 2^{c_2})$.
  - *Búsquedas Paralelas sobre $d$ Rutas Disjuntas:* El resolvedor lanza consultas paralelas por caminos completamente independientes, logrando una tasa de éxito de búsqueda superior al 99% incluso en presencia de un 20% de nodos adversarios maliciosos.

### 1.4. Separación Estricta de Planos de Control y Datos
* **Autores:** Carlos Giongo y Adriano Fiorese (UDESC, 2026).
* **Arquitectura de Transporte Híbrido:**
  - **Plano de Control sobre UDP:** Mantenimiento de la DHT, sondeos de disponibilidad (*keep-alive*), pings y actualización de $k$-buckets con mensajes ultraligeros de baja latencia.
  - **Plano de Datos sobre QUIC:** Transferencia multiplexada de archivos Parquet, esquemas y fragmentos sin bloqueo de cabeza de línea (*head-of-line blocking*), con cifrado TLS 1.3 nativo y tiempos de conexión 0-RTT.
  - **Erasure Coding $(n, k)$:** Para garantizar disponibilidad sin la sobrecarga de replicar archivos enteros ($M+1$), los datasets masivos se fragmentan en $n$ partes de las cuales cualquier subconjunto $k$ permite la reconstrucción exacta.

---

## 2. Capacidad de Búsqueda y Recuperación de Información (P2P-IR)

Para superar la limitación de las DHTs tradicionales (que solo permiten búsquedas por clave exacta) y dotar a la red de búsqueda semántica distribuida, OKF implementa la arquitectura de **Peer-to-Peer Information Retrieval (P2P-IR)**:

```
  ===========================================================================
  CAPA 4: MODELO DE RECUPERACIÓN & RELEVANCIA (Aberer et al., EPFL)
  * Cálculo de distancia semántica, modelos de espacio vectorial (VSM) y ranking
  ===========================================================================
                                     |
  ===========================================================================
  CAPA 3: GESTIÓN DE CONTENIDO & DIGEST QUERIES
  * Mapeo de términos a claves KBR; posting lists locales y filtros Bloom
  ===========================================================================
                                     |
  ===========================================================================
  CAPA 2: ENRUTAMIENTO BASADO EN CLAVES (Key-Based Routing - KBR)
  * Overlay estructurado Kademlia DHT; resolución independiente de IPs
  ===========================================================================
                                     |
  ===========================================================================
  CAPA 1: TRANSPORTE Y COMUNICACIONES FÍSICAS
  * QUIC / UDP (Control y Datos) y TCP+TLS (Fallback)
  ===========================================================================
```

### 2.1. Búsqueda Semántica Local como Capa de Servicio
En lugar de forzar a la DHT a indexar vectores densos multidimensionales (lo cual fragmentaría la red y generaría un tráfico inmanejable), la red adopta el modelo de **Local Approximate Nearest Neighbor (ANN)**:
1. La consulta en lenguaje natural genera un embedding contextual.
2. La DHT identifica el subconjunto de nodos soberanos responsables del dominio temático.
3. Cada nodo candidato ejecuta localmente una búsqueda ANN (usando grafos HNSW) sobre sus conceptos registrados y retorna únicamente el *top-k* de resultados más relevantes.

---

## 3. Subespacios de Información y Dominios Soberanos (Data Spaces)

Uno de los grandes desafíos de Internet es permitir que organizaciones afines (universidades, alcaldías, centros de investigación) creen **subespacios de información acotados** dentro de la infraestructura global abierta sin quedar aisladas del resto de la red.

### 3.1. Particionamiento Topológico (Protocol Labs Research)
* **Autores:** João Monteiro, Pedro Ákos Costa, João Leitão, Alfonso de la Rocha y Yiannis Psaras (2022).
* **Soft Partitioning (Particionamiento Suave):**
  - Se antepone un prefijo de espacio (por ejemplo `bo:edu:` o `bo:gob:`) a los identificadores de nodos y recursos.
  - La métrica XOR agrupa a los nodos con el mismo prefijo más cerca lógicamente, llenando sus $k$-buckets preferentemente con pares de su propia institución o comunidad.
  - *Ventaja:* La resolución dentro del subespacio ocurre en 1 o 2 saltos locales, pero el nodo sigue siendo capaz de enrutar consultas hacia el resto de la red global.
* **Hard Partitioning (Particionamiento Duro):**
  - Sub-redes físicamente separadas para entornos con altos requisitos de privacidad o secreto de estado, interconectadas exclusivamente mediante nodos puente (*indexer services*) con acuerdos explícitos de federación.

### 3.2. Espacios de Datos Interoperables (Data Spaces Compass - DSSC)
- El contenedor OKF adopta el marco del **Data Spaces Interoperability Compass (DSIC)**:
- Cada subespacio se rige por un marco soberano común:
  - Catálogos DCAT v2 federados.
  - Contratos de uso y retención de políticas (políticas `zero_microdata`).
  - Modelo de **Federación-de-Federaciones (FoF)**: permite que el subespacio universitario de la UDABOL se federé transparentemente con el subespacio de datos públicos del Estado boliviano.

### 3.3. Canales Pub-Sub mediante GossipSub
- Para la sincronización de subespacios de alta rotación, se utiliza **GossipSub v1.1**:
  - *MeshSub:* Los nodos suscritos a un tópico específico (ej. `/okf/announce/salud_bolivia`) forman una malla densa y estructurada para el reenvío directo de mensajes de alta velocidad.
  - *Gossip Piggybacking:* Intercambio de metadatos de control (`IHAVE` / `IWANT`) con nodos fuera de la malla para reparar caminos rotos y mitigar el *churn* sin saturar el ancho de banda.

---

## 4. El Modelo de "DNS Soberano" para Datos: `dataset.recurso@catalogo`

La resolución de recursos combina la semántica de DNS con la robustez criptográfica de Kademlia:

```
                       ARQUITECTURA DE RESOLUCIÓN

  [Cliente / Investigador]
           |
           | 1. Consulta: okf-dig calidad_aire.pm25_horario@lapaz_alcaldia
           v
  [Resolvedor KBR / DHT]
           |
           | 2. Localiza nodo responsable en O(log N) vía Métrica XOR
           v
  [Nodo Soberano: lapaz_alcaldia]
           |
           | 3. Retorna registros criptográficos firmados (Ed25519)
           |    SOA -> Autoridad y Trust Tier
           |    CAT -> CID del bundle metadata.yaml
           |    RES -> Multiaddresses (QUIC/TCP/P2P) + SHA-256
           |    SCH -> Contrato Frictionless Table Schema
           |    LIN -> Linaje W3C PROV-O
           v
  [Descarga Directa Cero-Copia vía QUIC Streams]
```

### 4.1. Catálogo Completo de Registros Soberanos

| Registro | Nombre Formal | Contenido y Función |
| :--- | :--- | :--- |
| **`SOA`** | Start of Authority | Identificador del nodo soberano (`NodeID`), clave pública Ed25519, Trust Tier y número serial. |
| **`CAT`** | Catalog Manifest | Puntero CID al archivo `metadata.yaml` del bundle y estado de vigencia. |
| **`RES`** | Resource Location | Direcciones `multiaddr` de acceso (P2P, QUIC, HTTPS, S3) y digest SHA-256 del archivo físico. |
| **`SCH`** | Syntactic Schema | Enlace al contrato de esquema (Table Schema / JSON Schema) y hash del contrato. |
| **`LIN`** | Lineage & PROV-O | Actividad `prov:Activity`, script de transformación ejecutado y fuentes primarias consumidas. |
| **`SRV`** | Dynamic Service | Endpoints para streaming y consultas en vivo mediante Apache Arrow Flight SQL o DuckDB. |

---

## 5. Referencias Bibliográficas Fundacionales (APA 7ma / IEEE)

* **Maymounkov, P., & Mazières, D. (2002).** Kademlia: A peer-to-peer information system based on the XOR metric. In *International Workshop on Peer-to-Peer Systems (IPTPS '01)* (pp. 53-65). Springer, Berlin, Heidelberg.
* **Aberer, K., Klemm, F., Rajman, M., & Wu, J. (2004).** *An Architecture for Peer-to-Peer Information Retrieval*. Technical Report, School of Computer and Communication Sciences, EPFL, Lausanne, Switzerland.
* **Roos, S., Salah, H., & Strufe, T. (2013).** Comprehending Kademlia Routing: A Theoretical Framework for the Hop Count Distribution. In *IEEE 38th Conference on Local Computer Networks (LCN 2013)* (pp. 577-584). IEEE.
* **Baumgart, I., & Mies, S. (2007).** S/Kademlia: A Practicable Approach Towards Secure Key-Based Routing. In *2007 International Conference on Parallel and Distributed Systems* (pp. 1-8). IEEE.
* **Monteiro, J., Costa, P. Á., Leitão, J., de la Rocha, A., & Psaras, Y. (2022).** *Enriching Kademlia by Partitioning*. Protocol Labs Research Report.
* **Giongo, C., & Fiorese, A. (2026).** A Decentralized Architecture for Scientific Document Storage and Retrieval Using QUIC, Erasure Coding, and DHT-Based Overlays. In *53º Seminário Integrado de Software e Hardware (SEMISH 2026)* (pp. 1-10). UDESC.
* **Data Spaces Support Centre (DSSC). (2025).** *DSIC Data Spaces Interoperability Compass: Living Document*. Centre of Excellence for Data Sharing (CoE-DSC).
* **Vyzovitis, D., Dias, Y., & Psaras, Y. (2020).** GossipSub: A secure pubsub protocol for unstructured, decentralised P2P overlays. *arXiv preprint arXiv:2007.02754*.
