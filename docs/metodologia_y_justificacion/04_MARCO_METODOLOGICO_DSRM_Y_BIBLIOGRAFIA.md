# MARCO METODOLÓGICO DSRM Y REVISIÓN BIBLIOGRÁFICA
## Fundamentación Epistemológica y Literatura Científica de Respaldo

**Proyecto de Grado:** Licenciatura en Ingeniería de Sistemas — Universidad de Aquino Bolivia (UDABOL)  
**Título del Proyecto:** *Diseño e Implementación del Formato de Conocimiento Abierto Orientado a Datos*  
**Autor:** Andres Humberto Chirinos Lizondo  
**Ecosistema:** [`open-knowledge-format`](file:///home/andreschirinos/Proyectos/open-knowledge-format) (Estándar/Contenedor) y [`datamesh-sdk`](file:///home/andreschirinos/Proyectos/datamesh-sdk) (SDK Hexagonal)

---

## 1. La Metodología Design Science Research (DSRM) como Eje Rector

Para un proyecto de grado en Ingeniería de Sistemas de la UDABOL, la adopción de **DSRM (Peffers et al., 2008; Hevner et al., 2004; March & Smith, 1995)** garantiza que el diseño del formato de datos y de su contenedor constituya una investigación científica rigurosa y no una simple tarea de desarrollo rutinario.

### 1.1. Distribución de los Cuatro Artefactos de March y Smith (1995)

```
  +-------------------------------------------------------------------------+
  |              TAXONOMÍA DE ARTEFACTOS DSRM EN EL PROYECTO                |
  +-------------------------------------------------------------------------+
  | 1. CONSTRUCTOS (Lenguaje y Conceptos):                                  |
  |    * Bounded Contexts, Open Knowledge Bundle, Concept Document,        |
  |      Trust Tiers, Sovereign Domain, Semantic Triad.                     |
  |    * Respaldo: open-knowledge-format & datamesh-sdk                     |
  +-------------------------------------------------------------------------+
  | 2. MODELOS (Especificaciones y Arquitectura):                           |
  |    * Especificación normativa SPEC.md y esquemas JSON formales.        |
  |    * Mapeos W3C (DCAT v2, PROV-O, SKOS).                               |
  |    * Diagrama de Arquitectura Hexagonal del SDK.                       |
  |    * Respaldo: open-knowledge-format (Norma) & datamesh-sdk (Diseño)    |
  +-------------------------------------------------------------------------+
  | 3. MÉTODOS (Protocolos, Algoritmos y Reglas):                           |
  |    * Algoritmos de verificación de hashes SHA-256 y llaves SSH.        |
  |    * Procedimientos de validación sintáctica de Table Schema.          |
  |    * Protocolos de direccionamiento canónico por tríadas.              |
  |    * Respaldo: open-knowledge-format (Reglas) & datamesh-sdk (Código)   |
  +-------------------------------------------------------------------------+
  | 4. INSTANCIACIONES (Software Operativo):                                |
  |    * Validador agnóstico de esquemas en Go / Python.                    |
  |    * Paquete datamesh-sdk en Python con motor DuckDB y Arrow.          |
  |    * Línea de comandos (CLI dm) para gestión de bundles.               |
  |    * Respaldo: open-knowledge-format (Validator) & datamesh-sdk (SDK)   |
  +-------------------------------------------------------------------------+
```

---

## 2. Bibliografía Académica y Normativa Formal (APA 7ma / IEEE)

### 2.1. Metodología de Ciencia del Diseño (DSRM)
* **Peffers, K., Tuunanen, T., Rothenberger, M. A., & Chatterjee, S. (2008).** A Design Science Research Methodology for Information Systems Research. *Journal of Management Information Systems*, 24(3), 45-77. https://doi.org/10.2753/MIS0742-1222240302
* **Hevner, A. R., March, S. T., Park, J., & Ram, S. (2004).** Design Science in Information Systems Research. *MIS Quarterly*, 28(1), 75-105.
* **March, S. T., & Smith, G. F. (1995).** Design and natural science research on technology. *Decision Support Systems*, 15(4), 251-266.
* **Engström, E., Storey, M. A., Runeson, P., Höst, M., & Baldassarre, M. T. (2020).** How software engineering research aligns with design science: A review. *Empirical Software Engineering*, 25(4), 2630-2660.
* **Wieringa, R. J. (2014).** *Design Science Methodology for Information Systems and Software Engineering*. Springer.

### 2.2. Paradigma Data Mesh y Contratos de Datos
* **Dehghani, Z. (2022).** *Data Mesh: Delivering Data-Driven Value at Scale*. O'Reilly Media.
* **Sanderson, C., Freeman, M., & Schmidt, B. E. (2025).** *Data Contracts: Developing Production-Grade Pipelines at Scale*. O'Reilly Media.

### 2.3. Arquitectura de Software y Patrones de Diseño
* **Cockburn, A., & Garrido de Paz, J. M. (2024).** *Hexagonal Architecture Explained: Ports and Adapters in Practice*. Addison-Wesley.
* **Evans, E. (2003).** *Domain-Driven Design: Tackling Complexity in the Heart of Software*. Addison-Wesley.
* **Martin, R. C. (2017).** *Clean Architecture: A Craftsman's Guide to Software Structure and Design*. Prentice Hall.

### 2.4. Estándares Internacionales W3C y Formatos Abiertos
* **World Wide Web Consortium (W3C). (2020).** *Data Catalog Vocabulary (DCAT) - Version 2*. W3C Recommendation. https://www.w3.org/TR/vocab-dcat-2/
* **World Wide Web Consortium (W3C). (2013).** *PROV-O: The PROV Ontology*. W3C Recommendation. https://www.w3.org/TR/prov-o/
* **World Wide Web Consortium (W3C). (2009).** *SKOS Simple Knowledge Organization System Reference*. W3C Recommendation. https://www.w3.org/TR/skos-reference/
* **Open Knowledge Foundation. (2024).** *Frictionless Data Specifications: Data Package & Table Schema*. https://datapackage.org/
* **Common Workflow Language (CWL) Working Group. (2020).** *Common Workflow Language Specifications (v1.2)*. https://w3id.org/cwl/v1.2/
* **Wilkinson, M. D., et al. (2016).** The FAIR Guiding Principles for scientific data management and stewardship. *Scientific Data*, 3, 160018.

### 2.5. Tecnologías de Ejecución Analítica Embebida
* **Raasveldt, M., & Mühleisen, H. (2019).** DuckDB: an Embeddable Analytical Database. In *SIGMOD '19* (pp. 1981-1987). ACM.
* **Apache Software Foundation. (2023).** *Apache Arrow: A cross-language development platform for in-memory analytics*. https://arrow.apache.org/
