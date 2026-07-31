---
name: gal-visa-credit
description: Lee la información de un resumen de tarjeta de crédito Visa de Banco Galicia, desde un PDF y la exporta en formato CSV.
---

# Especificaciones del archivo PDF (input)

A continuación se listan consideraciones importantes sobre el formato del resumen de tarjeta en formato PDF:

- Los consumos están agrupados por número de tarjeta de crédito.
- El detalle de los consumos, para cada tarjeta, comienzan a partir del texto *"DETALLE DE CONSUMO"*.
- El detalle de los consumos, para cada tarjeta, finalizan con un texto que contiene la frase *"Total consumos"*.
- La fechas de los consumos se encuentran en la columna *FECHA* y tienen formato `DD-MM-YY`.
- El concepto del consumo se encuentra en la columna *REFERENCIA*.
- Si es un consumo en un pago la columna *CUOTA* figura vacía. En caso contrario, la columna *CUOTA*, especifica la cuota actual y la cantidad de cuotas en formato `{CUOTA_ACTUAL}/{CANTIDAD_DE_CUOTAS}`.
- El identificador del consumo se encuentra en la columna *COMPROBANTE*.
- Si el consumo fue en moneda local (ARS), la columna *PESOS* contendrá el monto del consumo utilizando un "." como separador de miles y una "," como separador decimal. Si el consumo fue en moneda extranjera (USD) este campo estará vacío.
- Si el consumo fue en moneda extranjera (USD), la columna *DOLARES* contendrá el monto del consumo utilizando un "." como separador de miles y una "," como separador decimal. Si el consumo fue en moneda local (ARS) este campo estará vacío.

# Especificaciones del archivo CSV (output)

A continuación se listan consideraciones importantes sobre el formato del archivo de salida:

- Cada columna extraída del resumen PDF, debe estár en el archivo CSV. Respetando su orden en el archivo PDF.
- Utilizar ";" como separador de campos.
- Encerrar cadenas de texto con comillas.
- Los montos decimales deberán ser exportados sin el "." como separador de miles.
- Se debe indicar, en nuna nueva columna *BANCO*, el nombre del banco emisor de la tarjeta de crédito; en este caso "Banco Galicia".
- Se debe indicar, en nuna nueva columna *MARCA*, la marca de la tarjeta de crédito.
- Se debe indicar, en nuna nueva columna *TARJETA*, el número de tarjeta de crédito.
- Se debe indicar, en nuna nueva columna *PERIODO*, la fecha del resumen de tarjeta de crédito, en formato `YYYY-MM-DD`.
- Se debe indicar, en nuna nueva columna *TIPO_CONSUMO*; en este caso "CREDITO".

# Archivos de salidas

Generar un archivo de salida para cada moneda presente en el resumen (ARS y USD), teniendo en cuenta las siguientes consideraciones:

- El nombre del archivo de salida debe contener la fecha del resumen, el banco emisor, la marca de tarjeta y la moneda correspondiente (`{YYYYMM}_{ISSUER}_{BRAND}_{CURRENCY}`). Por ejemplo: `202301_GALICIA_VISA_ARS.csv` o `202301_GALICIA_VISA_ARS.csv`, según corresponda.
- El archivo en USD no debe contener la columna *PESOS*.
- El archivo en ARS no debe contener la columna *DOLARES*.
- El archivo en USD debe contener una nueva columna extra, llamada *MONEDA*, que debe contener el valor "USD" para todos los consumos.
- El archivo en ARS debe contener una nueva columna extra, llamada *MONEDA*, que debe contener el valor "ARS" para todos los consumos.