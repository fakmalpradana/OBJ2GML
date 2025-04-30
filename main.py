import subprocess as sp
import time

start = time.time()

nlp = "AG_09"
sub_grid = "A"
bo = "percepatan/OBJ/AG_09/AG_09_A/AG-09-A_BO_Caesar Yoga_BUFFER_Lengkap.geojson"

sp.call([
    "go", "run", "objseparator.go", 
    f"percepatan/OBJ/{nlp}/{nlp}_{sub_grid}/{nlp}_{sub_grid}.obj", 
    f"percepatan/OBJ/{nlp}/{nlp}_{sub_grid}/AG-09-A_BO_Caesar Yoga_BUFFER_Lengkap.geojson"
])

sp.call([
    "go", "run", "translate.go", 
    f"-input=export/{nlp}_{sub_grid}.obj", 
    "-tx=692827.46065", 
    "-ty=9326588.60235",
    "-tz=100"
])

sp.call([
    "go", "run", "obj2gml.go",
    "-input", f"export/{nlp}_{sub_grid}.obj_translated",
    "-output", f"export/{nlp}_{sub_grid}.obj_translated_gml"
])

sp.call([
    "go", "run", "mergegml.go",
    "-input", f"export/{nlp}_{sub_grid}.obj_translated_gml",
    "-output", f"percepatan/citygml/{nlp}_{sub_grid}.gml"
])

end = time.time() - start
print(f"durasi : {end} s")