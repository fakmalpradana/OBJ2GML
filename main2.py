import subprocess as sp
import time

start = time.time()

jln = "data/JALAN/AW10-11 Z37.obj"

tx = 694477.272
ty = 9310384.0900

# Translasi Objek Menuju Koordinat UTM
sp.call([
    "go", "run", "translate.go", 
    f"-input={jln}", 
    f"-tx={tx}", 
    f"-ty={ty}",
    "-tz=-13.848196"
])

# Convert OBJ ke CityGML
sp.call([
    "go", "run", "obj2gml-transport.go",
    "-input", f"{jln}_translated",
    "-output", f"export/{jln}_translated_gml"
])

# # Merge Keseluruhan CityGML file menjadi 1 file
# sp.call([
#     "go", "run", "mergegml.go",
#     "-input", f"export/{nlp}_{sub_grid}.obj_translated_gml",
#     "-output", f"percepatan/citygml/{nlp}_{sub_grid}.gml"
# ])

end = time.time() - start
print(f"durasi : {end} s")