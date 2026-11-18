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
The provider is not published to the Terraform Registry under that address.
CI builds the Coreplane source and stores a Linux amd64 binary as a workflow artifact.
CI does not create releases or publish providers.

### Use the Coreplane build before Registry publication

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
      source = "coreplanelabs/parasail"
    }
  }
}
```

Run `terraform plan` or `tofu plan` with the local override.
For a configuration that uses only this provider, skip `init`: the Registry does
not contain the Coreplane provider yet.
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

## Future Registry publication

This fork contains no release workflow. Publication requires a separate change.
The existing GoReleaser configuration and `terraform-registry-manifest.json`
provide the artifact format for that work. `task release:check` validates the
configuration without a release or upload.

Before publication, configure a GPG key, signed release artifacts, and the
`coreplanelabs/parasail` provider in the Terraform Registry.
Follow the [Terraform provider publication guide](https://developer.hashicorp.com/terraform/registry/providers/publishing).
