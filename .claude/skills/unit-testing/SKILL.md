---
name: unit-testing
description: Este skill contiene información sobre cómo escribir pruebas unitarias efectivas en Go, incluyendo el uso de la biblioteca de testing estándar y técnicas para aislar dependencias.
---

- Los nombres de los tests deben ser descriptivos y reflejar el comportamiento que se está probando. Por ejemplo, `TestCalculateTotal_WithValidInput_ReturnsCorrectSum` es más informativo que `TestCalculateTotal`.
- Los nombres de los tests deben estár en inglés.
- Los tests deben ser independientes entre sí. Cada test debe poder ejecutarse de manera aislada sin depender del estado o los resultados de otros tests.
- Lost tests deben ser deterministas. Evita el uso de datos aleatorios o dependencias externas que puedan cambiar entre ejecuciones.
- Lost tests deben deben seguir el patrón Given-When-Then (Dado-Cuando-Entonces) para estructurar las pruebas de manera clara y comprensible.
- Utiliza `assert` y `require` de la biblioteca `testify` para realizar afirmaciones en tus tests. `assert` permite continuar la ejecución del test incluso si una afirmación falla, mientras que `require` detiene la ejecución del test si una afirmación falla.
- Utiliza `t.Run` para organizar subtests dentro de un test principal. Esto permite agrupar pruebas relacionadas y proporciona una mejor visibilidad en los resultados de las pruebas.
- Utiliza `asserts` para comparar valores esperados y reales en tus tests. Esto ayuda a identificar rápidamente las diferencias entre lo que se esperaba y lo que realmente se obtuvo. Utiliza mensajes descriptivos en las afirmaciones, en inglés, para facilitar la comprensión de los errores.
- Siempre verifica `assert.Nil(err)` después de operaciones que puedan devolver un error. Esto asegura que los errores se manejen adecuadamente y evita que los tests pasen silenciosamente cuando deberían fallar.
- Los tests van dentro del mismo paquete que el código que están probando. Esto permite acceder a funciones y variables no exportadas, lo que facilita la prueba de la lógica interna del paquete.
- Los tests deben estar en archivos con el sufijo `_test.go`. Esto es una convención en Go que permite al compilador y a las herramientas de prueba identificar fácilmente los archivos que contienen pruebas.
- Los `asserts` y `requires` deben estár al final del test (en la sección Then) y no en la sección Given o When. Esto asegura que las afirmaciones se realicen después de que se haya ejecutado la lógica que se está probando, proporcionando resultados más precisos y relevantes.