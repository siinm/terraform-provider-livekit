---
page_title: "Livekit Provider"
subcategory: ""
description: |-
    The Livekit provider provides resources to manage access tokens for Livekit.
---

# Livekit Provider

The Livekit provider allows you to manage access tokens for [Livekit](https://livekit.io/).

The changelog for this provider can be found here: <https://github.com/siinm/terraform-provider-livekit/releases>.

This provider is independently developed and is not affiliated with the LiveKit company.

## Example Usage

### Creating a Livekit provider

```terraform
provider "livekit" {
  api_key    = var.livekit_api_key
  api_secret = var.livekit_api_secret
}

// Create an access token
resource "livekit_access_token" "example_token" {
  room     = "example_room"
  identity = "example_identity"
  valid_for = "1h"
}
```

## Schema

### Optional

- `api_key` (String) Livekit API Key. Can also be set via the `LIVEKIT_API_KEY` environment variable.
- `api_secret` (String) Livekit API Secret. Can also be set via the `LIVEKIT_API_SECRET` environment variable.


## Functions

Currently, the Livekit provider does not support any functions.

## Data Sources

Currently, the Livekit provider does not support any data sources.

