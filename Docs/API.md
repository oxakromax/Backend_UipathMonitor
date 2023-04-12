Documentación de la función `GetUsers`
======================================

Descripción
-----------

La función `GetUsers` es un controlador para la API que permite obtener usuarios de la base de datos. Puede realizar
consultas específicas y filtrar usuarios según diferentes criterios, como su pertenencia a una organización, rol o
proceso.

Parámetros
----------

La función acepta un JSON con los siguientes campos opcionales:

* `query`: Una consulta de filtrado que se aplica a la base de datos de usuarios. Ejemplo: `"name LIKE 'John%'"`.
* `relational_condition`: Una cadena que especifica la condición relacional que se aplicará al filtrado. Valores
  posibles: `"NotInOrg"`, `"InOrg"`, `"NotInRole"`, `"InRole"`, `"NotInProcess"`, `"InProcess"`.
* `relational_query`: Una cadena que contiene una lista de ID de elementos (organizaciones, roles o procesos) separados
  por comas, que se utilizarán en la condición relacional. Ejemplo: `"1,2,3"`.

Ejemplo de uso
--------------

Realizar una solicitud HTTP GET a la API con el siguiente formato:

```json
{
  "query": "name LIKE 'John%'",
  "relational_condition": "NotInOrg",
  "relational_query": "1,2,3"
}
```

Esto devuelve todos los usuarios cuyo nombre comienza con "John" y que no están en las organizaciones con ID 1, 2 y 3.

Código
------

A continuación se presenta el código completo de la función `GetUsers`:

\[Inserte el código aquí\]

Detalles de implementación
--------------------------

La función realiza las siguientes acciones:

1. Verifica si se proporciona el parámetro "id" en la URL. Si es así, busca al usuario con ese ID y lo devuelve como
   resultado.
2. Analiza los campos opcionales del JSON proporcionado (`query`, `relational_condition`, `relational_query`) y los
   utiliza para realizar consultas y filtrar usuarios en la base de datos.
3. Si se proporciona un valor para `query`, realiza una consulta en la base de datos de usuarios utilizando ese valor.
4. Aplica la condición relacional (`relational_condition`) y la consulta relacional (`relational_query`) si se
   proporcionan.
5. Devuelve el resultado en formato JSON.