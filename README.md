# tfau - Terraform Auto Upgrade

`tfau` is a command-line tool designed to simplify the process of upgrading Terraform modules, providers, and Terraform versions within your HCL (HashiCorp Configuration Language) files. It automates the retrieval of the latest versions and updates your files in place, streamlining the maintenance of your Terraform infrastructure.

## Features

-   **Module Upgrades**: Automatically fetches and updates module versions from the Terraform Registry or Git repositories.
-   **Provider Upgrades**: Retrieves and updates provider versions from the Terraform Registry.
-   **Terraform Version Upgrades**: Fetches the latest Terraform version and updates the `required_version` in your files.
-   **Selective Upgrades**: Allows you to specify which components (modules, providers, Terraform) to upgrade.
-   **Recursive File Discovery**: Automatically discovers `.tf` files in the current working directory if no specific files are provided.
-   **Command-Line Interface**: Easy-to-use CLI with flags for customization.
-   **Handles Git SSH URLs**: Supports Git SSH URLs (e.g., `git@github.com:user/repo.git`).

## Installation

To install `tfau`, ensure you have Go installed and configured. Then, run the following command:

```bash
make install
```

This will install the `tfau` executable in your `$GOPATH/bin` directory. Make sure this directory is in your `$PATH`.

### Makefile Commands

`make all` or `make build`: Builds the `tfau` executable.

`make run`: Runs the `tfau` executable.

`make test`: Runs the Go tests.

`make install`: Builds and installs the `tfau` executable in your `$GOPATH/bin`.

`make clean`: Removes the `tfau` executable.

`make fclean`: Cleans the build.

## Usage

```bash
tfau [flags]
```

### Flags
- `-f`, `--file` stringArray: HCL file(s) to be updated. You can specify multiple files.
- `--upgrades string`: Comma-separated list of upgrades (modules, providers, terraform). If not specified, all upgrades are performed.
- `-v`, `--verbose`: Enable verbose output.
- `--terraform-version string`: Desired Terraform version to update to (e.g., `~>1.9`). If not specified, the latest version is used.

### Examples

1. Upgrade all modules, providers, and Terraform versions in all .tf files in the current directory:
```bash
tfau
```

2. Upgrade only modules and providers in a specific file:
```bash
tfau -f main.tf --upgrades modules,providers
```

3. Upgrade Terraform version to a specific version in multiple files:
```bash
tfau -f main.tf -f modules/network.tf --terraform-version "~>1.9"
```

4. Upgrade only modules in a specific file with verbose output:
```bash
tfau -f main.tf --upgrades modules -v
```

## How It Works

### File Discovery

`tfau` either uses the files specified with the `-f` flag or recursively finds all `.tf` files in the current directory.

### Parsing

It parses the HCL files using the `hashicorp/hcl/v2` library to extract module, provider, and Terraform version information.

### Version Retrieval

- For modules, it fetches the latest version from the Terraform Registry or Git repositories.

- For providers, it fetches the latest version from the Terraform Registry.

- For Terraform, it fetches the latest version from the HashiCorp releases API.

### Updates

`tfau` updates the HCL files in place with the latest versions using the `hashicorp/hcl/v2/hclwrite` library.

### Selective Upgrades

The `--upgrades` flag allows you to specify which components to upgrade, providing flexibility and control.


## Dependencies

`github.com/spf13/cobra`: For the command-line interface.

`github.com/hashicorp/hcl/v2`: For parsing and writing HCL files.

`github.com/hashicorp/go-version`: For parsing and comparing semantic versions.

`github.com/go-git/go-git/v5`: For fetching Git tags.

