import subprocess as sp
import time
import os
import sys
from tqdm import tqdm
from pathlib import Path
from datetime import datetime

from findFile import find_complete_sets, read_and_convert_txt
from cacheHandling import delete_directories, delete_files

class OutputCapture:
    def __init__(self, log_file='processing.log'):
        self.log_file = log_file
        self.original_stdout = sys.stdout
        self.original_stderr = sys.stderr
        
    def __enter__(self):
        # Create directory if it doesn't exist
        os.makedirs(os.path.dirname(self.log_file), exist_ok=True)
        self.log_handle = open(self.log_file, 'w', encoding='utf-8')
        sys.stdout = self.log_handle
        sys.stderr = self.log_handle
        return self
        
    def __exit__(self, exc_type, exc_val, exc_tb):
        sys.stdout = self.original_stdout
        sys.stderr = self.original_stderr
        self.log_handle.close()

def log_with_timestamp(message):
    """Print message with timestamp"""
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{timestamp}] {message}")

def run_subprocess_with_capture(cmd, description=""):
    """Run subprocess and capture ALL its output to the log file"""
    log_with_timestamp(f"Starting: {description}")
    log_with_timestamp(f"Command: {' '.join(cmd)}")
    
    try:
        # Run subprocess and capture both stdout and stderr
        result = sp.run(
            cmd, 
            stdout=sp.PIPE, 
            stderr=sp.STDOUT,  # Redirect stderr to stdout
            text=True,
            check=False  # Don't raise exception on non-zero return code
        )
        
        # Log the output
        if result.stdout:
            log_with_timestamp("Command output:")
            print(result.stdout)
        
        log_with_timestamp(f"Command completed with return code: {result.returncode}")
        
        if result.returncode != 0:
            log_with_timestamp(f"WARNING: Command failed with return code {result.returncode}")
        
        return result.returncode
        
    except Exception as e:
        log_with_timestamp(f"ERROR running command: {str(e)}")
        return -1

def main():
    start = time.time()

    print(f"\n⚙️  Program is running... Please wait 😬🙏")
    
    root_dir = "test"
    
    # Set up log file path
    log_path = f'{root_dir}/detailed_processing.log'.replace('OBJ', 'CityGML')
    
    # Get file set first (before capturing output)
    file_set = find_complete_sets(root_dir)
    
    # Create progress bar OUTSIDE the output capture context
    pbar = tqdm(total=len(file_set), desc="⏳ Processing files", unit="file", 
                position=0, leave=True, file=sys.__stdout__)
    
    # Capture all output to log file
    with OutputCapture(log_path):
        log_with_timestamp("=== PROCESSING STARTED ===")
        log_with_timestamp(f"Root directory: {root_dir}")
        log_with_timestamp(f"Log file: {log_path}")
        log_with_timestamp(f"Found {len(file_set)} file sets to process")
        
        for i, file_data in enumerate(file_set):
            log_with_timestamp(f"--- Processing file set {i+1}/{len(file_set)} ---")
            
            obj = file_data[0]
            coord = read_and_convert_txt(file_data[1])
            bo = file_data[2]

            root_path = Path(root_dir)
            obj_path = Path(obj)

            rel_path = obj_path.relative_to(root_path)
            folder_name = rel_path.parts[0]
            
            log_with_timestamp(f"Processing folder: {folder_name}")
            log_with_timestamp(f"OBJ file: {obj}")
            log_with_timestamp(f"Coordinates: {coord}")
            log_with_timestamp(f"BO file: {bo}")

            output_path = f"{root_dir}/{folder_name}.gml".replace('OBJ', 'CityGML')
            os.makedirs(f"{root_dir}".replace('OBJ', 'CityGML'), exist_ok=True)
            log_with_timestamp(f"Output path: {output_path}")

            # Update progress bar description (this shows in terminal)
            pbar.set_description(f"Processing {folder_name}")

            # Step 1: Pemisahan Bangunan
            log_with_timestamp("STEP 1/5: Building separation")
            run_subprocess_with_capture([
                "go", "run", "objseparator.go", 
                f"-cx={coord[0]}", f"-cy={coord[1]}",
                f"{obj}", 
                f"{bo}",
                f"{root_dir}/{folder_name}/obj"
            ], "Building separation")

            # Step 2: Translasi Objek Menuju Koordinat UTM
            log_with_timestamp("STEP 2/5: Object translation")
            run_subprocess_with_capture([
                "go", "run", "translate.go", 
                f"-input={root_dir}/{folder_name}/obj", 
                f"-output={root_dir}/{folder_name}/translated", 
                f"-tx={coord[0]}", 
                f"-ty={coord[1]}",
                "-tz=0"
            ], "Object translation to UTM coordinates")

            # Step 3: Generate MTL
            log_with_timestamp("STEP 3/5: MTL generation")
            run_subprocess_with_capture([
                "python", "semantic_mapping.py",
                "--obj-dir", f"{root_dir}/{folder_name}/translated",
                "--geojson", f"{bo}"
            ], "MTL generation")

            # Step 4: Convert OBJ ke CityGML lod2
            log_with_timestamp("STEP 4/5: OBJ to CityGML conversion")
            run_subprocess_with_capture([
                "go", "run", "obj2lod2gml.go",
                "-input", f"{root_dir}/{folder_name}/translated",
                "-output", f"{root_dir}/{folder_name}/citygml"
            ], "OBJ to CityGML LOD2 conversion")

            # Step 5: Merge keseluruhan CityGMl lod2 file menjadi 1 file
            log_with_timestamp("STEP 5/5: CityGML file merging")
            run_subprocess_with_capture([
                "python", "lod2merge.py",
                f"{root_dir}/{folder_name}/citygml",
                f"{output_path}",
                "--name", f"{folder_name}"
            ], "CityGML file merging")

            # Final cleanup
            log_with_timestamp("Final cleanup")
            directories_to_delete = [
                f"{root_dir}/{folder_name}/obj",
                f"{root_dir}/{folder_name}/translated",
                f"{root_dir}/{folder_name}/citygml"
            ]
            log_with_timestamp(f"Deleting directories: {directories_to_delete}")
            delete_directories(directories_to_delete)
            
            log_with_timestamp(f"✅ Completed processing {folder_name}")
            
            # Update progress bar (this shows in terminal)
            pbar.update(1)
            pbar.set_description(f"✅ Completed all processing")

        end = time.time() - start
        log_with_timestamp("=== PROCESSING COMPLETED ===")
        log_with_timestamp(f"Total duration: {end:.2f} seconds")
        log_with_timestamp(f"Processed {len(file_set)} file sets")
        log_with_timestamp(f"Average time per file set: {end/len(file_set):.2f} seconds")

    # Close progress bar
    pbar.close()
    
    # This prints to terminal after log capture is done
    print(f"\n🎉 All processing completed!")
    print(f"📊 Processed {len(file_set)} file sets in {end:.2f} seconds")
    print(f"📝 Detailed logs with timestamps saved to '{log_path}'")
    print("\n© 2025. Fairuz Akmal Pradana 👱")

if __name__ == "__main__":
    main()
