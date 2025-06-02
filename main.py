import subprocess as sp
import time
import os
import shutil

start = time.time()

nlp = "AG_09"
sub_grid = "C"
bo = "percepatan/OBJ/AG_09/AG_09_C/AG-09-C_BO_Caesar Yoga_BUFFER_Lengkap.geojson"

tx = 692542.174723411328159
ty = 9326588.18167

def delete_files(directory):
    files_to_delete = ["12030.obj", "12030.mtl", "12030.gml"]
    
    for filename in files_to_delete:
        file_path = os.path.join(directory, filename)
        if os.path.exists(file_path):
            try:
                os.remove(file_path)
                print(f"Deleted: {file_path}")
            except Exception as e:
                print(f"Error deleting {file_path}: {e}")
        else:
            print(f"File not found: {file_path}")

def delete_directories(directories):
    for directory in directories:
        if os.path.exists(directory) and os.path.isdir(directory):
            try:
                shutil.rmtree(directory)
                print(f"Deleted directory: {directory}")
            except Exception as e:
                print(f"Error deleting directory {directory}: {e}")
        else:
            print(f"Directory not found or not a directory: {directory}")

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
delete_files(f"export/{nlp}_{sub_grid}.obj_translated")

# Convert OBJ ke CityGML lod2
sp.call([
    "go", "run", "obj2lod2gml.go",
    "-input", f"export/{nlp}_{sub_grid}.obj_translated",
    "-output", f"export/{nlp}_{sub_grid}.obj_translated_gml"
])
delete_files(f"export/{nlp}_{sub_grid}.obj_translated_gml")

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

delete_directories(
    [
        f"export/{nlp}_{sub_grid}.obj",
        f"export/{nlp}_{sub_grid}.obj_translated",
        f"export/{nlp}_{sub_grid}.obj_translated_gml"
    ]
)

end = time.time() - start
print(f"durasi : {end} s")