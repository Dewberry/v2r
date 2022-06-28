# v2r
Vector to Raster Interpolation Routines 

# Features
v2r contains two algorithms with an associated testing suite.

- Inverse Distance Weighting (IDW)
- Cleaner 
## IDW
![idw-sample](images/idw_sample.png)
<br>
The interpolation becomes more drastic as _p_ increases.

### **Equation** <br>
$$z_p= \frac{\displaystyle\sum_{i=1}^{n} (\frac {z_i}{d_i^p}) } {\displaystyle\sum_{i=1}^{n} (\frac {1}{d_i^p})}$$
<br>

**Read from Geopackage** <br>
`./main idw -g -f [FILEPATH] --layer [LAYERNAME] --field [FIELDNAME]`

**Read from txt file** <br>
`./main idw -f [FILEPATH]`

**Flags**<br>
| Shorthand | Full Name     | Type   | Default                          | Description |
| --------- | ------------- | ------ | -------------------------------- | ----------- |
|           | --ascii       | bool   | _false_                          | Perform an additional write to an ascii file |
| -c        | --concurrent  | bool   | _false_                          | Run concurrently? |
|           | --cx          | int    | _200_                            | Set chunk size in x-direction  |
|           | --cy          | int    | _200_                            | Set chunk size in y-direction  |
|           | --ee          | float  | _1.5_                            | End for exponent (inclusive) |
|           | --ei          | float  | _1.5_                            | Exponential increment for calculations between start and end  |
|           | --epsg        | int    | _2284_                           | Set EPSG |
|           | --es          | float  | _0.5_                            | Start for exponent (inclusive)  |
|           | --excel       | bool   | _false_                          | Perform an additional write to excel spreadsheet |
|           | --field       | string | _elevation_                      | Set name of field in geopackage file  |
| -f        | --file        | string | _tests/idw_files/idw_in.txt_     | File to run |
| -g        | --gpkg        | bool   | _false_                          | Read from gpkg (true) or txt file (false)  |
|           | --layer       | string | _layer1_                         | Set name of layer in geopackage file  |
|           | --outPath     | string | _data/idw_                       | Set outfile directory  |
|           | --sx          | float  | _100.0_                          | Set step size in x-direction |
|           | --sy          | float  | _100.0_                          | Set step size in y-direction |

**Notes**
- txt file input requires special formatting (example [idw_in.txt](tests/idw_files/idw_in.txt))
- cx, cy only used if --concurrent=true
- invalid chunk sizes are converted to 1/4 of respective direction (~16 subprocesses)
- epsg only used if --gpkg=false
- field, layer required if --gpkg=true 
- ascii, excel only used if --concurrent=false
- not recommended to run ascii or excel prints for large datasets


## Cleaner
![cleaner_before_after](images/cleaner_before_after.png)

**Usage**<br>
`./main clean -f [FILEPATH]`

**Flags**<br>
| Shorthand | Full Name     | Type   | Default      | Description |
| --------- | ------------- | ------ | --------     | ----------- |
| -a        | --adjacent    | int    | _8_          | Set adjacency type to include ordinal directions  [4 \| 8] |
| -c        | --concurrent  | bool   | _false_       | Run concurrently? |
|           | --cx          | int    | _2560_       | Set chunk size in x-direction  |
|           | --cy          | int    | _2560_       | Set chunk size in y-direction  |
| -f        | --file        | string | _required_   | File to run |
|           | --ti          | float  | _40,000.0_   | Set tolerance level for islands (sq. footage)|
|           | --tv          | float  | _22,500.0_    | Set tolerance level for voids (sq. footage)|

**Notes**
- monitor memory usage (process can use up to 80% of free memory)
- cx, cy only used if --concurrent=true
- cx, cy number of columns, rows to partition file into for subprocess calculations
- invalid chunk sizes are converted to 1/4 of respective direction (~16 subprocesses)

# Testing Suite
**Usage** <br>
`./main test`

**Notes** <br>
- for silent mode, use -e=true, only failed tests will be printed
- outputs ascii files to compare against correct outputs

# Logging
By default, logs are sent to Stdout at the INFO level. 

**Flags**
| Shorthand | Full Name | Type | Default   | Description |
| --------- | --------- | ---- | --------- | ----------- |
| -d        | --debug   | bool | _false_   | Set logging level to DEBUG |
| -e        | --error   | bool | _false_   | Set logging level to ERROR |
| -l        | --log     | bool | _false_   | Log outputs to separate file |

**Notes**
- output logs written to _logs/_
- sample log filename: _2022-24-06_22:06:25.txt_
- if -d or -e are not passed, level=INFO used