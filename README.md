# Terraform Provider for Parasail (Unofficial)

> **Unofficial / community project.** This provider is not built, maintained, or
> endorsed by Parasail. It is an independent project that talks to Parasail's
> **public** Control API only, and depends on no Parasail internal code. "Parasail"
> is used solely to describe the API the provider targets.

This repository contains a Terraform (and OpenTofu) provider for Parasail's public
Control API. The initial implementation focuses on dedicated model deployments using
the documented `/api/v1/dedicated` API surface.

## Coreplane fork

Coreplane maintains this fork at
[coreplanelabs/terraform-provider-parasail](https://github.com/coreplanelabs/terraform-provider-parasail).
[loewenthal-corp/terraform-provider-parasail](https://github.com/loewenthal-corp/terraform-provider-parasail)
is the upstream source. The fork keeps the upstream history and MPL-2.0 license.

Coreplane builds use the provider address `coreplanelabs/parasail`.
Install the provider from the [Terraform Registry](https://registry.terraform.io/providers/coreplanelabs/parasail/latest).
CI builds the Coreplane source and stores a Linux amd64 binary as a workflow artifact.
The manual release workflow publishes signed artifacts from `main`.

### Install from the Registry

```hcl
terraform {
  required_providers {
    parasail = {
      source  = "registry.terraform.io/coreplanelabs/parasail"
      version = "~> 0.2.1"
    }
  }
}
```

Run `terraform init` or `tofu init` to install the signed provider.
Set `PARASAIL_API_KEY` before you plan or apply deployments.

### Use a local Coreplane build

```sh
git clone https://github.com/coreplanelabs/terraform-provider-parasail.git
cd terraform-provider-parasail
source bin/activate-hermit
task dev:setup
export TF_CLI_CONFIG_FILE="$PWD/.task/opentofu/provider-dev.tfrc"
```

`task dev:setup` builds a native binary in `dist/` and writes a CLI configuration
that maps the Coreplane provider address to that directory.
Keep `TF_CLI_CONFIG_FILE` set in the shell where you run Terraform or OpenTofu.
Use this provider requirement in your configuration:

```hcl
terraform {
  required_providers {
    parasail = {
      source = "registry.terraform.io/coreplanelabs/parasail"
    }
  }
}
```

Run `terraform plan` or `tofu plan` with the local override.
Run `task dev:setup` after source changes to rebuild the binary.
The CLI configuration applies to any working directory in the same shell.

### Git remotes

Configure each clone so that normal pushes target Coreplane:

```sh
git remote add upstream https://github.com/loewenthal-corp/terraform-provider-parasail.git
git config remote.pushDefault origin
git config push.default current
```

Fetch changes with `git fetch upstream`.
Merge `upstream/main` on a branch, keep the Coreplane namespace and workflow
settings, and submit a PR to this fork.
CI checks upstream updates through the same tasks as other changes.

## Development

```sh
source bin/activate-hermit
task init
task do
```

Provider configuration can come from Terraform configuration or environment variables:

- `PARASAIL_API_KEY`
- `PARASAIL_ENDPOINT`, defaults to `https://api.parasail.io/api/v1`

### Acceptance Tests

Ordinary CI runs unit tests without Parasail credentials.
Live acceptance tests create real deployments and can incur charges.
The manual `acceptance` workflow runs on `main` and defaults to read-only tests.
Set `PARASAIL_API_KEY` in the `acceptance` environment before you run it.
Choose `mutating` only when you intend to create and destroy deployments.

For local live tests, put `PARASAIL_API_KEY` in `.env` or export it:

```sh
task test:acc:readonly # read-only support, device, and list checks
task test:acc          # create, update, import, list, and destroy checks
```

The mutating suite sets `wait_for_online = false` to check the deployment
lifecycle without a wait for instances to serve traffic.

## Example

See [examples/basic/main.tf](examples/basic/main.tf).

The resource is intentionally shaped around deployment intent rather than the raw
API payload:

```hcl
resource "parasail_dedicated_deployment" "llama" {
  name  = "llama-prod"
  model = "meta-llama/Llama-3.1-8B-Instruct"

  # Repeat gpu blocks to allow more than one acceptable hardware profile.
  # The API chooses from these selected profiles when provisioning.
  gpu {
    type  = "H100SXM"
    count = 2
  }

  gpu {
    type  = "H200SXM"
    count = 4
  }

  autoscaling {
    min_replicas = 1
    max_replicas = 5
  }

  scale_down_after = "never"
}

output "model_alias" {
  value = parasail_dedicated_deployment.llama.model_alias
}
```

## Documentation

Registry docs under [docs/](docs/) are generated from the provider schema and the
files in [examples/](examples/) with
[`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs):

```sh
task docs           # regenerate docs/ and validate them
task docs:validate  # validate only
```

`task do` and CI regenerate and validate docs, so edit the schema/examples rather
than the generated Markdown.

## Releases

The manual [release workflow](.github/workflows/release.yaml) builds from `main`.
Its tag must match `.release-please-manifest.json` and `internal/buildinfo/buildinfo.go`.
The first Coreplane release uses `v0.2.1`.

The workflow runs the full quality checks, builds 13 platform archives, and signs
the checksums. It verifies the uploaded archives, Registry manifest, checksums,
and GPG signature before it publishes the draft GitHub release.
It refuses to replace a published version.

The `release` environment contains the `GPG_PRIVATE_KEY` and `PASSPHRASE` secrets.
Restrict this environment to `main`.
The public key lives in [keys/parasail.asc](keys/parasail.asc).
Register that key in the `coreplanelabs` Registry namespace before publication.

Run a release after the version change merges and CI passes:

```sh
gh workflow run release.yaml --repo coreplanelabs/terraform-provider-parasail --ref main -f tag=v0.2.1
```

The Registry publishes this repository as `coreplanelabs/parasail`.
It imports each signed GitHub release through the Terraform Cloud GitHub App.
See the [provider publication guide](https://developer.hashicorp.com/terraform/registry/providers/publishing).

`task release:check` builds unsigned snapshot artifacts and verifies the Registry
file contract without an upload.
