import subprocess as sp
import time

start = time.time()

nlp = "AG_09"
sub_grid = "C"
bo = "percepatan/OBJ/AG_09/AG_09_C/AG-09-C_BO_Caesar Yoga_BUFFER_Lengkap.geojson"

tx = 692542.174723411328159
ty = 9326588.18167

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

# Generate MTL
sp.call([
    "python", "semantic_mapping.py",
    "--obj-dir", f"export/{nlp}_{sub_grid}.obj_translated",
    "--geojson", f"{bo}"
])

# Convert OBJ ke CityGML lod2
sp.call([
    "go", "run", "obj2lod2gml.go",
    "-input", f"export/{nlp}_{sub_grid}.obj_translated",
    "-output", f"export/{nlp}_{sub_grid}.obj_translated_gml"
])

# # Convert OBJ ke CityGML lod1
# sp.call([
#     "go", "run", "obj2gml.go",
#     "-input", f"export/{nlp}_{sub_grid}.obj_translated",
#     "-output", f"export/{nlp}_{sub_grid}.obj_translated_gml"
# ])

# Merge keseluruhan CityGMl lod2 file menjadi 1 file
sp.call([
    "python", "lod2merge.py",
    f"export/{nlp}_{sub_grid}.obj_translated_gml",
    f"percepatan/citygml/{nlp}_{sub_grid}.gml",
    "--name", f"{nlp}_{sub_grid}"
])

# # Merge Keseluruhan CityGML lod1 file menjadi 1 file
# sp.call([
#     "go", "run", "mergegml.go",
#     "-input", f"export/{nlp}_{sub_grid}.obj_translated_gml",
#     "-output", f"percepatan/citygml/{nlp}_{sub_grid}.gml"
# ])

end = time.time() - start
print(f"durasi : {end} s")