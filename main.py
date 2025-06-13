import subprocess as sp
import time
import os
import shutil

from findFile import find_complete_sets, read_and_convert_txt
from pathlib import Path

start = time.time()

root_dir = "percepatan_new/OBJ/2025_06_13"

file_set = find_complete_sets(root_dir)

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

for i in file_set:
    obj = i[0]
    coord = read_and_convert_txt(i[1])
    bo = i[2]

    root_path = Path(root_dir)
    obj_path = Path(obj)

    rel_path = obj_path.relative_to(root_path)
    folder_name = rel_path.parts[0]
    print(folder_name)

    output_path = f"{root_dir}/{folder_name}.gml".replace('OBJ', 'CityGML')
    os.makedirs(f"{root_dir}".replace('OBJ', 'CityGML'), exist_ok=True)

    # Pemisahan Bangunan
    sp.call([
        "go", "run", "objseparator.go", 
        f"-cx={coord[0]}", f"-cy={coord[1]}",
        f"{obj}", 
        f"{bo}"
    ])

    # Translasi Objek Menuju Koordinat UTM
    sp.call([
        "go", "run", "translate.go", 
        f"-input=export/{rel_path.parts[1]}", 
        f"-tx={coord[0]}", 
        f"-ty={coord[1]}",
        "-tz=0"
    ])

    # Generate MTL
    sp.call([
        "python", "semantic_mapping.py",
        "--obj-dir", f"export/{rel_path.parts[1]}_translated",
        "--geojson", f"{bo}"
    ])
    delete_files(f"export/{rel_path.parts[1]}_translated")

    # Convert OBJ ke CityGML lod2
    sp.call([
        "go", "run", "obj2lod2gml.go",
        "-input", f"export/{rel_path.parts[1]}_translated",
        "-output", f"export/{rel_path.parts[1]}_translated_gml"
    ])
    delete_files(f"export/{rel_path.parts[1]}_translated_gml")

    # # Convert OBJ ke CityGML lod1
    # sp.call([
    #     "go", "run", "obj2gml.go",
    #     "-input", f"export/{folder_name}.obj_translated",
    #     "-output", f"export/{folder_name}.obj_translated_gml"
    # ])

    # Merge keseluruhan CityGMl lod2 file menjadi 1 file
    sp.call([
        "python", "lod2merge.py",
        f"export/{rel_path.parts[1]}_translated_gml",
        f"{output_path}",
        "--name", f"{folder_name}"
    ])

    # # Merge Keseluruhan CityGML lod1 file menjadi 1 file
    # sp.call([
    #     "go", "run", "mergegml.go",
    #     "-input", f"export/{folder_name}.obj_translated_gml",
    #     "-output", f"percepatan/citygml/{folder_name}.gml"
    # ])

    delete_directories(
        [
            f"export/{rel_path.parts[1]}",
            f"export/{rel_path.parts[1]}_translated",
            f"export/{rel_path.parts[1]}_translated_gml"
        ]
    )

end = time.time() - start
print(f"durasi : {end} s")