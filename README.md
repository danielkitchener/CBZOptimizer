# CBZOptimizer

CBZOptimizer is a Go-based tool designed to optimize CBZ (Comic Book Zip) files by converting images to a specified format and quality. This tool is useful for reducing the size of comic book archives while maintaining acceptable image quality.

## Features

- Convert images within CBZ files to different formats (e.g., WebP).
- Adjust the quality of the converted images.
- Process multiple chapters in parallel.
- Option to override the original CBZ files.

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

The tool provides a CLI command to optimize CBZ files. Below is an example of how to use it:

```sh
go run main.go optimize --quality 85 --parallelism 2 --override /path/to/cbz/files
```

### Flags

- `--quality`, `-q`: Quality for conversion (0-100). Default is 85.
- `--parallelism`, `-n`: Number of chapters to convert in parallel. Default is 2.
- `--override`, `-o`: Override the original CBZ files. Default is false.

## Testing

To run the tests, use the following command:

```sh
go test ./... -v
```

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
