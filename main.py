import subprocess as sp
import time

start = time.time()

nlp = "AN_13"
sub_grid = "A"
bo = "percepatan/OBJ/AN_13/AN_13_A/AN-13-A-BO.geojson"

tx = 696436.71322
ty = 9320053.52220

# Pemisahan Bangunan
sp.call([
    "go", "run", "objseparator.go", 
    f"-cx={tx}", f"-cy={ty}",
    f"percepatan/OBJ/{nlp}/{nlp}_{sub_grid}/{nlp}_{sub_grid}.obj", 
    f"{bo}"
])

# Translasi Objek Menuju Koordinat UTM
sp.call([
    "go", "run", "translate.go", 
    f"-input=export/{nlp}_{sub_grid}.obj", 
    f"-tx={tx}", 
    f"-ty={ty}",
    "-tz=0"
])

# Convert OBJ ke CityGML
sp.call([
    "go", "run", "obj2gml.go",
    "-input", f"export/{nlp}_{sub_grid}.obj_translated",
    "-output", f"export/{nlp}_{sub_grid}.obj_translated_gml"
])

# Merge Keseluruhan CityGML file menjadi 1 file
sp.call([
    "go", "run", "mergegml.go",
    "-input", f"export/{nlp}_{sub_grid}.obj_translated_gml",
    "-output", f"percepatan/citygml/{nlp}_{sub_grid}.gml"
])

end = time.time() - start
print(f"durasi : {end} s")