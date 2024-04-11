## USPTO-Bulk-Data-Tool
A configurable tool for concurrent processing of U.S. Patent and Trademark Office (USPTO) [bulk data zip files](https://bulkdata.uspto.gov/).

At this time, the tool supports the following USPTO bulk data products:
- **Patent Grant Full Text Data (No Images) (2004 - Present)**
- **Patent Application Full Text Data (No Images) (2004 - Present)**

Given a directory of USPTO zip files, the application will produce one of the following outputs:
- Complete XML files of individual documents split out from the zip
- JSON files of individual documents
    - Selective (non-exhaustive) parsing of main document fields
    - Structured patent claims representing referential relationships, as in the original [PatentPublicData](https://github.com/USPTO/patentpublicdata) tool
    - HTML formatting of Abstract and Description fields
- Apache Parquet files corresponding to bulk zip files


## Usage

Clone this repository.  Edit the `config.toml` as needed - the most important config values are the first three:

```toml
[required]
inputdirectory = "data/in"
outputdirectory = "data/out"
outputmode = "json"
```

For the most basic setup, create `data/in` directories within the project root, and populate the `/in` directory with zip files to process.

Then, from the root of project directory:
```zsh
make run
```

For more advanced usage running the application from somewhere other than the root of the project directory, the executable accepts a single optional argument specifying the path to a `config.toml` file.



