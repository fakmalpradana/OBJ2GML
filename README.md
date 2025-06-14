# OBJ to CityGML Converter
OBJ Data extractor untuk keperluan 3D City Building, dapat mengakomodir tingkat kedetailan LOD1 - LOD3 Dengan catatan memiliki data OBJ (hasil sketchup) dan GeoJSON untuk Building Outline

## Installation & Requirements

Silahkan install GO dan miniconda terlebih dahulu. Selanjutnya silahkan clone repo ini, lalu lanjutkan dengan setup environment pada miniconda

**Clone repository**

HTTPS
```bash
git clone https://github.com/fakmalpradana/OBJ2GML.git
```
SSH
```bash
git clone git@github.com:fakmalpradana/OBJ2GML.git
```

**Setup environment dengan Python 3.10**
```bash
conda create --name py310 python=3.10
```

**Install dependensi**
```bash
pip install -r requirements.txt
```
    
## Petnjuk Penggunaan

Cukup jalankan file `main.py` dengan command berikut

```bash
python main.py
```

Program akan jalan dalam CLI dan memiliki loading bar untuk memantau proses berlangsung, berikut adalah contohnya
```bash
(py310) (base) mal@Mac OBJ2GML % python main.py

⚙️  Program is running... Please wait 😬🙏
✅ Completed all processing: 100%|████████████████████████████████| 3/3 [00:06<00:00,  2.27s/file]

🎉 All processing completed!
📊 Processed 3 file sets in 6.84 seconds
📝 Detailed logs with timestamps saved to 'percepatan_new/CityGML/2025_06_13/detailed_processing.log'
```

Untuk konfigurasi input silahkan edit di dalam file `main.py` cari line 72 dan ganti variabel `root_dir` dengan folder yang berisikan hasil export
```python
root_dir = "percepatan_new/OBJ/2025_06_13"
```
Didalam folder `2025_06_13` seharusnya terdapat sample file seperti berikut
```bash
percepatan_new/OBJ/2025_06_13
├── AG_09_A
│   ├── AG_09_A_sudah rev_fix.mtl
│   ├── AG_09_A_sudah rev_fix.obj
│   ├── AG_09_A_sudah rev_fix.obj.csv
│   ├── BO_AG_09_A.geojson
│   ├── GCP_AG_09_A.cpg
│   ├── GCP_AG_09_A.dbf
│   ├── GCP_AG_09_A.prj
│   ├── GCP_AG_09_A.shp
│   ├── GCP_AG_09_A.shx
│   └── Koordinat_AG_09_A.txt
├── AG_09_B
│   ├── AG_09_A_sudah rev_fix.mtl
│   ├── AG_09_A_sudah rev_fix.obj.csv
│   ├── AG_09_B_sudah rev_fix.obj
│   ├── AG_09_B_sudah rev_fix.obj.csv
│   ├── BO_AG_09_B.geojson
│   ├── GCP_AG_09_A.cpg
│   ├── GCP_AG_09_A.dbf
│   ├── GCP_AG_09_A.prj
│   ├── GCP_AG_09_A.shp
│   ├── GCP_AG_09_A.shx
│   └── Koordinat_AG_09_B.txt
└── AU_10_D
    ├── AG_09_A_sudah rev_fix.mtl
    ├── AG_09_A_sudah rev_fix.obj.csv
    ├── AG_10_D_sudah rev_fix.obj
    ├── AG_10_D_sudah rev_fix.obj.csv
    ├── BO_AG_10_D.geojson
    ├── GCP_AG_09_A.cpg
    ├── GCP_AG_09_A.dbf
    ├── GCP_AG_09_A.prj
    ├── GCP_AG_09_A.shp
    ├── GCP_AG_09_A.shx
    └── Koordinat_AG_10_D.txt
```