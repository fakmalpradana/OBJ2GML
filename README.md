# OBJ to CityGML Converter

OBJ Data extractor untuk keperluan 3D City Building, dapat mengakomodir tingkat kedetailan LOD1 - LOD3
Dengan catatan memiliki data OBJ dan GeoJSON untuk Building Outline
## Petnjuk Penggunaan
Urutan Algoritma
1. OBJ Separator -> memisahkan setiap OBJ
2. Translate -> translasi OBJ menuju UTM
3. OBJ to GML -> konversi OBJ kedalam format CityGML
4. elevate -> translasi Z setiap GML
5. Merge GML -> merge seluruh GML menjadi 1 file

### Run sysntax
#### Separator
```bash
go run objseparator.go [file path OBJ] [file path BO GeoJSON]
```
#### Translation
edit input file dan koordinat dalam script dahulu
```bash
go run translate.go
```
#### OBJ to CityGML
```bash
go run obj2gml.go -input [folder path] -output [folder path]
```
#### Merge CityGML
```bash
go run mergegml.go -input [folder path] -output [folder path]
```