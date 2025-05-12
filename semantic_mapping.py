import os
import json
import numpy as np
from pathlib import Path
import shapely.geometry as geometry
from shapely.geometry import Polygon

# Define colors for semantic types
COLORS = {
    'Roof': (1.0, 0.65, 0.0),    # Orange
    'Wall': (0.5, 0.5, 0.5),     # Grey
    'Ground': (0.82, 0.41, 0.12) # Chocolate
}

class BuildingColorizer:
    def __init__(self, obj_dir, geojson_path):
        self.obj_dir = Path(obj_dir)
        self.geojson_path = Path(geojson_path)
        self.building_outlines = self.load_all_building_outlines()
        
    def load_all_building_outlines(self):
        """Load all building outlines from a single GeoJSON file"""
        building_outlines = {}
        try:
            with open(self.geojson_path, 'r') as f:
                data = json.load(f)
                for feature in data['features']:
                    # Get the coordinates of the building outline
                    if feature['geometry']['type'] == 'MultiPolygon':
                        # Take the first polygon from MultiPolygon
                        coords = feature['geometry']['coordinates'][0][0]
                        polygon = Polygon(coords)
                        # Store polygon with its centroid as key
                        centroid = polygon.centroid
                        building_outlines[f"{centroid.x}_{centroid.y}"] = polygon
                print(f"Loaded {len(building_outlines)} building outlines from GeoJSON")
                return building_outlines
        except Exception as e:
            print(f"Error loading GeoJSON file: {str(e)}")
            return {}

    def load_obj_file(self, obj_path):
        vertices = []
        faces = []
        
        with open(obj_path, 'r') as f:
            for line in f:
                if line.startswith('v '):
                    # Parse vertex
                    _, x, y, z = line.split()
                    vertices.append([float(x), float(y), float(z)])
                elif line.startswith('f '):
                    # Parse face
                    _, *face_vertices = line.split()
                    # Extract vertex indices (handle different OBJ formats)
                    face = [int(v.split('/')[0]) - 1 for v in face_vertices]
                    faces.append(face)
                    
        return np.array(vertices), faces

    def get_face_type(self, vertices, face, building_outline):
        # Calculate face normal
        v0 = vertices[face[0]]
        v1 = vertices[face[1]]
        v2 = vertices[face[2]]
        
        normal = np.cross(v1 - v0, v2 - v0)
        if np.all(normal == 0):
            return 'Wall'  # Default to wall if normal calculation fails
        normal = normal / np.linalg.norm(normal)
        
        # Project face vertices to 2D for ground check
        face_points_2d = [(vertices[idx][0], vertices[idx][1]) for idx in face]
        
        try:
            face_polygon_2d = Polygon(face_points_2d)
            if not face_polygon_2d.is_valid:
                return 'Wall'  # Default to wall if polygon is invalid
            
            # Check if face is ground (parallel to XY plane and overlapping with outline)
            if abs(normal[2]) > 0.95:  # Almost parallel to ground
                # Check Z coordinate - ground should be at or near Z=0
                avg_z = sum(vertices[idx][2] for idx in face) / len(face)
                if avg_z < 0.1:  # Assuming ground level is near 0
                    if building_outline:
                        intersection = building_outline.intersection(face_polygon_2d)
                        if not intersection.is_empty:
                            overlap_ratio = intersection.area / face_polygon_2d.area
                            if overlap_ratio > 0.99:
                                return 'Ground'
            
            # Check if face is wall (perpendicular to ground)
            if abs(normal[2]) < 0.1:  # Almost perpendicular to ground
                return 'Wall'
            
            # Otherwise, it's a roof
            return 'Roof'
            
        except Exception as e:
            print(f"Error determining face type: {str(e)}")
            return 'Wall'  # Default to wall in case of errors

    def create_mtl_file(self, obj_path, face_types):
        mtl_path = obj_path.with_suffix('.mtl')
        unique_types = set(face_types)
        
        with open(mtl_path, 'w') as f:
            for type_name in unique_types:
                f.write(f'newmtl {type_name}\n')
                color = COLORS[type_name]
                f.write(f'Kd {color[0]} {color[1]} {color[2]}\n\n')

    def update_obj_file(self, obj_path, face_types):
        temp_path = obj_path.with_suffix('.tmp')
        
        with open(obj_path, 'r') as src, open(temp_path, 'w') as dst:
            # Write MTL reference
            dst.write(f'mtllib {obj_path.stem}.mtl\n')
            
            face_idx = 0
            for line in src:
                if line.startswith('f '):
                    # Add material before face
                    dst.write(f'usemtl {face_types[face_idx]}\n')
                    dst.write(line)
                    face_idx += 1
                elif not line.startswith('mtllib'):
                    dst.write(line)
        
        # Replace original file
        os.replace(temp_path, obj_path)

    def process_building(self, obj_path):
        # Load OBJ vertices first
        vertices, faces = self.load_obj_file(obj_path)
        
        # Calculate centroid of the OBJ building
        x_coords = [v[0] for v in vertices]
        y_coords = [v[1] for v in vertices]
        obj_centroid_x = sum(x_coords) / len(x_coords)
        obj_centroid_y = sum(y_coords) / len(y_coords)
        
        # Find the closest building outline
        min_distance = float('inf')
        closest_outline = None
        
        for centroid_key, outline in self.building_outlines.items():
            geojson_x, geojson_y = map(float, centroid_key.split('_'))
            distance = ((obj_centroid_x - geojson_x)**2 + (obj_centroid_y - geojson_y)**2)**0.5
            if distance < min_distance:
                min_distance = distance
                closest_outline = outline
        
        # Use a threshold to determine if we found a matching outline
        if min_distance < 1.0:  # 1 meter threshold, adjust as needed
            building_outline = closest_outline
        else:
            building_outline = None
            print(f"No building outline found for {obj_path.stem}")

        if building_outline is None:
            # If no outline found, treat all horizontal faces as roof and vertical faces as walls
            face_types = []
            for face in faces:
                v0 = vertices[face[0]]
                v1 = vertices[face[1]]
                v2 = vertices[face[2]]
                normal = np.cross(v1 - v0, v2 - v0)
                if np.all(normal == 0):
                    face_types.append('Wall')
                else:
                    normal = normal / np.linalg.norm(normal)
                    face_types.append('Wall' if abs(normal[2]) < 0.1 else 'Roof')
        else:
            # Determine face types using building outline
            face_types = []
            for face in faces:
                face_type = self.get_face_type(vertices, face, building_outline)
                face_types.append(face_type)
        
        # Create MTL file
        self.create_mtl_file(obj_path, face_types)
        
        # Update OBJ file with material references
        self.update_obj_file(obj_path, face_types)

    def process_all_buildings(self):
        for obj_path in self.obj_dir.glob('*.obj'):
            print(f'Processing building: {obj_path.name}')
            self.process_building(obj_path)

def main():
    # Set up directories
    obj_dir = 'export/AG_09_semantic_test'
    geojson_dir = 'percepatan/OBJ/AG_09/AG_09_D/AG-09-D_BO_Caesar Yoga_BUFFER_Lengkap.geojson'
    
    # Create and run colorizer
    colorizer = BuildingColorizer(obj_dir, geojson_dir)
    colorizer.process_all_buildings()

if __name__ == '__main__':
    main()
