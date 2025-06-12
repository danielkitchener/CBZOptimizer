# CBZOptimizer

CBZOptimizer is a Go-based tool designed to optimize CBZ (Comic Book Zip) and CBR (Comic Book RAR) files by converting images to a specified format and quality. This tool is useful for reducing the size of comic book archives while maintaining acceptable image quality.

**Note**: CBR files are supported as input but are always converted to CBZ format for output.

## Features

- Convert images within CBZ and CBR files to different formats (e.g., WebP).
- Support for multiple archive formats including CBZ and CBR (CBR files are converted to CBZ format).
- Adjust the quality of the converted images.
- Process multiple chapters in parallel.
- Option to override the original files (CBR files are converted to CBZ and original CBR is deleted).
- Watch a folder for new CBZ/CBR files and optimize them automatically.

## Installation

1. Clone the repository:

```sh
git clone https://github.com/belphemur/CBZOptimizer.git
cd CBZOptimizer
```

2. Install dependencies:
   ```sh
   go mod tidy
   ```

## Usage

### Command Line Interface

The tool provides CLI commands to optimize and watch CBZ/CBR files. Below are examples of how to use them:

#### Optimize Command

Optimize all CBZ/CBR files in a folder recursively:

```sh
go run main.go optimize [folder] --quality 85 --parallelism 2 --override --format webp --split
```

#### Watch Command

Watch a folder for new CBZ/CBR files and optimize them automatically:

```sh
go run main.go watch [folder] --quality 85 --override --format webp --split
```

### Flags

- `--quality`, `-q`: Quality for conversion (0-100). Default is 85.
- `--parallelism`, `-n`: Number of chapters to convert in parallel. Default is 2.
- `--override`, `-o`: Override the original files. For CBZ files, overwrites the original. For CBR files, deletes the original CBR and creates a new CBZ. Default is false.
- `--split`, `-s`: Split long pages into smaller chunks. Default is false.
- `--format`, `-f`: Format to convert the images to (e.g., webp). Default is webp.

## Testing

To run the tests, use the following command:

```sh
go test ./... -v
```

## Requirement

Needs to have libwep installed on the machine if you're not using the docker image

## Docker

`ghcr.io/belphemur/cbzoptimizer:latest`

## GitHub Actions

The project includes a GitHub Actions workflow to run tests on every push and pull request to the `main` branch. The workflow is defined in `.github/workflows/go.yml`.

## Contributing

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature-branch`).
5. Create a new Pull Request.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
